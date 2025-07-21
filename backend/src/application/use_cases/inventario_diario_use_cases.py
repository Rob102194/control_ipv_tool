from datetime import date
import uuid
import json
from src.core.domain.inventario_diario import InventarioDiario
from src.infrastructure.repositories.sqlite_inventario_diario_repository import SQLiteInventarioDiarioRepository
from src.infrastructure.repositories.sqlite_venta_repository import SQLiteVentaRepository
from src.infrastructure.repositories.sqlite_receta_repository import SQLiteRecetaRepository
from src.infrastructure.repositories.sqlite_producto_repository import SQLiteProductoRepository
from src.infrastructure.repositories.sqlite_area_repository import SQLiteAreaRepository

# Caso de uso para obtener el estado del inventario para una fecha dada.
class ObtenerEstadoInventarioDiarioUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository, producto_repository: SQLiteProductoRepository, area_repository: SQLiteAreaRepository):
        self.inventario_repository = inventario_repository
        self.producto_repository = producto_repository
        self.area_repository = area_repository

    def execute(self, fecha: date) -> dict[str, list[InventarioDiario]]:
        registros_existentes = self.inventario_repository.find_by_date(fecha)
        areas = self.area_repository.find_all()

        if not registros_existentes:
            return self._crear_plantilla_vacia(fecha, areas)

        registros_por_area = {area.nombre: [] for area in areas}
        areas_by_id = {area.id: area for area in areas}

        for registro_model in registros_existentes:
            area = areas_by_id.get(registro_model.area_id)
            if area:
                registro_dominio = InventarioDiario(
                    id=registro_model.id,
                    fecha=registro_model.fecha,
                    area_id=registro_model.area_id,
                    producto_id=registro_model.producto_id,
                    inicio=registro_model.inicio,
                    entradas=registro_model.entradas,
                    consumo=registro_model.consumo,
                    merma=registro_model.merma,
                    otras_salidas=registro_model.otras_salidas,
                    final_fisico=registro_model.final_fisico,
                    final_teorico=registro_model.final_teorico,
                    diferencia=registro_model.diferencia,
                    producto_nombre=registro_model.producto.nombre if registro_model.producto else "Producto no encontrado",
                    area_nombre=area.nombre,
                    comentario=registro_model.comentario
                )
                if area.nombre not in registros_por_area:
                    registros_por_area[area.nombre] = []
                registros_por_area[area.nombre].append(registro_dominio)
        
        return registros_por_area

    def _crear_plantilla_vacia(self, fecha: date, areas: list) -> dict[str, list[InventarioDiario]]:
        modelos = self.inventario_repository.get_modelos()
        productos_by_id = {p.id: p for p in self.producto_repository.obtener_todos()}
        plantilla = {area.nombre: [] for area in areas}

        for area in areas:
            productos_modelo = modelos.get(str(area.id), [])
            for producto_data in productos_modelo:
                producto = productos_by_id.get(producto_data['producto_id'])
                if not producto: continue

                inicio = self.inventario_repository.get_inicio_from_previous_day(fecha, area.id, producto.id)
                nuevo_registro = InventarioDiario(
                    id=str(uuid.uuid4()),
                    fecha=fecha,
                    area_id=area.id,
                    producto_id=producto.id,
                    inicio=inicio,
                    producto_nombre=producto.nombre,
                    area_nombre=area.nombre,
                    comentario=""
                )
                if area.nombre not in plantilla:
                    plantilla[area.nombre] = []
                plantilla[area.nombre].append(nuevo_registro)
        
        return plantilla

# Caso de uso para calcular el consumo de productos basado en las ventas.
class CalcularConsumoUseCase:
    def __init__(self, venta_repository: SQLiteVentaRepository, receta_repository: SQLiteRecetaRepository):
        self.venta_repository = venta_repository
        self.receta_repository = receta_repository

    def execute(self, fecha: date) -> dict[str, float]:
        ventas = self.venta_repository.find_by_date(fecha)
        consumo_total = {}

        for venta in ventas:
            receta = self.receta_repository.find_by_name(venta.receta_nombre)
            if receta:
                for ingrediente in receta.ingredientes:
                    # Clave única para el consumo por producto y área.
                    clave = f"{ingrediente.producto_id}|{ingrediente.area_id}"
                    consumo_total[clave] = consumo_total.get(clave, 0) + (venta.cantidad * ingrediente.cantidad)
        
        return consumo_total

# Caso de uso para guardar el estado completo del inventario diario.
class GuardarInventarioDiarioUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository):
        self.inventario_repository = inventario_repository

    def execute(self, data: list[dict]) -> list[InventarioDiario]:
        inventarios_a_guardar = []
        for item in data:
            registro = InventarioDiario(
                id=item.get('id', str(uuid.uuid4())),
                fecha=date.fromisoformat(item['fecha']),
                area_id=item['area_id'],
                producto_id=item['producto_id'],
                inicio=float(item['inicio']),
                entradas=float(item['entradas']),
                consumo=float(item['consumo']),
                merma=float(item['merma']),
                otras_salidas=float(item['otras_salidas']),
                final_fisico=float(item['final_fisico']),
                final_teorico=0, # Se recalculará
                diferencia=0,    # Se recalculará
                producto_nombre=item.get('producto_nombre'),
                area_nombre=item.get('area_nombre'),
                comentario=item.get('comentario')
            )
            # Recalcula los valores finales antes de guardar.
            registro.calcular_diferencias()
            inventarios_a_guardar.append(registro)
        
        self.inventario_repository.save_all(inventarios_a_guardar)
        return inventarios_a_guardar

# Caso de uso para obtener los modelos de IPV.
class ObtenerModelosIPVUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository):
        self.inventario_repository = inventario_repository

    def execute(self) -> dict:
        return self.inventario_repository.get_modelos()

# Caso de uso para guardar un modelo de IPV.
class GuardarModeloIPVUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository):
        self.inventario_repository = inventario_repository

    def execute(self, area_id: str, productos: list[dict]) -> dict:
        self.inventario_repository.save_modelo(area_id, productos)
        return {"area_id": area_id, "productos": productos}

# Caso de uso para obtener las fechas de todos los registros de IPV.
class ObtenerRegistrosIPVUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository):
        self.inventario_repository = inventario_repository

    def execute(self) -> list[dict]:
        fechas = self.inventario_repository.find_all_dates()
        return [{"fecha": fecha.isoformat()} for fecha in fechas]

# Caso de uso para generar un reporte de IPV para una fecha específica.
class GenerarReporteIPVUseCase:
    def __init__(self, inventario_repository: SQLiteInventarioDiarioRepository, area_repository: SQLiteAreaRepository, producto_repository: SQLiteProductoRepository):
        self.inventario_repository = inventario_repository
        self.area_repository = area_repository
        self.producto_repository = producto_repository

    def execute(self, fecha: date) -> dict:
        registros = self.inventario_repository.find_by_date(fecha)
        if not registros:
            raise ValueError("No se encontraron registros para la fecha especificada.")

        productos = {p.id: p for p in self.producto_repository.obtener_todos()}
        areas = {a.id: a for a in self.area_repository.find_all()}
        
        reporte = {
            "fecha": fecha.isoformat(),
            "areas": {},
            "resumen": {},
            "notas": []
        }

        # Inicializar resumen por área
        for area in areas.values():
            reporte["resumen"][area.nombre] = {
                "faltantes": [],
                "sobrantes": [],
                "mermas": []
            }

        # Organizar registros por área
        for reg in registros:
            area = areas.get(reg.area_id)
            producto = productos.get(reg.producto_id)
            if not area or not producto:
                continue

            area_nombre = area.nombre
            if area_nombre not in reporte["areas"]:
                reporte["areas"][area_nombre] = []

            reg_dict = {
                "producto": producto.nombre,
                "um": producto.unidad_medida,
                "inicio": reg.inicio,
                "entradas": reg.entradas,
                "consumo": reg.consumo,
                "merma": reg.merma,
                "otras_salidas": reg.otras_salidas,
                "final_teorico": reg.final_teorico,
                "final_fisico": reg.final_fisico,
                "diferencia": reg.diferencia
            }
            reporte["areas"][area_nombre].append(reg_dict)

            # Resumen de diferencias y mermas por área
            if reg.diferencia < 0:  # Faltante
                reporte["resumen"][area_nombre]["faltantes"].append(f"{producto.nombre}: {abs(reg.diferencia)} {producto.unidad_medida}")
            elif reg.diferencia > 0:  # Sobrante
                reporte["resumen"][area_nombre]["sobrantes"].append(f"{producto.nombre}: {reg.diferencia} {producto.unidad_medida}")

            if reg.merma > 0:
                reporte["resumen"][area_nombre]["mermas"].append(f"{producto.nombre}: {reg.merma} {producto.unidad_medida}")

            # Notas de comentarios
            if reg.comentario and reg.comentario.strip():
                try:
                    comentarios = json.loads(reg.comentario)
                    for campo, texto in comentarios.items():
                        if texto and texto.strip():
                            nota = f"{producto.nombre} ({campo}): {texto}"
                            reporte["notas"].append(nota)
                except json.JSONDecodeError:
                    # Si no es un JSON válido, tratarlo como texto plano
                    nota = f"{producto.nombre}: {reg.comentario}"
                    reporte["notas"].append(nota)

        return reporte

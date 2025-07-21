from abc import ABC, abstractmethod
from src.core.domain import Receta, Ingrediente
from src.application.use_cases.historial_use_cases import RegistrarCambioUseCase
import pandas as pd
import io

# Interfaz abstracta para el repositorio de recetas, definiendo los métodos obligatorios
class IRecetaRepository(ABC):
    @abstractmethod
    def crear(self, receta: Receta) -> Receta:
        """Crea una nueva receta en el repositorio."""
        pass
    
    @abstractmethod
    def obtener_todos(self, sort_by: str = 'nombre', filter_by: str = None) -> list[Receta]:
        """Obtiene todas las recetas del repositorio."""
        pass
    
    @abstractmethod
    def obtener_por_id(self, id: str) -> Receta:
        """Obtiene una receta por su ID."""
        pass
    
    @abstractmethod
    def actualizar(self, receta: Receta) -> Receta:
        """Actualiza una receta existente."""
        pass
    
    @abstractmethod
    def eliminar(self, id: str) -> bool:
        """Elimina una receta por su ID."""
        pass

    @abstractmethod
    def find_by_name(self, nombre: str):
        """Obtiene una receta por su nombre."""
        pass

    @abstractmethod
    def crear_multiples(self, recetas: list[Receta]) -> list[Receta]:
        """Crea múltiples recetas en el repositorio."""
        pass


# Caso de uso para crear una nueva receta
class CrearRecetaUseCase:
    def __init__(self, repository: IRecetaRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, receta_data: dict) -> Receta:
        """
        Ejecuta la creación de una receta, evitando duplicados por nombre.
        """
        nombre_receta = receta_data.get("nombre").upper()
        if self.repository.find_by_name(nombre_receta):
            raise ValueError(f"La receta con el nombre '{nombre_receta}' ya existe.")

        receta_data['nombre'] = nombre_receta
        receta = Receta.from_dict(receta_data)
        receta.ingredientes = [
            Ingrediente.from_dict(ing) for ing in receta_data.get("ingredientes", [])
        ]
        nueva_receta = self.repository.crear(receta)

        self.registrar_cambio_uc.execute(
            entidad_tipo='Receta',
            entidad_id=nueva_receta.id,
            campo_modificado='Creación',
            valor_anterior='',
            valor_nuevo=f"Receta '{nueva_receta.nombre}' creada"
        )

        return nueva_receta

# Caso de uso para obtener todas las recetas
class ObtenerRecetasUseCase:
    def __init__(self, repository: IRecetaRepository):
        self.repository = repository
        
    def execute(self, sort_by: str = 'nombre', filter_by: str = None) -> list[Receta]:
        """Ejecuta la obtención de todas las recetas."""
        return self.repository.obtener_todos(sort_by=sort_by, filter_by=filter_by)

# Caso de uso para obtener una receta por su ID
class ObtenerRecetaPorIdUseCase:
    def __init__(self, repository: IRecetaRepository):
        self.repository = repository
        
    def execute(self, id: str) -> Receta:
        """Ejecuta la obtención de una receta específica por su ID."""
        return self.repository.obtener_por_id(id)

# Caso de uso para actualizar una receta
class ActualizarRecetaUseCase:
    def __init__(self, repository: IRecetaRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, id: str, receta_data: dict) -> Receta:
        """
        Ejecuta la actualización de una receta.
        Combina los datos nuevos con el ID para crear un objeto Receta y actualizarlo.
        """
        receta_data['nombre'] = receta_data.get('nombre').upper()
        receta = Receta.from_dict({**receta_data, "id": id})
        receta_actual = self.repository.obtener_por_id(id)
        if not receta_actual:
            return None

        receta_actualizada = Receta.from_dict({**receta_data, "id": id})
        receta_actualizada.ingredientes = [
            Ingrediente.from_dict(ing) for ing in receta_data.get("ingredientes", [])
        ]

        if receta_actual.nombre != receta_actualizada.nombre:
            self.registrar_cambio_uc.execute(
                entidad_tipo='Receta',
                entidad_id=id,
                campo_modificado='nombre',
                valor_anterior=receta_actual.nombre,
                valor_nuevo=receta_actualizada.nombre
            )
        
        # Simplificado para el ejemplo, se puede expandir para los ingredientes
        if len(receta_actual.ingredientes) != len(receta_actualizada.ingredientes):
             self.registrar_cambio_uc.execute(
                entidad_tipo='Receta',
                entidad_id=id,
                campo_modificado='ingredientes',
                valor_anterior=f"Receta '{receta_actual.nombre}' tenía {len(receta_actual.ingredientes)} ingredientes",
                valor_nuevo=f"Receta '{receta_actualizada.nombre}' ahora tiene {len(receta_actualizada.ingredientes)} ingredientes"
            )

        return self.repository.actualizar(receta_actualizada)

# Caso de uso para eliminar una receta
class EliminarRecetaUseCase:
    def __init__(self, repository: IRecetaRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, id: str) -> bool:
        """Ejecuta la eliminación de una receta por su ID."""
        receta = self.repository.obtener_por_id(id)
        if not receta:
            return False

        # Registrar el cambio antes de eliminar
        self.registrar_cambio_uc.execute(
            entidad_tipo='Receta',
            entidad_id=id,
            campo_modificado='Eliminación',
            valor_anterior=f"Receta '{receta.nombre}' eliminada",
            valor_nuevo=''
        )
        
        return self.repository.eliminar(id)

# Caso de uso para importar recetas desde una lista de diccionarios
class ImportarRecetasUseCase:
    def __init__(self, repository: IRecetaRepository):
        self.repository = repository
        
    def execute(self, recetas_data: list[dict]) -> dict:
        """
        Ejecuta la importación masiva de recetas.
        Verifica si las recetas ya existen por nombre para evitar duplicados.
        Retorna un resumen de las recetas importadas y omitidas.
        """
        importadas = 0
        omitidas = 0
        for receta_data in recetas_data:
            nombre_receta = receta_data.get("nombre")
            if not nombre_receta:
                omitidas += 1
                continue

            # Evitar duplicados por nombre
            existente = self.repository.find_by_name(nombre_receta)
            if existente:
                omitidas += 1
                continue

            receta = Receta.from_dict(receta_data)
            receta.ingredientes = [
                Ingrediente.from_dict(ing) for ing in receta_data.get("ingredientes", [])
            ]
            self.repository.crear(receta)
            importadas += 1
            
        return {"importadas": importadas, "omitidas": omitidas}

from src.application.use_cases.producto_use_cases import IProductoRepository
from src.application.use_cases.area_use_cases import IAreaRepository
from src.core.domain import Area, Producto

# Caso de uso para exportar recetas a Excel
class ExportRecetasExcel:
    def __init__(self, repository: IRecetaRepository, producto_repository: IProductoRepository, area_repository: IAreaRepository):
        self.repository = repository
        self.producto_repository = producto_repository
        self.area_repository = area_repository

    def execute(self):
        recetas = self.repository.obtener_todos()
        
        recetas_data = []
        for r in recetas:
            if not r.ingredientes:
                recetas_data.append({
                    "receta_nombre": r.nombre,
                    "producto_nombre": "",
                    "unidad_medida": "",
                    "cantidad": "",
                    "area_nombre": ""
                })
            else:
                for i in r.ingredientes:
                    producto = self.producto_repository.obtener_por_id(i.producto_id)
                    area = self.area_repository.obtener_por_id(i.area_id)
                    recetas_data.append({
                        "receta_nombre": r.nombre,
                        "producto_nombre": producto.nombre if producto else "Producto no encontrado",
                        "unidad_medida": producto.unidad_medida if producto else "",
                        "cantidad": i.cantidad,
                        "area_nombre": area.nombre if area else "Área no encontrada"
                    })
        
        df = pd.DataFrame(recetas_data)
        
        output = io.BytesIO()
        with pd.ExcelWriter(output, engine='openpyxl') as writer:
            df.to_excel(writer, index=False, sheet_name='Recetas')
        
        output.seek(0)
        return output

# Caso de uso para importar recetas desde Excel
class ImportRecetasExcel:
    def __init__(self, repository: IRecetaRepository, producto_repository: IProductoRepository, area_repository: IAreaRepository):
        self.repository = repository
        self.producto_repository = producto_repository
        self.area_repository = area_repository

    def execute(self, file):
        df = pd.read_excel(file).fillna('')
        
        required_columns = ["receta_nombre", "producto_nombre", "unidad_medida", "cantidad", "area_nombre"]
        if not all(col in df.columns for col in required_columns):
            raise ValueError(f"El archivo Excel debe contener las columnas: {', '.join(required_columns)}")

        recetas_a_crear = {}
        for _, row in df.iterrows():
            receta_nombre = row["receta_nombre"]
            if not receta_nombre:
                continue

            if receta_nombre not in recetas_a_crear:
                if not self.repository.find_by_name(receta_nombre):
                    recetas_a_crear[receta_nombre] = {"nombre": receta_nombre, "ingredientes": []}

            if receta_nombre in recetas_a_crear and row["producto_nombre"]:
                producto = self.producto_repository.find_by_name(row["producto_nombre"])
                if not producto:
                    producto_data = {
                        "nombre": row["producto_nombre"],
                        "unidad_medida": row["unidad_medida"]
                    }
                    producto = self.producto_repository.crear(Producto.from_dict(producto_data))

                area = self.area_repository.find_by_name(row["area_nombre"])
                if not area:
                    area_data = {"nombre": row["area_nombre"]}
                    area = self.area_repository.crear(Area.from_dict(area_data))
                
                recetas_a_crear[receta_nombre]["ingredientes"].append({
                    "producto_id": producto.id,
                    "area_id": area.id,
                    "cantidad": row["cantidad"]
                })

        for receta_data in recetas_a_crear.values():
            if not self.repository.find_by_name(receta_data["nombre"]):
                receta = Receta.from_dict(receta_data)
                receta.ingredientes = [
                    Ingrediente.from_dict(ing) for ing in receta_data.get("ingredientes", [])
                ]
                self.repository.crear(receta)

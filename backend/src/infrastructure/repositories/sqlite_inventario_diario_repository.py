from datetime import date, timedelta
from sqlalchemy.orm import Session, joinedload, lazyload
from src.core.domain.inventario_diario import InventarioDiario
from src.infrastructure.db.models import InventarioDiario as InventarioDiarioModel, ModeloIPV as ModeloIPVModel
from src.infrastructure.db.models import Producto as ProductoModel
from src.infrastructure.db.models import Area as AreaModel

# Repositorio para gestionar los datos del inventario diario en la base de datos SQLite.
class SQLiteInventarioDiarioRepository:
    def __init__(self, db_session: Session):
        self.db_session = db_session

    # Busca un registro de inventario por fecha, área y producto.
    def find_by_date_area_producto(self, fecha: date, area_id: str, producto_id: str) -> InventarioDiario | None:
        model = self.db_session.query(InventarioDiarioModel).filter_by(
            fecha=fecha, area_id=area_id, producto_id=producto_id
        ).first()
        return self._to_domain(model) if model else None

    # Obtiene todos los registros de inventario para una fecha específica.
    def find_by_date(self, fecha: date) -> list[InventarioDiarioModel]:
        return self.db_session.query(InventarioDiarioModel).options(
            lazyload(InventarioDiarioModel.producto),
            lazyload(InventarioDiarioModel.area)
        ).filter_by(fecha=fecha).all()

    # Encuentra todas las fechas únicas de los registros de inventario.
    def find_all_dates(self) -> list[date]:
        return [result[0] for result in self.db_session.query(InventarioDiarioModel.fecha).distinct().order_by(InventarioDiarioModel.fecha.desc()).all()]

    # Obtiene el inventario físico final del día anterior para un producto y área.
    def get_inicio_from_previous_day(self, fecha: date, area_id: str, producto_id: str) -> float:
        dia_anterior = fecha - timedelta(days=1)
        model = self.db_session.query(InventarioDiarioModel).filter_by(
            fecha=dia_anterior, area_id=area_id, producto_id=producto_id
        ).first()
        return model.final_fisico if model else 0.0

    # Guarda o actualiza una lista de registros de inventario diario.
    def save_all(self, inventarios: list[InventarioDiario]):
        for domain_obj in inventarios:
            model = self.db_session.query(InventarioDiarioModel).filter_by(
                fecha=domain_obj.fecha, area_id=domain_obj.area_id, producto_id=domain_obj.producto_id
            ).first()

            if model:
                # Actualiza el modelo existente.
                model.inicio = domain_obj.inicio
                model.entradas = domain_obj.entradas
                model.consumo = domain_obj.consumo
                model.merma = domain_obj.merma
                model.otras_salidas = domain_obj.otras_salidas
                model.final_fisico = domain_obj.final_fisico
                model.final_teorico = domain_obj.final_teorico
                model.diferencia = domain_obj.diferencia
                model.comentario = domain_obj.comentario
            else:
                # Crea un nuevo modelo si no existe.
                model = self._to_model(domain_obj)
                self.db_session.add(model)
        
        self.db_session.commit()

    # Convierte un modelo de base de datos a un objeto de dominio.
    def _to_domain(self, model: InventarioDiarioModel) -> InventarioDiario:
        if not model:
            return None
        return InventarioDiario(
            id=model.id,
            fecha=model.fecha,
            area_id=model.area_id,
            producto_id=model.producto_id,
            inicio=model.inicio,
            entradas=model.entradas,
            consumo=model.consumo,
            merma=model.merma,
            otras_salidas=model.otras_salidas,
            final_fisico=model.final_fisico,
            final_teorico=model.final_teorico,
            diferencia=model.diferencia,
            comentario=model.comentario,
            producto_nombre=model.producto.nombre if model.producto else "Producto no encontrado",
            area_nombre=model.area.nombre if model.area else "Área no encontrada"
        )

    # Convierte un objeto de dominio a un modelo de base de datos.
    def _to_model(self, domain_obj: InventarioDiario) -> InventarioDiarioModel:
        return InventarioDiarioModel(
            id=domain_obj.id,
            fecha=domain_obj.fecha,
            area_id=domain_obj.area_id,
            producto_id=domain_obj.producto_id,
            inicio=domain_obj.inicio,
            entradas=domain_obj.entradas,
            consumo=domain_obj.consumo,
            merma=domain_obj.merma,
            otras_salidas=domain_obj.otras_salidas,
            final_fisico=domain_obj.final_fisico,
            final_teorico=domain_obj.final_teorico,
            diferencia=domain_obj.diferencia,
            comentario=domain_obj.comentario
        )

    # Obtiene todos los modelos de IPV.
    def get_modelos(self) -> dict[str, list[dict]]:
        modelos = self.db_session.query(ModeloIPVModel).order_by(ModeloIPVModel.orden).all()
        modelos_dict = {}
        for modelo in modelos:
            area_id_str = str(modelo.area_id)
            if area_id_str not in modelos_dict:
                modelos_dict[area_id_str] = []
            modelos_dict[area_id_str].append({
                "producto_id": str(modelo.producto_id),
                "orden": modelo.orden
            })
        return modelos_dict

    # Guarda un modelo de IPV para un área.
    def save_modelo(self, area_id: str, productos: list[dict]):
        # Elimina el modelo existente para el área.
        self.db_session.query(ModeloIPVModel).filter_by(area_id=area_id).delete()
        
        # Crea los nuevos registros del modelo.
        for producto_data in productos:
            nuevo_modelo = ModeloIPVModel(
                area_id=area_id,
                producto_id=producto_data['id'],
                orden=producto_data['orden']
            )
            self.db_session.add(nuevo_modelo)
            
        self.db_session.commit()

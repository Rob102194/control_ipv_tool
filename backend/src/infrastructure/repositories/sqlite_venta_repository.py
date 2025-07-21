from datetime import datetime
from src.core.domain.venta import Venta
from src.application.use_cases.venta_use_cases import IVentaRepository
from src.infrastructure.db import models as db_models

# Implementación del repositorio de ventas para SQLite
class SQLiteVentaRepository(IVentaRepository):
    def __init__(self, db_session):
        """Inicializa el repositorio con una sesión de base de datos."""
        self.db_session = db_session

    def crear(self, venta: Venta) -> Venta:
        """Crea una nueva venta en la base de datos."""
        try:
            fecha_obj = datetime.strptime(venta.fecha, '%Y-%m-%d').date() if isinstance(venta.fecha, str) else venta.fecha
            venta_db = db_models.Venta(
                receta_nombre=venta.receta_nombre,
                cantidad=venta.cantidad,
                fecha=fecha_obj
            )
            self.db_session.add(venta_db)
            self.db_session.commit()
            # Retorna el objeto de dominio con el ID asignado por la BD
            return Venta(
                receta_nombre=venta_db.receta_nombre,
                cantidad=venta_db.cantidad,
                fecha=venta_db.fecha.isoformat(),
                id=str(venta_db.id)
            )
        except Exception as e:
            self.db_session.rollback()
            raise e

    def obtener_todos(self) -> list[Venta]:
        """Obtiene todas las ventas de la base de datos."""
        ventas_db = self.db_session.query(db_models.Venta).all()
        # Mapea los resultados de la BD a objetos de dominio
        return [
            Venta(
                receta_nombre=v.receta_nombre,
                cantidad=v.cantidad,
                fecha=v.fecha.isoformat(),
                id=str(v.id)
            ) for v in ventas_db
        ]

    def obtener_por_id(self, id: str) -> Venta:
        """Obtiene una venta por su ID."""
        venta_db = self.db_session.query(db_models.Venta).get(id)
        if not venta_db:
            return None
        # Mapea el resultado a un objeto de dominio
        return Venta(
            receta_nombre=venta_db.receta_nombre,
            cantidad=venta_db.cantidad,
            fecha=venta_db.fecha.isoformat(),
            id=str(venta_db.id)
        )

    def actualizar(self, venta: Venta) -> Venta:
        """Actualiza una venta existente en la base de datos."""
        try:
            venta_db = self.db_session.query(db_models.Venta).get(venta.id)
            if not venta_db:
                return None
            
            fecha_obj = datetime.strptime(venta.fecha, '%Y-%m-%d').date() if isinstance(venta.fecha, str) else venta.fecha
            venta_db.receta_nombre = venta.receta_nombre
            venta_db.cantidad = venta.cantidad
            venta_db.fecha = fecha_obj
            self.db_session.commit()
            # Retorna el objeto de dominio actualizado
            return Venta(
                receta_nombre=venta_db.receta_nombre,
                cantidad=venta_db.cantidad,
                fecha=venta_db.fecha.isoformat(),
                id=str(venta_db.id)
            )
        except Exception as e:
            self.db_session.rollback()
            raise e

    def eliminar(self, id: str) -> bool:
        """Elimina una venta de la base de datos por su ID."""
        try:
            venta_db = self.db_session.query(db_models.Venta).get(id)
            if not venta_db:
                return False
            
            self.db_session.delete(venta_db)
            self.db_session.commit()
            return True
        except Exception as e:
            self.db_session.rollback()
            raise e

    def eliminar_multiples(self, ids: list[str]) -> bool:
        """Elimina múltiples ventas de la base de datos por sus IDs."""
        try:
            self.db_session.query(db_models.Venta).filter(db_models.Venta.id.in_(ids)).delete(synchronize_session=False)
            self.db_session.commit()
            return True
        except Exception as e:
            self.db_session.rollback()
            raise e

    def find_by_date(self, fecha: datetime.date) -> list[Venta]:
        """Obtiene todas las ventas para una fecha específica."""
        ventas_db = self.db_session.query(db_models.Venta).filter_by(fecha=fecha).all()
        return [
            Venta(
                receta_nombre=v.receta_nombre,
                cantidad=v.cantidad,
                fecha=v.fecha.isoformat(),
                id=str(v.id)
            ) for v in ventas_db
        ]

    def crear_multiples(self, ventas: list[Venta]) -> list[Venta]:
        """Crea múltiples ventas en la base de datos de forma transaccional."""
        try:
            ventas_db = []
            for venta in ventas:
                fecha_obj = datetime.strptime(venta.fecha, '%Y-%m-%d').date() if isinstance(venta.fecha, str) else venta.fecha
                venta_db = db_models.Venta(
                    receta_nombre=venta.receta_nombre,
                    cantidad=venta.cantidad,
                    fecha=fecha_obj
                )
                ventas_db.append(venta_db)
            
            self.db_session.add_all(ventas_db)
            self.db_session.commit()
            
            # Asigna los IDs generados a las ventas de entrada
            for i, venta_db in enumerate(ventas_db):
                ventas[i].id = str(venta_db.id)
            
            return ventas
        except Exception as e:
            self.db_session.rollback()
            raise e

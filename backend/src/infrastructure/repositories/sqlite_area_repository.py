from src.core.domain.area import Area
from src.application.use_cases.area_use_cases import IAreaRepository
from src.infrastructure.db import models as db_models

# Implementación del repositorio de áreas para SQLite
class SQLiteAreaRepository(IAreaRepository):
    def __init__(self, db_session):
        """Inicializa el repositorio con una sesión de base de datos."""
        self.db_session = db_session
        
    def crear(self, area: Area) -> Area:
        """Crea una nueva área en la base de datos."""
        try:
            area_db = db_models.Area(
                nombre=area.nombre,
                codigo=area.codigo
            )
            self.db_session.add(area_db)
            self.db_session.commit()
            # Retorna el objeto de dominio con el ID asignado por la BD
            return Area(
                id=str(area_db.id),
                nombre=area_db.nombre,
                codigo=area_db.codigo
            )
        except Exception as e:
            self.db_session.rollback()
            raise e
            
    def find_by_name(self, nombre: str) -> Area:
        """Obtiene un área por su nombre."""
        area_db = self.db_session.query(db_models.Area).filter_by(nombre=nombre).first()
        if not area_db:
            return None
        return Area(
            id=str(area_db.id),
            nombre=area_db.nombre,
            codigo=area_db.codigo
        )
    
    def find_all(self) -> list[Area]:
        """Obtiene todas las áreas de la base de datos."""
        areas_db = self.db_session.query(db_models.Area).all()
        # Mapea los resultados de la BD a objetos de dominio
        return [
            Area(
                id=str(a.id),
                nombre=a.nombre,
                codigo=a.codigo
            ) for a in areas_db
        ]
    
    def obtener_por_id(self, id: str) -> Area:
        """Obtiene un área por su ID."""
        area_db = self.db_session.query(db_models.Area).get(id)
        if not area_db:
            return None
        # Mapea el resultado a un objeto de dominio
        return Area(
            id=str(area_db.id),
            nombre=area_db.nombre,
            codigo=area_db.codigo
        )
    
    def actualizar(self, area: Area) -> Area:
        """Actualiza un área existente en la base de datos."""
        try:
            area_db = self.db_session.query(db_models.Area).get(area.id)
            if not area_db:
                return None
                
            area_db.nombre = area.nombre
            area_db.codigo = area.codigo
            self.db_session.commit()
            # Retorna el objeto de dominio actualizado
            return Area(
                id=str(area_db.id),
                nombre=area_db.nombre,
                codigo=area_db.codigo
            )
        except Exception as e:
            self.db_session.rollback()
            raise e
    
    def eliminar(self, id: str) -> bool:
        """Elimina un área de la base de datos por su ID."""
        try:
            area_db = self.db_session.query(db_models.Area).get(id)
            if not area_db:
                return False
                
            self.db_session.delete(area_db)
            self.db_session.commit()
            return True
        except Exception as e:
            self.db_session.rollback()
            raise e

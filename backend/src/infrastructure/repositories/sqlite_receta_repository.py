from src.core.domain import Receta, Ingrediente
from src.application.use_cases.receta_use_cases import IRecetaRepository
from src.infrastructure.db import models as db_models

# Implementación del repositorio de recetas para SQLite
class SQLiteRecetaRepository(IRecetaRepository):
    def __init__(self, db_session):
        """Inicializa el repositorio con una sesión de base de datos."""
        self.db_session = db_session
        
    def crear(self, receta: Receta) -> Receta:
        """
        Crea una nueva receta en la base de datos.
        Maneja la creación de la receta y sus ingredientes asociados de forma transaccional.
        """
        try:
            # Mapeo del objeto de dominio a modelo de base de datos
            receta_db = db_models.Receta(
                nombre=receta.nombre,
                activa=receta.activa
            )
            self.db_session.add(receta_db)
            self.db_session.flush()  # Para obtener el ID de la receta antes del commit
            
            # Creación de los ingredientes asociados
            for ingrediente in receta.ingredientes:
                ingrediente_db = db_models.Ingrediente(
                    receta_id=receta_db.id,
                    producto_id=ingrediente.producto_id,
                    area_id=ingrediente.area_id,
                    cantidad=ingrediente.cantidad
                )
                self.db_session.add(ingrediente_db)
                
            self.db_session.commit()
            
            # Actualiza el objeto de dominio con los IDs generados por la base de datos
            receta.id = str(receta_db.id)
            for i, ingrediente_db in enumerate(receta_db.ingredientes):
                receta.ingredientes[i].id = str(ingrediente_db.id)
                
            return receta
        except Exception as e:
            self.db_session.rollback()  # Revertir cambios en caso de error
            raise e
    
    def obtener_todos(self, sort_by: str = 'nombre', filter_by: str = None) -> list[Receta]:
        """
        Obtiene todas las recetas de la base de datos y las convierte a objetos de dominio.
        """
        try:
            query = self.db_session.query(db_models.Receta)

            if filter_by == 'sin_ingredientes':
                query = query.outerjoin(db_models.Ingrediente).filter(db_models.Ingrediente.id == None)

            if sort_by == 'modificado':
                query = query.outerjoin(db_models.HistorialCambios, db_models.Receta.id == db_models.HistorialCambios.entidad_id).order_by(db_models.HistorialCambios.fecha_cambio.desc())
            elif sort_by == 'nombre':
                query = query.order_by(db_models.Receta.nombre)
            else:
                query = query.order_by(db_models.Receta.id.desc())

            recetas_db = query.all()
            recetas = []
            for receta_db in recetas_db:
                # Mapeo del modelo de base de datos a objeto de dominio
                receta = Receta(
                    nombre=receta_db.nombre,
                    activa=receta_db.activa,
                    id=str(receta_db.id)
                )
                receta.ingredientes = [
                    Ingrediente(
                        producto_id=str(ing.producto_id),
                        area_id=str(ing.area_id),
                        cantidad=ing.cantidad,
                        id=str(ing.id),
                        receta_id=str(ing.receta_id)
                    ) for ing in receta_db.ingredientes
                ]
                recetas.append(receta)
            return recetas
        except Exception as e:
            raise e

    def crear_multiples(self, recetas: list[Receta]) -> list[Receta]:
        """
        Crea múltiples recetas en la base de datos de forma transaccional.
        """
        try:
            recetas_db = []
            for receta in recetas:
                receta_db = db_models.Receta(
                    nombre=receta.nombre,
                    activa=receta.activa
                )
                recetas_db.append(receta_db)
            
            self.db_session.add_all(recetas_db)
            self.db_session.commit()
            
            for i, receta_db in enumerate(recetas_db):
                recetas[i].id = str(receta_db.id)
            
            return recetas
        except Exception as e:
            self.db_session.rollback()
            raise e
    def obtener_por_id(self, id: str) -> Receta:
        """
        Obtiene una receta específica por su ID y la convierte a un objeto de dominio.
        """
        try:
            receta_db = self.db_session.query(db_models.Receta).get(id)
            if not receta_db:
                return None
                
            # Mapeo del modelo de base de datos a objeto de dominio
            receta = Receta(
                nombre=receta_db.nombre,
                activa=receta_db.activa,
                id=str(receta_db.id)
            )
            receta.ingredientes = [
                Ingrediente(
                    producto_id=str(ing.producto_id),
                    area_id=str(ing.area_id),
                    cantidad=ing.cantidad,
                    id=str(ing.id),
                    receta_id=str(ing.receta_id)
                ) for ing in receta_db.ingredientes
            ]
            return receta
        except Exception as e:
            raise e
    
    def actualizar(self, receta: Receta) -> Receta:
        """
        Actualiza una receta existente en la base de datos.
        Elimina los ingredientes antiguos y los reemplaza con los nuevos.
        """
        try:
            receta_db = self.db_session.query(db_models.Receta).get(receta.id)
            if not receta_db:
                return None
                
            receta_db.nombre = receta.nombre
            receta_db.activa = receta.activa
            
            # Estrategia de actualización: eliminar y volver a crear ingredientes
            self.db_session.query(db_models.Ingrediente).filter_by(receta_id=receta.id).delete()
            
            for ingrediente in receta.ingredientes:
                ingrediente_db = db_models.Ingrediente(
                    receta_id=receta.id,
                    producto_id=ingrediente.producto_id,
                    area_id=ingrediente.area_id,
                    cantidad=ingrediente.cantidad
                )
                self.db_session.add(ingrediente_db)
                
            self.db_session.commit()
            
            return receta
        except Exception as e:
            self.db_session.rollback()
            raise e
    
    def eliminar(self, id: str) -> bool:
        """
        Elimina una receta de la base de datos por su ID.
        La eliminación de ingredientes se maneja en cascada por la configuración de la BD.
        """
        try:
            receta_db = self.db_session.query(db_models.Receta).get(id)
            if not receta_db:
                return False
                
            self.db_session.delete(receta_db)
            self.db_session.commit()
            return True
        except Exception as e:
            self.db_session.rollback()
            raise e


    def find_by_name(self, nombre: str) -> Receta:
        """
        Obtiene una receta por su nombre. Útil para evitar duplicados.
        """
        try:
            receta_db = self.db_session.query(db_models.Receta).filter_by(nombre=nombre).first()
            if not receta_db:
                return None
            
            # Mapeo del modelo de base de datos a objeto de dominio
            receta = Receta(
                nombre=receta_db.nombre,
                activa=receta_db.activa,
                id=str(receta_db.id)
            )
            receta.ingredientes = [
                Ingrediente(
                    producto_id=str(ing.producto_id),
                    area_id=str(ing.area_id),
                    cantidad=ing.cantidad,
                    id=str(ing.id),
                    receta_id=str(ing.receta_id)
                ) for ing in receta_db.ingredientes
            ]
            return receta
        except Exception as e:
            raise e

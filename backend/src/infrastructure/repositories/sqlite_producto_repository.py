from src.core.domain.producto import Producto
from src.application.use_cases.producto_use_cases import IProductoRepository
from src.infrastructure.db import models as db_models

# Implementación del repositorio de productos para SQLite
class SQLiteProductoRepository(IProductoRepository):
    def __init__(self, db_session):
        """Inicializa el repositorio con una sesión de base de datos."""
        self.db_session = db_session
        
    def crear(self, producto: Producto) -> Producto:
        """Crea un nuevo producto en la base de datos."""
        try:
            producto_db = db_models.Producto(
                nombre=producto.nombre,
                unidad_medida=producto.unidad_medida
            )
            self.db_session.add(producto_db)
            self.db_session.commit()
            # Retorna el objeto de dominio con el ID asignado por la BD
            return Producto(
                id=str(producto_db.id),
                nombre=producto_db.nombre,
                unidad_medida=producto_db.unidad_medida
            )
        except Exception as e:
            self.db_session.rollback()
            raise e
    
    def find_by_name(self, nombre: str) -> Producto:
        """Obtiene un producto por su nombre."""
        producto_db = self.db_session.query(db_models.Producto).filter_by(nombre=nombre).first()
        if not producto_db:
            return None
        return Producto(
            id=str(producto_db.id),
            nombre=producto_db.nombre,
            unidad_medida=producto_db.unidad_medida
        )

    def obtener_todos(self, sort_by: str = 'nombre') -> list[Producto]:
        """Obtiene todos los productos de la base de datos."""
        if sort_by == 'modificado':
            query = self.db_session.query(db_models.Producto).outerjoin(db_models.HistorialCambios, db_models.Producto.id == db_models.HistorialCambios.entidad_id).order_by(db_models.HistorialCambios.fecha_cambio.desc())
        elif sort_by == 'nombre':
            query = self.db_session.query(db_models.Producto).order_by(db_models.Producto.nombre)
        else:
            query = self.db_session.query(db_models.Producto).order_by(db_models.Producto.id.desc())

        productos_db = query.all()
        # Mapea los resultados de la BD a objetos de dominio
        return [
            Producto(
                id=str(p.id),
                nombre=p.nombre,
                unidad_medida=p.unidad_medida
            ) for p in productos_db
        ]
    
    def obtener_por_id(self, id: str) -> Producto:
        """Obtiene un producto por su ID."""
        producto_db = self.db_session.query(db_models.Producto).filter_by(id=str(id)).first()
        if not producto_db:
            return None
        # Mapea el resultado a un objeto de dominio
        return Producto(
            id=str(producto_db.id),
            nombre=producto_db.nombre,
            unidad_medida=producto_db.unidad_medida
        )
    
    def actualizar(self, producto: Producto) -> Producto:
        """Actualiza un producto existente en la base de datos."""
        try:
            producto_db = self.db_session.query(db_models.Producto).get(producto.id)
            if not producto_db:
                return None
                
            producto_db.nombre = producto.nombre
            producto_db.unidad_medida = producto.unidad_medida
            self.db_session.commit()
            # Retorna el objeto de dominio actualizado
            return Producto(
                id=str(producto_db.id),
                nombre=producto_db.nombre,
                unidad_medida=producto_db.unidad_medida
            )
        except Exception as e:
            self.db_session.rollback()
            raise e
    
    def eliminar(self, id: str) -> bool:
        """Elimina un producto de la base de datos por su ID."""
        try:
            producto_db = self.db_session.query(db_models.Producto).get(id)
            if not producto_db:
                return False
                
            self.db_session.delete(producto_db)
            self.db_session.commit()
            return True
        except Exception as e:
            self.db_session.rollback()
            raise e

    def producto_en_uso(self, producto_id: str) -> bool:
        """Verifica si un producto está en uso en alguna de las tablas dependientes."""
        tablas_dependientes = [
            db_models.Ingrediente,
            db_models.MovimientoInventario,
            db_models.InventarioDiario,
            db_models.ModeloIPV
        ]
        
        for tabla in tablas_dependientes:
            en_uso = self.db_session.query(tabla).filter_by(producto_id=producto_id).first()
            if en_uso:
                return True
        
        return False

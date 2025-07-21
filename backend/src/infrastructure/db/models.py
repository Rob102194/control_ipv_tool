from flask_sqlalchemy import SQLAlchemy
import uuid

# Inicialización de la extensión SQLAlchemy para la base de datos.
db = SQLAlchemy()

# Función para generar identificadores únicos universales (UUIDs) para las claves primarias.
def generate_uuid():
    return str(uuid.uuid4())

# Modelo para los productos del inventario.
class Producto(db.Model):
    __tablename__ = 'productos'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    nombre = db.Column(db.String(100), unique=True, nullable=False)
    unidad_medida = db.Column(db.String(10), nullable=False)  # Ej: kg, g, l, unidades

# Modelo para las áreas del restaurante (ej: Cocina, Bar).
class Area(db.Model):
    __tablename__ = 'areas'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    nombre = db.Column(db.String(50), unique=True, nullable=False)
    codigo = db.Column(db.String(10))

# Modelo para las recetas de los platos que se venden.
class Receta(db.Model):
    __tablename__ = 'recetas'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    nombre = db.Column(db.String(100), unique=True, nullable=False)
    activa = db.Column(db.Boolean, default=True)
    # Relación uno a muchos con los ingredientes de la receta.
    ingredientes = db.relationship('Ingrediente', back_populates='receta', cascade="all, delete-orphan")

# Modelo para los ingredientes que componen una receta.
class Ingrediente(db.Model):
    __tablename__ = 'ingredientes'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    receta_id = db.Column(db.String(36), db.ForeignKey('recetas.id'), nullable=False)
    producto_id = db.Column(db.String(36), db.ForeignKey('productos.id'), nullable=False)
    area_id = db.Column(db.String(36), db.ForeignKey('areas.id'), nullable=False)
    cantidad = db.Column(db.Float, nullable=False)
    
    # Relaciones para acceder fácilmente a la receta, producto y área asociados.
    receta = db.relationship('Receta', back_populates='ingredientes')
    producto = db.relationship('Producto')
    area = db.relationship('Area')

# Modelo para registrar los movimientos de inventario (entradas y salidas).
class MovimientoInventario(db.Model):
    __tablename__ = 'movimientos'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    tipo = db.Column(db.Enum('ENTRADA', 'SALIDA', 'INICIAL', name='tipo_movimiento'), nullable=False)
    producto_id = db.Column(db.String(36), db.ForeignKey('productos.id'), nullable=False)
    area_id = db.Column(db.String(36), db.ForeignKey('areas.id'), nullable=False)
    cantidad = db.Column(db.Float, nullable=False)
    fecha = db.Column(db.DateTime, default=db.func.current_timestamp())
    motivo = db.Column(db.Enum('compra', 'merma', 'transferencia', 'inicial', 'ajuste', name='motivo_movimiento'))
    comentarios = db.Column(db.Text)

# Modelo para registrar las ventas diarias.
class Venta(db.Model):
    __tablename__ = 'ventas'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    receta_nombre = db.Column(db.String(100), nullable=False)
    cantidad = db.Column(db.Integer, nullable=False)
    fecha = db.Column(db.Date, nullable=False, default=db.func.current_date())
    
    # Índice para acelerar las búsquedas por fecha.
    __table_args__ = (db.Index('idx_venta_fecha', 'fecha'),)

# Modelo para el registro del inventario diario (IPV).
class InventarioDiario(db.Model):
    __tablename__ = 'inventario_diario'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    fecha = db.Column(db.Date, nullable=False)
    area_id = db.Column(db.String(36), db.ForeignKey('areas.id'), nullable=False)
    producto_id = db.Column(db.String(36), db.ForeignKey('productos.id'), nullable=False)
    inicio = db.Column(db.Float, default=0.0)
    entradas = db.Column(db.Float, default=0.0)
    consumo = db.Column(db.Float, default=0.0)
    merma = db.Column(db.Float, default=0.0)
    otras_salidas = db.Column(db.Float, default=0.0)
    final_fisico = db.Column(db.Float, default=0.0)
    final_teorico = db.Column(db.Float, default=0.0)
    diferencia = db.Column(db.Float, default=0.0)
    comentario = db.Column(db.Text, nullable=True)

    # Relaciones para acceder al área y producto asociados.
    area = db.relationship('Area')
    producto = db.relationship('Producto')

    # Restricciones y índices de la tabla.
    __table_args__ = (
        # Restricción para asegurar que no haya registros duplicados para el mismo producto, área y fecha.
        db.UniqueConstraint('fecha', 'area_id', 'producto_id', name='_fecha_area_producto_uc'),
        # Índice para acelerar las búsquedas por fecha.
        db.Index('idx_inventario_fecha', 'fecha'),
    )

# Modelo para definir qué productos se incluyen en el inventario de cada área.
class ModeloIPV(db.Model):
    __tablename__ = 'modelo_ipv'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    area_id = db.Column(db.String(36), db.ForeignKey('areas.id'), nullable=False)
    producto_id = db.Column(db.String(36), db.ForeignKey('productos.id'), nullable=False)
    orden = db.Column(db.Integer, nullable=False, default=0)

    # Relaciones para acceder al área y producto asociados.
    area = db.relationship('Area')
    producto = db.relationship('Producto')

    # Restricción para asegurar que no haya productos duplicados en el modelo de un área.
    __table_args__ = (
        db.UniqueConstraint('area_id', 'producto_id', name='_area_producto_uc'),
    )

# Modelo para el historial de cambios
class HistorialCambios(db.Model):
    __tablename__ = 'historial_cambios'
    id = db.Column(db.String(36), primary_key=True, default=generate_uuid)
    entidad_tipo = db.Column(db.String(50), nullable=False)
    entidad_id = db.Column(db.String(36), nullable=False)
    campo_modificado = db.Column(db.String(50), nullable=False)
    valor_anterior = db.Column(db.String(255))
    valor_nuevo = db.Column(db.String(255))
    fecha_cambio = db.Column(db.DateTime, default=db.func.current_timestamp())

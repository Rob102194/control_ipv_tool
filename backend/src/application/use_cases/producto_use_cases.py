from abc import ABC, abstractmethod
from src.core.domain.producto import Producto
from src.application.use_cases.historial_use_cases import RegistrarCambioUseCase
import pandas as pd
import io

# Interfaz abstracta para el repositorio de productos
class IProductoRepository(ABC):
    @abstractmethod
    def crear(self, producto: Producto) -> Producto:
        """Crea un nuevo producto."""
        pass
    
    @abstractmethod
    def obtener_todos(self, sort_by: str = 'nombre') -> list[Producto]:
        """Obtiene todos los productos."""
        pass

    @abstractmethod
    def find_by_name(self, nombre: str) -> Producto:
        """Obtiene un producto por su nombre."""
        pass
    
    @abstractmethod
    def obtener_por_id(self, id: str) -> Producto:
        """Obtiene un producto por su ID."""
        pass
    
    @abstractmethod
    def actualizar(self, producto: Producto) -> Producto:
        """Actualiza un producto existente."""
        pass
    
    @abstractmethod
    def eliminar(self, id: str) -> bool:
        """Elimina un producto por su ID."""
        pass

    @abstractmethod
    def producto_en_uso(self, producto_id: str) -> bool:
        """Verifica si un producto está en uso."""
        pass

# Caso de uso para crear un producto
class CrearProductoUseCase:
    def __init__(self, repository: IProductoRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, producto_data: dict) -> Producto:
        """
        Ejecuta la creación de un producto, evitando duplicados por nombre.
        """
        nombre_producto = producto_data.get('nombre').upper()
        if self.repository.find_by_name(nombre_producto):
            raise ValueError(f"El producto con el nombre '{nombre_producto}' ya existe.")
            
        producto_data['nombre'] = nombre_producto
        producto_data['unidad_medida'] = producto_data.get('unidad_medida').upper()
        producto = Producto.from_dict(producto_data)
        nuevo_producto = self.repository.crear(producto)
        
        self.registrar_cambio_uc.execute(
            entidad_tipo='Producto',
            entidad_id=nuevo_producto.id,
            campo_modificado='Creación',
            valor_anterior='',
            valor_nuevo=f"Producto '{nuevo_producto.nombre}' creado"
        )
        
        return nuevo_producto

# Caso de uso para obtener todos los productos
class ObtenerProductosUseCase:
    def __init__(self, repository: IProductoRepository):
        self.repository = repository
        
    def execute(self, sort_by: str = 'nombre') -> list[Producto]:
        """Ejecuta la obtención de todos los productos."""
        return self.repository.obtener_todos(sort_by=sort_by)

# Caso de uso para obtener un producto por su ID
class ObtenerProductoPorIdUseCase:
    def __init__(self, repository: IProductoRepository):
        self.repository = repository
        
    def execute(self, id: str) -> Producto:
        """Ejecuta la obtención de un producto por su ID."""
        return self.repository.obtener_por_id(id)

# Caso de uso para actualizar un producto
class ActualizarProductoUseCase:
    def __init__(self, repository: IProductoRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, id: str, producto_data: dict) -> Producto:
        """Ejecuta la actualización de un producto."""
        producto_actual = self.repository.obtener_por_id(id)
        if not producto_actual:
            return None

        producto_data['nombre'] = producto_data.get('nombre').upper()
        producto_data['unidad_medida'] = producto_data.get('unidad_medida').upper()
        producto_actualizado = Producto.from_dict({**producto_data, "id": id})
        
        if producto_actual.nombre != producto_actualizado.nombre:
            self.registrar_cambio_uc.execute(
                entidad_tipo='Producto',
                entidad_id=id,
                campo_modificado='nombre',
                valor_anterior=producto_actual.nombre,
                valor_nuevo=producto_actualizado.nombre
            )
        if producto_actual.unidad_medida != producto_actualizado.unidad_medida:
            self.registrar_cambio_uc.execute(
                entidad_tipo='Producto',
                entidad_id=id,
                campo_modificado='unidad_medida',
                valor_anterior=producto_actual.unidad_medida,
                valor_nuevo=producto_actualizado.unidad_medida
            )

        return self.repository.actualizar(producto_actualizado)

# Caso de uso para eliminar un producto
class EliminarProductoUseCase:
    def __init__(self, repository: IProductoRepository, registrar_cambio_uc: RegistrarCambioUseCase):
        self.repository = repository
        self.registrar_cambio_uc = registrar_cambio_uc
        
    def execute(self, id: str) -> bool:
        """
        Ejecuta la eliminación de un producto, verificando dependencias.
        """
        producto = self.repository.obtener_por_id(id)
        if not producto:
            return False

        # Verificar si el producto está siendo utilizado en alguna receta
        if self.repository.producto_en_uso(id):
            raise ValueError("El producto no se puede eliminar porque está siendo utilizado en una o más recetas, inventarios o modelos.")

        # Registrar el cambio antes de eliminar
        self.registrar_cambio_uc.execute(
            entidad_tipo='Producto',
            entidad_id=id,
            campo_modificado='Eliminación',
            valor_anterior=f"Producto '{producto.nombre}' eliminado",
            valor_nuevo=''
        )
        
        return self.repository.eliminar(id)

# Caso de uso para exportar productos a Excel
class ExportProductosExcel:
    def __init__(self, repository: IProductoRepository):
        self.repository = repository

    def execute(self):
        productos = self.repository.obtener_todos()
        
        productos_data = [
            {
                "nombre": p.nombre,
                "unidad_medida": p.unidad_medida,
            }
            for p in productos
        ]
        
        df = pd.DataFrame(productos_data)
        
        output = io.BytesIO()
        with pd.ExcelWriter(output, engine='openpyxl') as writer:
            df.to_excel(writer, index=False, sheet_name='Productos')
        
        output.seek(0)
        return output

# Caso de uso para importar productos desde Excel
class ImportProductosExcel:
    def __init__(self, repository: IProductoRepository):
        self.repository = repository

    def execute(self, file):
        df = pd.read_excel(file)
        
        required_columns = ["nombre", "unidad_medida"]
        if not all(col in df.columns for col in required_columns):
            raise ValueError(f"El archivo Excel debe contener las columnas: {', '.join(required_columns)}")

        for _, row in df.iterrows():
            producto_data = {
                "nombre": row["nombre"],
                "unidad_medida": row["unidad_medida"]
            }
            
            if not self.repository.find_by_name(producto_data["nombre"]):
                producto = Producto.from_dict(producto_data)
                self.repository.crear(producto)

from abc import ABC, abstractmethod
from src.core.domain.venta import Venta
import pandas as pd
from datetime import datetime

# Interfaz abstracta para el repositorio de ventas
class IVentaRepository(ABC):
    @abstractmethod
    def crear(self, venta: Venta) -> Venta:
        """Crea una nueva venta."""
        pass
    
    @abstractmethod
    def obtener_todos(self) -> list[Venta]:
        """Obtiene todas las ventas."""
        pass
    
    @abstractmethod
    def obtener_por_id(self, id: str) -> Venta:
        """Obtiene una venta por su ID."""
        pass
    
    @abstractmethod
    def actualizar(self, venta: Venta) -> Venta:
        """Actualiza una venta existente."""
        pass
    
    @abstractmethod
    def eliminar(self, id: str) -> bool:
        """Elimina una venta por su ID."""
        pass

    @abstractmethod
    def eliminar_multiples(self, ids: list[str]) -> bool:
        """Elimina múltiples ventas por sus IDs."""
        pass

    @abstractmethod
    def crear_multiples(self, ventas: list[Venta]) -> list[Venta]:
        """Crea múltiples ventas."""
        pass

    @abstractmethod
    def find_by_date(self, fecha: datetime.date) -> list[Venta]:
        """Obtiene todas las ventas para una fecha específica."""
        pass

# Caso de uso para crear una venta
class CrearVentaUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self, venta_data: dict) -> Venta:
        """Ejecuta la creación de una venta."""
        venta = Venta.from_dict(venta_data)
        return self.repository.crear(venta)

# Caso de uso para obtener todas las ventas
class ObtenerVentasUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self) -> list[Venta]:
        """Ejecuta la obtención de todas las ventas."""
        return self.repository.obtener_todos()

# Caso de uso para obtener una venta por su ID
class ObtenerVentaPorIdUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self, id: str) -> Venta:
        """Ejecuta la obtención de una venta por su ID."""
        return self.repository.obtener_por_id(id)

# Caso de uso para actualizar una venta
class ActualizarVentaUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self, id: str, venta_data: dict) -> Venta:
        """Ejecuta la actualización de una venta."""
        venta = Venta.from_dict({**venta_data, "id": id})
        return self.repository.actualizar(venta)

# Caso de uso para eliminar una venta
class EliminarVentaUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self, id: str) -> bool:
        """Ejecuta la eliminación de una venta."""
        return self.repository.eliminar(id)

# Caso de uso para eliminar múltiples ventas
class EliminarVentasMultiplesUseCase:
    def __init__(self, repository: IVentaRepository):
        self.repository = repository
        
    def execute(self, ids: list[str]) -> bool:
        """Ejecuta la eliminación de múltiples ventas."""
        return self.repository.eliminar_multiples(ids)

from src.core.domain.receta import Receta

# Caso de uso para importar ventas desde un archivo
class ImportarVentasUseCase:
    def __init__(self, repository: IVentaRepository, receta_repository):
        """
        Inicializa el caso de uso con los repositorios necesarios.
        
        Args:
            repository (IVentaRepository): Repositorio para acceder a los datos de ventas.
            receta_repository: Repositorio para acceder a los datos de recetas.
        """
        self.repository = repository
        self.receta_repository = receta_repository
        
    def execute(self, file_stream, fecha=None):
        """
        Ejecuta la importación de ventas desde un archivo Excel.
        Si las recetas no existen, las crea automáticamente.
        """
        df = pd.read_excel(file_stream)
        
        if 'Nombre' not in df.columns or 'Cantidad' not in df.columns:
            raise ValueError("El archivo debe contener las columnas 'Nombre' y 'Cantidad'")
            
        nombres_recetas_archivo = set(df['Nombre'].unique())
        recetas_existentes = self.receta_repository.obtener_todos()
        nombres_recetas_db = {r.nombre for r in recetas_existentes}
        
        recetas_faltantes = nombres_recetas_archivo - nombres_recetas_db
        nuevas_recetas_creadas = []
        if recetas_faltantes:
            for nombre_receta in recetas_faltantes:
                nueva_receta = Receta(nombre=nombre_receta, activa=True)
                self.receta_repository.crear(nueva_receta)
                nuevas_recetas_creadas.append(nueva_receta)
            
        ventas_a_crear = []
        errores = []
        
        for idx, row in df.iterrows():
            nombre_receta = row['Nombre']
            cantidad = row['Cantidad']
            
            if not isinstance(cantidad, (int, float)) or cantidad <= 0:
                errores.append(f"Fila {idx+2}: Cantidad inválida ({cantidad}) para la receta '{nombre_receta}'")
                continue
            
            venta = Venta(
                receta_nombre=nombre_receta,
                cantidad=int(cantidad),
                fecha=fecha or datetime.now().strftime('%Y-%m-%d')
            )
            ventas_a_crear.append(venta)
            
        if errores:
            raise ValueError("\n".join(errores))
            
        ventas_creadas = self.repository.crear_multiples(ventas_a_crear)
        return ventas_creadas, nuevas_recetas_creadas

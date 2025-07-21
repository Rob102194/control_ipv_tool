from abc import ABC, abstractmethod
from src.core.domain.area import Area

# Interfaz abstracta para el repositorio de áreas
class IAreaRepository(ABC):
    @abstractmethod
    def crear(self, area: Area) -> Area:
        """Crea una nueva área."""
        pass
    
    @abstractmethod
    def find_all(self) -> list[Area]:
        """Obtiene todas las áreas."""
        pass

    @abstractmethod
    def find_by_name(self, nombre: str) -> Area:
        """Obtiene un área por su nombre."""
        pass
    
    @abstractmethod
    def obtener_por_id(self, id: str) -> Area:
        """Obtiene un área por su ID."""
        pass
    
    @abstractmethod
    def actualizar(self, area: Area) -> Area:
        """Actualiza un área existente."""
        pass
    
    @abstractmethod
    def eliminar(self, id: str) -> bool:
        """Elimina un área por su ID."""
        pass

# Caso de uso para crear un área
class CrearAreaUseCase:
    def __init__(self, repository: IAreaRepository):
        self.repository = repository
        
    def execute(self, area_data: dict) -> Area:
        """Ejecuta la creación de un área."""
        area_data['nombre'] = area_data.get('nombre').upper()
        area = Area.from_dict(area_data)
        return self.repository.crear(area)

# Caso de uso para obtener todas las áreas
class ObtenerAreasUseCase:
    def __init__(self, repository: IAreaRepository):
        self.repository = repository
        
    def execute(self) -> list[Area]:
        """Ejecuta la obtención de todas las áreas."""
        return self.repository.find_all()

# Caso de uso para obtener un área por su ID
class ObtenerAreaPorIdUseCase:
    def __init__(self, repository: IAreaRepository):
        self.repository = repository
        
    def execute(self, id: str) -> Area:
        """Ejecuta la obtención de un área por su ID."""
        return self.repository.obtener_por_id(id)

# Caso de uso para actualizar un área
class ActualizarAreaUseCase:
    def __init__(self, repository: IAreaRepository):
        self.repository = repository
        
    def execute(self, id: str, area_data: dict) -> Area:
        """Ejecuta la actualización de un área."""
        area_data['nombre'] = area_data.get('nombre').upper()
        area = Area.from_dict({**area_data, "id": id})
        return self.repository.actualizar(area)

# Caso de uso para eliminar un área
class EliminarAreaUseCase:
    def __init__(self, repository: IAreaRepository):
        self.repository = repository
        
    def execute(self, id: str) -> bool:
        """Ejecuta la eliminación de un área."""
        return self.repository.eliminar(id)

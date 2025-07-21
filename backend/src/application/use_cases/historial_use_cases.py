from abc import ABC, abstractmethod
from src.core.domain.historial_cambios import HistorialCambios

class IHistorialRepository(ABC):
    @abstractmethod
    def registrar_cambio(self, historial: HistorialCambios):
        pass

    @abstractmethod
    def obtener_historial_por_entidad(self, entidad_tipo: str) -> list[HistorialCambios]:
        pass

class RegistrarCambioUseCase:
    def __init__(self, repository: IHistorialRepository):
        self.repository = repository

    def execute(self, entidad_tipo: str, entidad_id: str, campo_modificado: str, valor_anterior: str, valor_nuevo: str):
        historial = HistorialCambios(
            entidad_tipo=entidad_tipo,
            entidad_id=entidad_id,
            campo_modificado=campo_modificado,
            valor_anterior=valor_anterior,
            valor_nuevo=valor_nuevo
        )
        self.repository.registrar_cambio(historial)

class ObtenerHistorialUseCase:
    def __init__(self, repository: IHistorialRepository):
        self.repository = repository

    def execute(self, entidad_tipo: str) -> list[HistorialCambios]:
        return self.repository.obtener_historial_por_entidad(entidad_tipo)

from src.infrastructure.db import models as db_models
from src.core.domain.historial_cambios import HistorialCambios

class SQLiteHistorialRepository:
    def __init__(self, db_session):
        self.db_session = db_session

    def registrar_cambio(self, historial: HistorialCambios):
        historial_db = db_models.HistorialCambios(
            entidad_tipo=historial.entidad_tipo,
            entidad_id=historial.entidad_id,
            campo_modificado=historial.campo_modificado,
            valor_anterior=historial.valor_anterior,
            valor_nuevo=historial.valor_nuevo
        )
        self.db_session.add(historial_db)
        self.db_session.commit()

    def obtener_historial_por_entidad(self, entidad_tipo: str) -> list[HistorialCambios]:
        historial_db = self.db_session.query(db_models.HistorialCambios).filter_by(
            entidad_tipo=entidad_tipo
        ).order_by(db_models.HistorialCambios.fecha_cambio.desc()).all()

        return [
            HistorialCambios(
                id=str(h.id),
                entidad_tipo=h.entidad_tipo,
                entidad_id=h.entidad_id,
                campo_modificado=h.campo_modificado,
                valor_anterior=h.valor_anterior,
                valor_nuevo=h.valor_nuevo,
                fecha_cambio=h.fecha_cambio
            ) for h in historial_db
        ]

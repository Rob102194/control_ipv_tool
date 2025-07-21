from datetime import datetime

class HistorialCambios:
    def __init__(self, entidad_tipo: str, entidad_id: str, campo_modificado: str, valor_anterior: str = None, valor_nuevo: str = None, id: str = None, fecha_cambio: datetime = None):
        self.id = id
        self.entidad_tipo = entidad_tipo
        self.entidad_id = entidad_id
        self.campo_modificado = campo_modificado
        self.valor_anterior = valor_anterior
        self.valor_nuevo = valor_nuevo
        self.fecha_cambio = fecha_cambio

    def to_dict(self):
        return {
            "id": self.id,
            "entidad_tipo": self.entidad_tipo,
            "entidad_id": self.entidad_id,
            "campo_modificado": self.campo_modificado,
            "valor_anterior": self.valor_anterior,
            "valor_nuevo": self.valor_nuevo,
            "fecha_cambio": self.fecha_cambio.isoformat() if self.fecha_cambio else None
        }

    @classmethod
    def from_dict(cls, data: dict):
        return cls(
            id=data.get("id"),
            entidad_tipo=data["entidad_tipo"],
            entidad_id=data["entidad_id"],
            campo_modificado=data["campo_modificado"],
            valor_anterior=data.get("valor_anterior"),
            valor_nuevo=data.get("valor_nuevo"),
            fecha_cambio=datetime.fromisoformat(data["fecha_cambio"]) if data.get("fecha_cambio") else None
        )

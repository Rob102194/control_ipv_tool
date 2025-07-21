class Venta:
    def __init__(self, receta_nombre: str, cantidad: int, fecha: str, id: str = None):
        self.id = id
        self.receta_nombre = receta_nombre
        self.cantidad = cantidad
        self.fecha = fecha
        
    def to_dict(self):
        return {
            "id": self.id,
            "receta_nombre": self.receta_nombre,
            "cantidad": self.cantidad,
            "fecha": self.fecha
        }
        
    @classmethod
    def from_dict(cls, data: dict):
        return cls(
            receta_nombre=data["receta_nombre"],
            cantidad=data["cantidad"],
            fecha=data.get("fecha"),
            id=data.get("id")
        )

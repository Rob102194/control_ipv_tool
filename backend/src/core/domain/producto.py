class Producto:
    def __init__(self, nombre: str, unidad_medida: str, id: str = None):
        self.id = id
        self.nombre = nombre
        self.unidad_medida = unidad_medida
        
    def to_dict(self):
        return {
            "id": self.id,
            "nombre": self.nombre,
            "unidad_medida": self.unidad_medida
        }
        
    @classmethod
    def from_dict(cls, data: dict):
        return cls(
            id=data.get("id"),
            nombre=data["nombre"],
            unidad_medida=data["unidad_medida"]
        )

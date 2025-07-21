class Area:
    def __init__(self, id: str, nombre: str, codigo: str = None):
        self.id = id
        self.nombre = nombre
        self.codigo = codigo

    def __hash__(self):
        return hash(self.id)

    def __eq__(self, other):
        if not isinstance(other, Area):
            return NotImplemented
        return self.id == other.id
        
    def to_dict(self):
        return {
            "id": self.id,
            "nombre": self.nombre,
            "codigo": self.codigo
        }
        
    @classmethod
    def from_dict(cls, data: dict):
        return cls(
            id=data.get("id"),
            nombre=data["nombre"],
            codigo=data.get("codigo")
        )

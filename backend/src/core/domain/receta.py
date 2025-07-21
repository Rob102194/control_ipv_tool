# Define el modelo de dominio para una Receta
class Receta:
    def __init__(self, nombre: str, activa: bool = True, id: str = None):
        """
        Inicializa una instancia de Receta.
        
        Args:
            nombre (str): El nombre de la receta.
            activa (bool): Indica si la receta está activa. Por defecto es True.
            id (str, optional): El identificador único de la receta. Se genera automáticamente.
        """
        self.id = id
        self.nombre = nombre
        self.activa = activa
        self.ingredientes = []  # Lista para almacenar objetos de tipo Ingrediente

    def to_dict(self):
        """Convierte la instancia de Receta a un diccionario."""
        return {
            "id": self.id,
            "nombre": self.nombre,
            "activa": self.activa,
            "ingredientes": [ing.to_dict() for ing in self.ingredientes]
        }

    @classmethod
    def from_dict(cls, data: dict):
        """
        Crea una instancia de Receta a partir de un diccionario.
        Los ingredientes se deben asignar por separado.
        """
        return cls(
            nombre=data["nombre"],
            activa=data.get("activa", True),
            id=data.get("id")
        )

# Define el modelo de dominio para un Ingrediente
class Ingrediente:
    def __init__(self, producto_id: str, area_id: str, cantidad: float, id: str = None, receta_id: str = None):
        """
        Inicializa una instancia de Ingrediente.
        
        Args:
            producto_id (str): El ID del producto utilizado.
            area_id (str): El ID del área de donde se descuenta el producto.
            cantidad (float): La cantidad de producto utilizada.
            id (str, optional): El identificador único del ingrediente. Se genera automáticamente.
            receta_id (str, optional): El ID de la receta a la que pertenece.
        """
        self.id = id
        self.producto_id = producto_id
        self.area_id = area_id
        self.cantidad = cantidad
        self.receta_id = receta_id

    def to_dict(self):
        """Convierte la instancia de Ingrediente a un diccionario."""
        return {
            "id": self.id,
            "producto_id": self.producto_id,
            "area_id": self.area_id,
            "cantidad": self.cantidad
        }

    @classmethod
    def from_dict(cls, data: dict):
        """Crea una instancia de Ingrediente a partir de un diccionario."""
        return cls(
            producto_id=data["producto_id"],
            area_id=data["area_id"],
            cantidad=data["cantidad"],
            id=data.get("id"),
            receta_id=data.get("receta_id")
        )

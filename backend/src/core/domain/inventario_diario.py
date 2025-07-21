from datetime import date

# Representa un registro del inventario diario para un producto en un área específica.
class InventarioDiario:
    def __init__(self, id: str, fecha: date, area_id: str, producto_id: str,
                 inicio: float = 0.0, entradas: float = 0.0, consumo: float = 0.0,
                 merma: float = 0.0, otras_salidas: float = 0.0, final_fisico: float = 0.0,
                 final_teorico: float = 0.0, diferencia: float = 0.0,
                 producto_nombre: str = None, area_nombre: str = None, comentario: str = None):
        self.id = id
        self.fecha = fecha
        self.area_id = area_id
        self.producto_id = producto_id
        self.inicio = inicio if inicio is not None else 0.0
        self.entradas = entradas if entradas is not None else 0.0
        self.consumo = consumo if consumo is not None else 0.0
        self.merma = merma if merma is not None else 0.0
        self.otras_salidas = otras_salidas if otras_salidas is not None else 0.0
        self.final_fisico = final_fisico if final_fisico is not None else 0.0
        self.final_teorico = final_teorico if final_teorico is not None else 0.0
        self.diferencia = diferencia if diferencia is not None else 0.0
        self.producto_nombre = producto_nombre
        self.area_nombre = area_nombre
        self.comentario = comentario

    # Calcula el inventario final teórico y la diferencia con el conteo físico.
    def calcular_diferencias(self):
        self.final_teorico = (self.inicio + self.entradas) - self.consumo - self.merma - self.otras_salidas
        self.diferencia = self.final_fisico - self.final_teorico
        return self

    # Convierte el objeto de dominio a un diccionario para la serialización.
    def to_dict(self):
        return {
            "id": self.id,
            "fecha": self.fecha.isoformat(),
            "area_id": self.area_id,
            "producto_id": self.producto_id,
            "inicio": self.inicio,
            "entradas": self.entradas,
            "consumo": self.consumo,
            "merma": self.merma,
            "otras_salidas": self.otras_salidas,
            "final_fisico": self.final_fisico,
            "final_teorico": self.final_teorico,
            "diferencia": self.diferencia,
            "producto_nombre": self.producto_nombre,
            "area_nombre": self.area_nombre,
            "comentario": self.comentario
        }

"""Fixtures de caracterización (Fase 0).

Levantan la app Flask real contra una BD SQLite temporal y la siembran con un
escenario determinista (IDs fijos) para poder capturar *goldens* estables que
la implementación Go deberá reproducir en la Fase 5.
"""
import os
import sys
import tempfile
from datetime import date

import pytest

HERE = os.path.dirname(os.path.abspath(__file__))
BACKEND = os.path.dirname(HERE)
sys.path.insert(0, BACKEND)

# --- Escenario determinista -------------------------------------------------

FECHA_PREV = "2026-09-04"
FECHA = "2026-09-05"

AREAS = [
    {"id": "area-cocina", "nombre": "COCINA", "codigo": "COC"},
    {"id": "area-bar", "nombre": "BAR", "codigo": "BAR"},
]

PRODUCTOS = [
    {"id": "prod-aceite", "nombre": "ACEITE", "unidad_medida": "L"},
    {"id": "prod-sal", "nombre": "SAL", "unidad_medida": "KG"},
    {"id": "prod-ron", "nombre": "RON", "unidad_medida": "L"},
    {"id": "prod-limon", "nombre": "LIMON", "unidad_medida": "U"},
]

# receta -> lista de (ingrediente_id, producto_id, area_id, cantidad)
RECETAS = [
    {"id": "rec-pasta", "nombre": "PASTA", "activa": True, "ingredientes": [
        ("ing-1", "prod-aceite", "area-cocina", 0.05),
        ("ing-2", "prod-sal", "area-cocina", 0.01),
    ]},
    {"id": "rec-mojito", "nombre": "MOJITO", "activa": True, "ingredientes": [
        ("ing-3", "prod-ron", "area-bar", 0.05),
        ("ing-4", "prod-limon", "area-bar", 1.0),
    ]},
]

# modelo_ipv: (id, area_id, producto_id, orden)
MODELO_IPV = [
    ("m-1", "area-cocina", "prod-aceite", 0),
    ("m-2", "area-cocina", "prod-sal", 1),
    ("m-3", "area-bar", "prod-ron", 0),
    ("m-4", "area-bar", "prod-limon", 1),
]

VENTAS = [
    {"id": "v-1", "receta_nombre": "PASTA", "cantidad": 10, "fecha": FECHA},
    {"id": "v-2", "receta_nombre": "MOJITO", "cantidad": 8, "fecha": FECHA},
]

# Cierre del día anterior: alimenta el `inicio` del día FECHA.
INV_PREV = [
    # (id, area_id, producto_id, final_fisico)
    ("inv-p1", "area-cocina", "prod-aceite", 5.0),
    ("inv-p2", "area-cocina", "prod-sal", 2.0),
    ("inv-p3", "area-bar", "prod-ron", 3.0),
    ("inv-p4", "area-bar", "prod-limon", 20.0),
]


def _seed(db):
    from src.infrastructure.db import models as m

    for a in AREAS:
        db.session.add(m.Area(**a))
    for p in PRODUCTOS:
        db.session.add(m.Producto(**p))
    for r in RECETAS:
        db.session.add(m.Receta(id=r["id"], nombre=r["nombre"], activa=r["activa"]))
        for ing_id, prod_id, area_id, cant in r["ingredientes"]:
            db.session.add(m.Ingrediente(
                id=ing_id, receta_id=r["id"], producto_id=prod_id,
                area_id=area_id, cantidad=cant,
            ))
    for mid, area_id, prod_id, orden in MODELO_IPV:
        db.session.add(m.ModeloIPV(id=mid, area_id=area_id, producto_id=prod_id, orden=orden))
    for v in VENTAS:
        db.session.add(m.Venta(
            id=v["id"], receta_nombre=v["receta_nombre"],
            cantidad=v["cantidad"], fecha=date.fromisoformat(v["fecha"]),
        ))
    for inv_id, area_id, prod_id, ff in INV_PREV:
        db.session.add(m.InventarioDiario(
            id=inv_id, fecha=date.fromisoformat(FECHA_PREV),
            area_id=area_id, producto_id=prod_id,
            inicio=0.0, entradas=0.0, consumo=0.0, merma=0.0, otras_salidas=0.0,
            final_fisico=ff, final_teorico=ff, diferencia=0.0,
        ))
    db.session.commit()


@pytest.fixture(scope="session")
def client():
    tmpdir = tempfile.mkdtemp(prefix="ipv-char-")
    os.environ["DB_URI"] = f"sqlite:///{os.path.join(tmpdir, 'test.db')}"

    from app import create_app
    from src.infrastructure.db.models import db

    app = create_app()
    app.config.update(TESTING=True)

    with app.app_context():
        _seed(db)
        with app.test_client() as c:
            yield c

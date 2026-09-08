"""Tests de caracterización de la versión Python (Fase 0).

Fijan el comportamiento observable de la lógica delicada del IPV y escriben
*goldens* en migration/goldens/ que el port Go deberá reproducir byte a byte
(salvo diferencias documentadas) en la Fase 5.

Ejecutar:  backend/.venv/bin/python -m pytest backend/tests -q
"""
import json
import os

from conftest import FECHA, FECHA_PREV  # noqa: F401

GOLDENS_DIR = os.path.join(
    os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))),
    "migration", "goldens",
)


def _write_golden(name, request_desc, response):
    os.makedirs(GOLDENS_DIR, exist_ok=True)
    body = response.get_json()
    doc = {
        "request": request_desc,
        "response": {"status": response.status_code, "body": body},
    }
    path = os.path.join(GOLDENS_DIR, f"{name}.json")
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(doc, fh, indent=2, ensure_ascii=False, sort_keys=True)
        fh.write("\n")
    return body


# --- CRUD simple ----------------------------------------------------------


def test_golden_listados(client):
    for name, url in [
        ("productos_list", "/api/productos/"),
        ("areas_list", "/api/areas/"),
        ("recetas_list", "/api/recetas/"),
        ("ventas_list", "/api/ventas/"),
    ]:
        resp = client.get(url)
        assert resp.status_code == 200
        _write_golden(name, {"method": "GET", "url": url}, resp)


# --- IPV: la secuencia importa (consumo -> estado -> guardar -> reporte) --


def test_golden_ipv_flujo(client):
    # 1) Consumo calculado a partir de ventas x ingredientes.
    url = f"/api/ipv/calcular-consumo?fecha={FECHA}"
    resp = client.get(url)
    assert resp.status_code == 200
    consumo = _write_golden("ipv_calcular_consumo", {"method": "GET", "url": url}, resp)
    assert consumo == {
        "prod-aceite|area-cocina": 0.5,   # 10 * 0.05
        "prod-sal|area-cocina": 0.1,      # 10 * 0.01
        "prod-ron|area-bar": 0.4,         # 8 * 0.05
        "prod-limon|area-bar": 8.0,       # 8 * 1
    }

    # 2) Estado del día antes de guardar: plantilla con `inicio` arrastrado
    #    del final_fisico del día anterior.
    url = f"/api/ipv/estado?fecha={FECHA}"
    resp = client.get(url)
    assert resp.status_code == 200
    estado = _write_golden("ipv_estado_plantilla", {"method": "GET", "url": url}, resp)
    inicio_por_producto = {
        item["producto_id"]: item["inicio"]
        for items in estado.values() for item in items
    }
    assert inicio_por_producto == {
        "prod-aceite": 5.0, "prod-sal": 2.0, "prod-ron": 3.0, "prod-limon": 20.0,
    }
    for items in estado.values():
        for item in items:
            assert item["entradas"] == 0
            assert item["consumo"] == 0
            assert item["final_teorico"] == 0

    # 3) Guardar: el backend recalcula final_teorico y diferencia.
    payload = [
        _row("s-1", "area-cocina", "prod-aceite", "COCINA", "ACEITE",
             inicio=5.0, entradas=2.0, consumo=0.5, merma=0.1, final_fisico=6.2),
        _row("s-2", "area-cocina", "prod-sal", "COCINA", "SAL",
             inicio=2.0, entradas=0.0, consumo=0.1, merma=0.0, final_fisico=1.85,
             comentario={"merma": "", "diferencia": "faltan 50g"}),
        _row("s-3", "area-bar", "prod-ron", "BAR", "RON",
             inicio=3.0, entradas=1.0, consumo=0.4, merma=0.0, final_fisico=3.6),
        _row("s-4", "area-bar", "prod-limon", "BAR", "LIMON",
             inicio=20.0, entradas=0.0, consumo=8.0, merma=2.0, final_fisico=10.0),
    ]
    url = "/api/ipv/guardar"
    resp = client.post(url, json=payload)
    assert resp.status_code == 201
    guardado = _write_golden(
        "ipv_guardar", {"method": "POST", "url": url, "body": payload}, resp,
    )
    by_id = {r["id"]: r for r in guardado}
    # aceite: teorico = (5 + 2) - 0.5 - 0.1 - 0 = 6.4 ; diff = 6.2 - 6.4 = -0.2
    assert round(by_id["s-1"]["final_teorico"], 6) == 6.4
    assert round(by_id["s-1"]["diferencia"], 6) == -0.2
    # limon: teorico = (20 + 0) - 8 - 2 - 0 = 10 ; diff = 0
    assert round(by_id["s-4"]["final_teorico"], 6) == 10.0
    assert round(by_id["s-4"]["diferencia"], 6) == 0.0

    # 4) Reporte del día ya guardado.
    url = f"/api/ipv/reporte?fecha={FECHA}"
    resp = client.get(url)
    assert resp.status_code == 200
    reporte = _write_golden("ipv_reporte", {"method": "GET", "url": url}, resp)
    assert reporte["fecha"] == FECHA
    assert set(reporte["areas"]) == {"COCINA", "BAR"}
    # El faltante de aceite (-0.2 L) aparece en el resumen de COCINA.
    faltantes_cocina = " ".join(reporte["resumen"]["COCINA"]["faltantes"])
    assert "ACEITE" in faltantes_cocina


def _row(rid, area_id, producto_id, area_nombre, producto_nombre, *,
         inicio, entradas, consumo, merma, final_fisico, otras_salidas=0.0,
         comentario=None):
    return {
        "id": rid,
        "fecha": FECHA,
        "area_id": area_id,
        "producto_id": producto_id,
        "area_nombre": area_nombre,
        "producto_nombre": producto_nombre,
        "inicio": inicio,
        "entradas": entradas,
        "consumo": consumo,
        "merma": merma,
        "otras_salidas": otras_salidas,
        "final_fisico": final_fisico,
        "comentario": json.dumps(comentario) if comentario is not None else "",
    }

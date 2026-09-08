"""Genera ficheros .xlsx de referencia con el MISMO código que la versión Python
(pandas + openpyxl), para que los tests del adaptador Go verifiquen paridad de
formato de importación/exportación (Fase 3).

Uso:  backend/.venv/bin/python backend/tests/make_excel_fixtures.py
Salida: migration/goldens/fixtures/*.xlsx
"""
import io
import os
import sys

import pandas as pd

HERE = os.path.dirname(os.path.abspath(__file__))
BACKEND = os.path.dirname(HERE)
ROOT = os.path.dirname(BACKEND)
OUT = os.path.join(ROOT, "migration", "goldens", "fixtures")
sys.path.insert(0, BACKEND)


def _write(df: pd.DataFrame, name: str, sheet: str) -> None:
    os.makedirs(OUT, exist_ok=True)
    path = os.path.join(OUT, name)
    buf = io.BytesIO()
    with pd.ExcelWriter(buf, engine="openpyxl") as w:
        df.to_excel(w, index=False, sheet_name=sheet)
    with open(path, "wb") as fh:
        fh.write(buf.getvalue())
    print("escrito", os.path.relpath(path, ROOT))


def main() -> None:
    # --- productos.xlsx: formato de ExportProductosExcel ---
    _write(
        pd.DataFrame([
            {"nombre": "ACEITE", "unidad_medida": "L"},
            {"nombre": "SAL", "unidad_medida": "KG"},
            {"nombre": "RON", "unidad_medida": "L"},
        ]),
        "productos.xlsx", "Productos",
    )

    # --- recetas.xlsx: formato de ExportRecetasExcel (una fila por ingrediente;
    #     receta sin ingredientes -> fila con campos vacíos) ---
    _write(
        pd.DataFrame([
            {"receta_nombre": "PASTA", "producto_nombre": "ACEITE",
             "unidad_medida": "L", "cantidad": 0.05, "area_nombre": "COCINA"},
            {"receta_nombre": "PASTA", "producto_nombre": "SAL",
             "unidad_medida": "KG", "cantidad": 0.01, "area_nombre": "COCINA"},
            {"receta_nombre": "MOJITO", "producto_nombre": "RON",
             "unidad_medida": "L", "cantidad": 0.05, "area_nombre": "BAR"},
            {"receta_nombre": "AGUA", "producto_nombre": "",
             "unidad_medida": "", "cantidad": "", "area_nombre": ""},
        ]),
        "recetas.xlsx", "Recetas",
    )

    # --- ventas_import.xlsx: lo que sube un usuario (columnas Nombre, Cantidad),
    #     con filas problemáticas para probar la validación ---
    _write(
        pd.DataFrame([
            {"Nombre": "PASTA", "Cantidad": 10},
            {"Nombre": "MOJITO", "Cantidad": 8},
            {"Nombre": "CAFE", "Cantidad": 0},        # inválida: <= 0
            {"Nombre": "TE", "Cantidad": "abc"},      # inválida: no numérica
            {"Nombre": "PASTA", "Cantidad": 2.5},     # válida (float)
        ]),
        "ventas_import.xlsx", "Sheet1",
    )


if __name__ == "__main__":
    main()

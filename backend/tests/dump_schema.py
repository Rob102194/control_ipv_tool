"""Regenera migration/schema_actual.sql a partir de models.py (Fase 0).

Uso:  backend/.venv/bin/python backend/tests/dump_schema.py
"""
import os
import sqlite3
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
BACKEND = os.path.dirname(HERE)
ROOT = os.path.dirname(BACKEND)
sys.path.insert(0, BACKEND)

HEADER = """\
-- Esquema real de la versión Python 0.1.0 (Fase 0).
-- Generado por backend/tests/dump_schema.py a partir de
-- backend/src/infrastructure/db/models.py.
--
-- Estado final tras las migraciones Alembic 7a5988a22b87 y b29f7ae8b140:
-- inventario_diario.comentario EXISTE, modelo_ipv.comentario NO.
--
-- Los DEFAULT de columnas (activa=1, floats=0.0) viven solo en la capa ORM.
-- En Go los fija la capa de dominio/repositorio explícitamente.

"""


def main() -> None:
    tmp = os.path.join(tempfile.gettempdir(), "ipv_schema_dump.db")
    if os.path.exists(tmp):
        os.remove(tmp)
    os.environ["DB_URI"] = f"sqlite:///{tmp}"

    import app as app_module  # noqa: F401  (create_app corre db.create_all)

    app_module.create_app()

    con = sqlite3.connect(tmp)
    rows = con.execute(
        "SELECT sql FROM sqlite_master "
        "WHERE sql IS NOT NULL AND name NOT LIKE 'sqlite_%' "
        "ORDER BY (type='index'), name"
    ).fetchall()
    con.close()

    out = os.path.join(ROOT, "migration", "schema_actual.sql")
    with open(out, "w", encoding="utf-8") as fh:
        fh.write(HEADER)
        for (stmt,) in rows:
            fh.write(stmt.strip() + ";\n\n")

    print(f"escrito {out} ({len(rows)} objetos)")


if __name__ == "__main__":
    main()

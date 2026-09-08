# Migración a Go + Wails — progreso

Plan completo: artefacto "Control IPV a Go + Wails". Rama: `feature/go-wails-migration`.

## Estado

| Fase | Estado | Notas |
|------|--------|-------|
| 0 — Red de seguridad | ✅ hecho | `openapi.yaml`, `migration/schema_actual.sql`, `migration/goldens/`, tests de caracterización en `backend/tests/` |
| 1 — Esqueleto Go | ✅ hecho | `cmd/server` arranca, migra con goose y responde `/healthz`; `go test ./...` verde |
| 2 — Core + puertos | ⏳ pendiente | |
| 3 — Adapters | ⏳ pendiente | |
| 4 — Casos de uso | ⏳ pendiente | |
| 5 — Capa HTTP | ⏳ pendiente | arnés de paridad contra `migration/goldens/` |
| 6 — Frontend | ⏳ pendiente | |
| 7 — Shell Wails | ⏳ pendiente | `cmd/desktop` aún no existe |
| 8 — Migración de datos | ⏳ pendiente | |
| 9 — Web / móvil | ⏳ pendiente | `cmd/server` ya es la base |

## Cómo trabajar

Requisitos: Go 1.26+, Python 3.14 (solo Fase 0), Node 24 (frontend), Wails v2 (Fase 7).

```sh
make help          # lista de objetivos
make dev           # arranca cmd/server en :5175 (CONTROL_IPV_ENV=dev)
make check         # go mod tidy + go vet + go test ./...
make goldens       # regenera migration/goldens/ desde la versión Python
make schema-dump   # regenera migration/schema_actual.sql
```

El venv de Python de la Fase 0 se crea con:
`python3 -m venv backend/.venv && backend/.venv/bin/pip install -r backend/requirements.txt pytest pyyaml`

## Configuración (variables de entorno)

| Variable | Default | Uso |
|----------|---------|-----|
| `CONTROL_IPV_ENV` | `dev` | `dev` (logs texto) / `prod` (logs JSON) |
| `CONTROL_IPV_ADDR` | `127.0.0.1:5175` | dirección de escucha HTTP |
| `CONTROL_IPV_LOG_LEVEL` | `info` | debug / info / warn / error |
| `CONTROL_IPV_DATA_DIR` | `os.UserConfigDir()/ControlIPV` | carpeta de datos del usuario |
| `CONTROL_IPV_DB_PATH` | `<DATA_DIR>/inventario.db` | fichero SQLite |
| `CONTROL_IPV_CORS_ORIGINS` | (vacío) | lista blanca CORS separada por comas (solo web) |

## Estructura añadida

```
cmd/server/            # entrega HTTP independiente (dev + web)
internal/platform/     # config, datadir, logging
internal/adapters/sqlite/
    db.go              # Open() con PRAGMA WAL/foreign_keys/busy_timeout
    migrate.go         # goose embebido
    migrations/00001_init.sql
internal/httpapi/       # router chi, middleware, errores tipados -> HTTP
migration/             # artefactos de paridad (Fase 0)
openapi.yaml           # contrato (Fase 0)
backend/tests/         # caracterización de la versión Python (Fase 0)
```

## Desviaciones respecto al plan

- El paquete de la capa HTTP es `internal/httpapi` (no `internal/http`) para no
  chocar con el `net/http` de la stdlib.
- Se usa `Makefile` en vez de `Taskfile.yml` (`task` no está instalado; `make` sí).

## Hallazgos que condicionan la Fase 5

- **Formato de números en JSON.** Los goldens conservan la representación de
  Python: `diferencia: -0.20000000000000018`, no `-0.2`. `encoding/json` de Go
  emite la forma corta. El arnés de paridad debe comparar números con tolerancia,
  o el DTO debe replicar el formato `repr` de Python. Ver `migration/goldens/ipv_guardar.json`.
- `/ipv/reporte` devuelve `notas` como **lista** (no como objeto por área, que es
  lo que arma el frontend actual). El contrato manda: lista.
- `guardar` y `estado` devuelven `producto_nombre` y `area_nombre`; el core no
  debe cargar con esos campos (van en el DTO de salida).

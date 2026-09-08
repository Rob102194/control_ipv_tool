# Migración a Go + Wails — progreso

Plan completo: artefacto "Control IPV a Go + Wails". Rama: `feature/go-wails-migration`.

## Estado

| Fase | Estado | Notas |
|------|--------|-------|
| 0 — Red de seguridad | ✅ hecho | `openapi.yaml`, `migration/schema_actual.sql`, `migration/goldens/`, tests de caracterización en `backend/tests/` |
| 1 — Esqueleto Go | ✅ hecho | `cmd/server` arranca, migra con goose y responde `/healthz`; `go test ./...` verde |
| 2 — Core + puertos | ✅ hecho | `internal/core/domain` (entidades puras, `CalcularDiferencias`, `CalcularConsumo`, tipo `Date`) sin deps externas; `internal/core/ports` (7 repos + `UnitOfWork`/`Repos` + `Clock`/`IDGen`); tests contra los goldens |
| 3 — Adapters | ✅ hecho | `internal/adapters/sqlite` (`Store` = `Repos` + `UnitOfWork`; 7 repos con SQL a mano vía `database/sql`, mapean fila→dominio; `withTx` para atomicidad) + `internal/adapters/excel` (excelize; parse/write productos, recetas, ventas). Tests de repo sobre BD temporal y de Excel contra fixtures generados por Python. `platform.SystemClock`/`UUIDGen`. |
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
internal/core/
    domain/            # entidades puras + reglas; CERO deps externas
        date.go        #   tipo Date (YYYY-MM-DD, sin zona horaria)
        inventario.go  #   InventarioDiario.CalcularDiferencias()
        consumo.go     #   CalcularConsumo(ventas, recetasPorNombre)
        errors.go      #   ValidationError / ConflictError
    ports/             # interfaces: repos, UnitOfWork/Repos, Clock, IDGen
internal/adapters/sqlite/
    db.go              # Open() con PRAGMA WAL/foreign_keys/busy_timeout
    migrate.go         # goose embebido
    migrations/00001_init.sql
    store.go           # Store: ports.Repos + ports.UnitOfWork; querier; withTx
    *_repo.go          # 7 repos, SQL a mano, fila -> dominio
    scan.go            # coerción de tipos, mapErr (UNIQUE -> ConflictError)
internal/adapters/excel/  # excelize: Parse/Write productos, recetas, ventas
internal/httpapi/       # router chi, middleware, errores tipados -> HTTP
migration/             # artefactos de paridad (Fase 0)
openapi.yaml           # contrato (Fase 0)
backend/tests/         # caracterización de la versión Python (Fase 0)
```

## Desviaciones respecto al plan

- El paquete de la capa HTTP es `internal/httpapi` (no `internal/http`) para no
  chocar con el `net/http` de la stdlib.
- Se usa `Makefile` en vez de `Taskfile.yml` (`task` no está instalado; `make` sí).
- **Fase 3: SQL a mano con `database/sql` en vez de sqlc.** El esquema es pequeño
  y congelado; hacerlo a mano evita el tooling de codegen (relevante para el
  build de Wails y CI) y da control total sobre la coerción de tipos dinámicos de
  SQLite (`FLOAT` puede volver como int64/float64/NULL). El objetivo del plan
  (SQL crudo, sin ORM, repos que devuelven dominio) se cumple igual.

## Decisiones de la Fase 3

- **`Store` es a la vez `ports.Repos` y `ports.UnitOfWork`.** Sin transacción
  para lecturas/escrituras sueltas; `Do(ctx, func(Repos) error)` agrupa varias en
  una transacción. Con `MaxOpenConns=1`, una transacción retiene la conexión —
  aceptable en escritorio monousuario.
- **`withTx`** hace que cada método de repo con varias sentencias (crear/actualizar
  receta, guardar modelo, guardar IPV) sea atómico aunque se llame suelto; si ya
  hay transacción (dentro de `Do`) la reutiliza.
- **IDs los asigna siempre el caso de uso** (`ports.IDGen`); los repos exigen
  `ID != ""` al insertar y nunca generan ids.
- **`GuardarTodos` del IPV** usa `INSERT … ON CONFLICT(fecha,area_id,producto_id)
  DO UPDATE` (upsert), sin tocar el `id` existente — como `save_all` de Python.
- **Borrado de recetas**: se borran los ingredientes primero (la FK no tiene
  `ON DELETE CASCADE`; en Python lo hacía el cascade del ORM).
- **`RecetaRepository.CrearMultiples`** inserta solo cabeceras, sin ingredientes
  (paridad con `crear_multiples`, usado al importar ventas).
- **Excel**: el adaptador solo convierte bytes ⇄ filas tipadas; resolver
  productos/áreas y persistir es del caso de uso (Fase 4). Fixtures de paridad en
  `migration/goldens/fixtures/` (`make excel-fixtures`).

## Decisiones de la Fase 2

- **`domain.Date`** (año/mes/día, sin hora ni zona) para las columnas DATE.
  Evita el drift de zona horaria de raíz. Trae `MarshalText`/`UnmarshalText`
  (única representación válida: `YYYY-MM-DD`), pero NO `Scan`/`Value` — la
  conversión a TEXT la hace el adaptador SQLite (Fase 3).
- **`InventarioDiario` no lleva `producto_nombre` ni `area_nombre`** (eran
  decoración de lectura en Python). Van en el DTO de salida (Fase 5).
- **`CalcularConsumo` es función pura** `(ventas, recetasPorNombre) -> map`. El
  caso de uso (Fase 4) arma el mapa de recetas; el dominio no toca repos. La
  clave serializada `producto|area` es cosa del DTO.
- **Contrato de repos**: `context.Context` en todo método; búsqueda por identidad
  devuelve `ports.ErrNoEncontrado`; entran y salen entidades de dominio.
- **`UnitOfWork.Do(ctx, func(Repos) error)`** para escrituras multi-paso;
  lecturas simples reciben el repo concreto.
- **`internal/core` no importa nada externo** (`go mod tidy` no añade deps).

## Hallazgos que condicionan la Fase 5

- **Formato de números en JSON.** Los goldens conservan la representación de
  Python: `diferencia: -0.20000000000000018`, no `-0.2`. Confirmado en Fase 2:
  `strconv.FormatFloat(x, 'g', -1, 64)` en Go produce **exactamente** la misma
  cadena que Python para estos casos (`internal/core/domain/inventario_test.go::
  TestCalcularDiferencias_ReprCoincideConPython`). Es decir: si el DTO serializa
  los floats con `'g',-1,64` la paridad de cuerpo JSON es exacta, sin tolerancia.
- `/ipv/reporte` devuelve `notas` como **lista** (no como objeto por área, que es
  lo que arma el frontend actual). El contrato manda: lista.
- `guardar` y `estado` devuelven `producto_nombre` y `area_nombre`; el core no
  debe cargar con esos campos (van en el DTO de salida).

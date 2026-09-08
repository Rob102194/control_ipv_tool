# Migración a Go + Wails — progreso

Plan completo: artefacto "Control IPV a Go + Wails". Rama: `feature/go-wails-migration`.

## Estado

| Fase | Estado | Notas |
|------|--------|-------|
| 0 — Red de seguridad | ✅ hecho | `openapi.yaml`, `migration/schema_actual.sql`, `migration/goldens/`, tests de caracterización en `backend/tests/` |
| 1 — Esqueleto Go | ✅ hecho | `cmd/server` arranca, migra con goose y responde `/healthz`; `go test ./...` verde |
| 2 — Core + puertos | ✅ hecho | `internal/core/domain` (entidades puras, `CalcularDiferencias`, `CalcularConsumo`, tipo `Date`) sin deps externas; `internal/core/ports` (7 repos + `UnitOfWork`/`Repos` + `Clock`/`IDGen`); tests contra los goldens |
| 3 — Adapters | ✅ hecho | `internal/adapters/sqlite` (`Store` = `Repos` + `UnitOfWork`; 7 repos con SQL a mano vía `database/sql`, mapean fila→dominio; `withTx` para atomicidad) + `internal/adapters/excel` (excelize; parse/write productos, recetas, ventas). Tests de repo sobre BD temporal y de Excel contra fixtures generados por Python. `platform.SystemClock`/`UUIDGen`. |
| 4 — Casos de uso | ✅ hecho | `internal/app/usecases` — solo importa `core/domain` y `core/ports`. Servicios por agregado (Producto/Area/Receta/Venta/IPV/Historial). Escrituras multi-paso + historial dentro de `UOW.Do`. `ObtenerEstado`/`GenerarReporte` con dominio; el reporte devuelve datos estructurados (el texto lo arma la Fase 5). Importaciones (productos/recetas/ventas) con la lógica de Python. Tests contra los goldens y fixtures. |
| 5 — Capa HTTP | ✅ hecho | `internal/httpapi` — DTOs (`dto.go`) con claves = `to_dict` de Python; 23 rutas montadas en `router.go` (con y sin barra final); handlers en `handlers.go`/`excel_handlers.go`; `validator` en los cuerpos struct; `statusFor` mapea `domain.ValidationError→422`, `ConflictError→409`, `NotFoundError→404`, `ErrNoEncontrado→404`. `cmd/server` cablea `Store`+`Services`. **Arnés de paridad** (`parity_test.go`): la API Go reproduce los 8 goldens (listados, consumo, estado, guardar, reporte) — comparación estructural con ids normalizados. |
| 6 — Frontend | ✅ hecho | `client.js` → `/api` relativo (+ proxy Vite a `:5175` en dev); `useIPV.handleCalcularDiferencias` ahora llama a `POST /ipv/calcular` (fin del recálculo duplicado en el cliente); `react-beautiful-dnd` → `@hello-pangea/dnd`; `vite build` → `web/dist/`, embebido en Go (`web/embed.go`, `//go:embed`) y servido con fallback SPA. `cmd/server` sirve el SPA embebido. |
| 7 — Shell Wails | ✅ hecho | `internal/appboot` (cableado compartido); `cmd/desktop/main.go` (tag `desktop`) usa `AssetServer.Handler = router` (SPA + /api por el mismo handler), `SingleInstanceLock`, `OnBeforeClose` oculta a bandeja, bandeja best-effort con `energye/systray` (Abrir/Salir). `cmd/desktop/wails.json`. Compila con `go build -tags desktop ./cmd/desktop`; `wails build` en la máquina destino. |
| 8 — Migración de datos | ✅ hecho | Primer arranque: `platform.ImportLegacyIfNeeded` copia la `inventario.db` de la versión Python al datadir con `VACUUM INTO` (+ `.pre-go.bak`, `.imported-from.txt`, sin tocar el origen); `sqlite.Migrate` detecta la BD preexistente y **sella 00001 como baseline** sin recrear tablas. **Probado contra la BD real de producción** (229 productos, 16 892 ventas, 24 994 filas de IPV, 163 fechas): todos los endpoints responden. El esquema `00001` ahora incluye `grupos` y `recetas.grupo_id` (existen en producción; el código Go los preserva aunque no los use). |
| 9 — Web / móvil | ✅ hecho | `cmd/server` y `cmd/desktop` comparten **el mismo** conjunto de `internal/` (verificado con `go list -deps`): la única diferencia es el envoltorio en `cmd/`. Añadido el *seam* `appboot.Options.AuthMiddleware` → `httpapi.Deps.AuthMiddleware` (envuelve `/api`; nil en escritorio). `docs/web-roadmap.md` detalla lo que falta (auth, `negocio_id`, adaptador Postgres, bloqueo optimista) — nada de ello toca `internal/core`. |

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
internal/app/usecases/  # servicios por agregado; solo depende de core/*
    services.go        # Deps, Services, New()
    ipv.go reporte.go importar.go views.go …
internal/httpapi/       # router, DTOs, handlers, errores tipados -> HTTP
    dto.go handlers.go excel_handlers.go router.go parity_test.go
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

## Decisiones de la Fase 8

- **Hallazgo**: la BD real de producción está por delante del código Python de
  este repo: tiene `alembic_version = 65b880ef501d` (migración ausente), una
  tabla `grupos` y `recetas.grupo_id`. El código Go NO usa esos campos pero el
  esquema `00001` los incluye (para converger fresh-install y migrado) y los
  repos no los tocan → los datos se preservan.
- **Importación por copia, no in-place**: `VACUUM INTO` produce un fichero limpio
  (WAL integrado) en el datadir; el original queda como respaldo natural y además
  se hace `<db>.pre-go.bak`. Búsqueda del origen: `$CONTROL_IPV_IMPORT_DB`, luego
  `./inventario.db`, `./backend/instance/inventario.db`, junto al ejecutable.
- **Baseline**: si la BD trae `productos` pero no `goose_db_version`, se crea la
  tabla de goose y se inserta `version_id=1` sin ejecutar el SQL de `00001`.
- Solo ocurre en el primer arranque (cuando el datadir aún no tiene BD);
  idempotente después.

## Decisiones de la Fase 7

- **`internal/appboot`**: cableado compartido (config → BD → migración → repos →
  casos de uso → router). `cmd/server` y `cmd/desktop` solo cambian el envoltorio.
- **`cmd/desktop` lleva la etiqueta de build `desktop`** (la que ponen `wails dev`
  y `wails build`). Así `go build ./...`, `go vet` y el CI normal NO arrastran el
  toolchain de Wails (CGO + WebKit). Comprobación de compilación aparte:
  `make desktop-compile`.
- **Wails sirve la app por `AssetServer.Handler`** = el mismo `chi` router
  (SPA embebido + `/api`). No hay bindings Go↔JS. Migrar a web = compilar
  `cmd/server`.
- `SingleInstanceLock` (crítico por la BD única). Cerrar la ventana la **oculta**
  (`OnBeforeClose` → `WindowHide`); se sale por "Salir" en la bandeja.
- **Bandeja best-effort** con `energye/systray` en una goroutine, con `recover`:
  si falla (típico en algunos SO), la app sigue — la ventana se oculta/reabre por
  el SO. Target real: Windows.
- `cmd/desktop/wails.json` con `frontend:dir: ../../frontend`; `wails build` se
  ejecuta desde `cmd/desktop/`. Iconos en `cmd/desktop/build/`.
- Deps nuevas: `wailsapp/wails/v2`, `energye/systray`.

## Decisiones de la Fase 6

- **`web/dist/` no se versiona** (solo `web/dist/.gitkeep` para que `//go:embed
  all:dist` compile); lo genera `make frontend` / CI antes de `make build`.
- El SPA se sirve desde `web.Handler()`: fichero si existe, si no `index.html`
  (fallback para el enrutador de React). Se monta como `r.Handle("/*", …)`
  después de `/api` y `/healthz`.
- En `npm run dev` Vite proxya `/api` → `http://127.0.0.1:5175` (el `cmd/server`).
  Se levantan los dos: `make dev` + `make dev-front`.
- Eliminado `frontend/public/icon.png:Zone.Identifier` (basura de Windows que
  rompía `//go:embed` por el `:` en el nombre).
- `eslint` del proyecto está roto de antes (mismatch eslint 8 / flat config); no
  se tocó. `vite build` sí pasa.

## Decisiones de la Fase 5

- **DTOs en el paquete `httpapi`** (no un subpaquete `dto/`): structs con tags
  `json` que replican el `to_dict` de Python; la serialización vive aquí.
- **`pyFloat`** para los textos del reporte: `strconv.FormatFloat('g',-1,64)` y,
  si el resultado no tiene `.`/`e`, se le añade `.0` (Python imprime `2.0`, Go
  `2`). Los NÚMEROS JSON no se tocan: el arnés compara de forma estructural
  (`2 == 2.0` como float64).
- **Rutas con y sin barra final** (`/api/productos` y `/api/productos/`) porque
  los blueprints Flask montaban con barra.
- **Códigos de estado afinados** respecto a Python (que casi siempre daba 400):
  `domain.ValidationError→422`, `ConflictError→409`, `NotFoundError→404`,
  `ports.ErrNoEncontrado→404`. El arnés de paridad tolera esta diferencia.
- **`validator` solo valida structs**; los cuerpos que son arrays
  (`/ipv/guardar`, `/ipv/calcular`) se validan al mapear a dominio.
- El **arnés de paridad** siembra el escenario con ids fijos directamente por el
  `Store` (sin pasar por los casos de uso) para poder comparar; normaliza los
  campos `"id"` antes del `reflect.DeepEqual` porque la plantilla de `/ipv/estado`
  genera ids nuevos.
- Pendiente (cosmético): el mensaje de error de `validator` sale crudo; se puede
  traducir a algo legible en una pasada posterior.

## Decisiones de la Fase 4

- **Servicios por agregado**, no un struct por caso de uso (más idiomático en Go,
  igual de explícito): `ProductoService`, `AreaService`, `RecetaService`,
  `VentaService`, `IPVService`, `HistorialService`, construidos por `usecases.New(Deps)`.
- `internal/app/usecases` (sin tests) **solo importa `core/domain` y `core/ports`**
  (verificado con `go list`).
- Historial + escritura principal van en la misma `UOW.Do` (transaccional).
- **`GenerarReporte` devuelve datos estructurados** (`ReporteIPV` con `[]ReporteDelta`,
  `[]ReporteNota`); el formateo `"ACEITE: 0.2 L"` / `"SAL (diferencia): …"` lo hará
  la Fase 5. Las notas conservan el orden de las claves del JSON del comentario
  (parser ordenado con `json.Decoder`), igual que Python.
- **Importaciones**: `ProductoService.Importar` / `RecetaService.Importar` /
  `VentaService.Importar` reproducen la lógica de Python, incluido que la
  importación NO normaliza nombres a MAYÚSCULAS y NO registra historial (quirk de
  `ImportProductosExcel`/`ImportRecetasExcel`). El error de ventas inválidas
  devuelve `*domain.ValidationError` con el texto `"Fila N: Cantidad inválida (raw)…"`
  unido por saltos de línea (Python devolvía 400; Go dará 422).
- **Tests de casos de uso sobre el `Store` de SQLite en fichero temporal**, no con
  fakes: es una implementación en memoria legítima de los puertos, con más
  fidelidad y menos código que reimplementar la lógica en un fake. Reloj e IDs sí
  son dobles deterministas.

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

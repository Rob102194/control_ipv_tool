# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Qué es esto

Control IPV: herramienta de control de inventario/IPV (Índice de Producción y
Ventas). Backend en Go, frontend en React, empaquetado como app de escritorio
(Wails v2) y desplegable como servidor web con el mismo binario/código.

Es una **migración desde una versión Python/Flask** ya eliminada del árbol
(sigue en el historial git, antes del commit "Se elimina la versión Python").
La migración se hizo en 9 fases, todas completas — el detalle fase por fase,
las decisiones tomadas y por qué, y los "hallazgos" que condicionaron el
diseño están en `MIGRATION.md`. Léelo antes de tocar formatos de
serialización, rutas HTTP o el esquema de BD: casi todas las rarezas del
código (rutas con y sin barra final, formato de floats en JSON, columnas
`grupos`/`grupo_id` sin usar, códigos de estado distintos entre Python y Go,
etc.) están explicadas ahí, no son descuido.

## Comandos

```sh
make help            # lista de objetivos del Makefile
make dev             # arranca cmd/server en :5175 (backend HTTP solo)
make dev-front       # arranca Vite con proxy /api -> :5175 (usar junto a `make dev`)
make check           # go mod tidy + go vet + go test ./...
make test            # go test ./...
make vet             # go vet ./...
make frontend        # compila el SPA (frontend/) en web/dist/ para que Go lo embeba
make build           # frontend + binario del servidor en bin/
make desktop-compile # comprueba que cmd/desktop compila (sin empaquetar)
make wails-dev       # app de escritorio en modo desarrollo (requiere Wails CLI)
make wails-build     # ejecutable de escritorio en cmd/desktop/build/bin
make wails-doctor    # verifica el entorno de Wails
```

Un test concreto: `go test ./internal/app/usecases/ -run TestIPV_GuardarYReporte`.

Requisitos: Go 1.26+, Node 24 (frontend). Para escritorio: Wails v2 CLI
(`go install github.com/wailsapp/wails/v2/cmd/wails@latest`); en Linux hace
falta `libgtk-3-dev libwebkit2gtk-4.1-dev` (CGO). `cmd/desktop` no tiene build
tag propio, así que `go build ./...`/`go test ./...` lo compilan también.

Frontend (`frontend/`): `npm run dev`, `npm run build`, `npm run lint`
(`eslint` está roto de antes por un mismatch eslint 8 / flat config — no es
un problema que introduzcas tú; `vite build` sí funciona).

CI (`.github/workflows/go.yml`): `go vet` + `go test -race ./...` + `go build
./...` en Ubuntu y Windows, más `go mod tidy` sin diff. Los releases
(`.github/workflows/release.yml`) compilan con Wails para las tres
plataformas al empujar un tag `v*`.

## Arquitectura

Hexagonal (puertos y adaptadores), de fuera hacia dentro:

```
cmd/server, cmd/desktop   → solo el envoltorio (HTTP suelto vs. ventana Wails)
internal/appboot          → cablea todo: config → datadir → BD → migraciones →
                             repos → casos de uso → router HTTP
internal/httpapi          → DTOs + rutas chi + handlers + mapeo error→código HTTP
internal/app/usecases     → lógica de negocio por agregado, transaccional
internal/core/ports       → interfaces (contratos) que necesita usecases
internal/core/domain      → entidades puras, cero dependencias externas
internal/adapters/sqlite  → implementación de ports.Repos/UnitOfWork sobre SQLite
internal/adapters/excel   → import/export de Excel (excelize)
internal/platform         → config, datadir, logging (transversal, sin negocio)
web/                       → //go:embed del SPA compilado (web/dist/)
frontend/                  → SPA React, consume /api
```

**Regla de dependencia que hay que respetar**: `internal/core` (domain +
ports) no importa nada fuera de sí mismo — ni frameworks, ni `database/sql`.
`internal/app/usecases` solo importa `core/domain` y `core/ports` (está
verificado a mano con `go list -deps`, no hay test automático). Si una
implementación necesita algo del exterior, se declara como interfaz en
`core/ports` y se implementa en `internal/adapters/*`.

**`cmd/server` y `cmd/desktop` comparten exactamente los mismos paquetes
`internal/`.** La única diferencia entre escritorio y web es el envoltorio en
`cmd/`: `appboot.New` construye el mismo `http.Handler` (router chi con
`/api` + `/healthz` + SPA embebido); `cmd/desktop` lo pasa como
`AssetServer.Handler` de Wails (sin bindings Go↔JS), `cmd/server` lo sirve con
`net/http`. Migrar algo a "web" nunca implica bifurcar `internal/`; el punto
de extensión para multiusuario/auth es `appboot.Options.AuthMiddleware` (hoy
`nil` en escritorio) — ver `docs/web-roadmap.md` para lo que falta (auth,
`negocio_id`, adaptador Postgres, bloqueo optimista).

### Contrato de repositorios (`internal/core/ports`)

- Todo método recibe `context.Context` como primer parámetro.
- Las búsquedas por identidad devuelven `ports.ErrNoEncontrado` si no hay fila
  (los casos de uso lo detectan con `errors.Is`).
- Reciben y devuelven **entidades de dominio**, nunca filas ni tipos de BD.
- Los ids los asigna siempre el caso de uso (`ports.IDGen`); los repos exigen
  `ID != ""` al insertar y nunca generan ids ellos mismos.

`ports.UnitOfWork.Do(ctx, func(Repos) error)` envuelve escrituras multi-paso
(crear receta + ingredientes, importar ventas, guardar la hoja de IPV) en una
transacción: commit si `fn` devuelve `nil`, rollback si devuelve error o
entra en pánico. Lecturas simples reciben el repositorio concreto sin pasar
por `Do`. `internal/adapters/sqlite.Store` implementa ambos (`ports.Repos` +
`ports.UnitOfWork`); su `withTx` hace que cada método de repo con varias
sentencias SQL sea atómico incluso llamado suelto, reutilizando la
transacción si ya hay una en curso.

### Errores de dominio → HTTP

`internal/core/domain/errors.go` define `ValidationError`, `ConflictError`,
`NotFoundError` (más el centinela `ports.ErrNoEncontrado`). `internal/httpapi`
los mapea: `ValidationError→422`, `ConflictError→409`,
`NotFoundError`/`ErrNoEncontrado→404`. Al añadir una regla de negocio nueva,
usa los constructores `domain.Invalid`/`domain.Conflictf`/`domain.NotFoundf`
en vez de errores planos, para que la capa HTTP siga sabiendo qué código
devolver.

### Migraciones de BD

`internal/adapters/sqlite/migrate.go` usa goose embebido
(`internal/adapters/sqlite/migrations/*.sql`). Es idempotente y tiene un caso
especial: si detecta una BD preexistente de la versión Python (tablas de
negocio pero sin la tabla de goose), sella `00001` como ya aplicada en vez de
recrear el esquema (`baselineIfLegacy`) — así conviven instalaciones nuevas y
BDs migradas desde Python. `platform.ImportLegacyIfNeeded` es lo que copia esa
BD legada al datadir en el primer arranque.

### Paridad con la versión Python

`migration/goldens/` + `migration/schema_actual.sql` + `openapi.yaml` son
artefactos **congelados** capturados desde la versión Python 0.1.0; no se
regeneran (el generador vivía en el `backend/` ya eliminado, recuperable del
historial git si hiciera falta). `internal/httpapi/parity_test.go` monta la
API Go sobre el mismo escenario determinista, la ejecuta y compara
estructuralmente el JSON contra esos goldens (normalizando los campos `id`).
Si tocas DTOs, formato de números (los floats se serializan con
`strconv.FormatFloat(x, 'g', -1, 64)` para igualar la representación de
Python) o rutas, corre este test — es la red de seguridad de la migración.

### Frontend

React + Vite, sin TypeScript. `frontend/src/api/*Api.js` son clientes finos
sobre `frontend/src/api/client.js` (axios, `baseURL` relativa `/api`, así
sirve igual en escritorio y en web). En dev, `vite.config.js` proxea `/api` a
`127.0.0.1:5175`; hace falta tener `make dev` y `make dev-front` corriendo a
la vez. `make frontend` compila a `web/dist/`, que Go embebe con
`//go:embed all:dist` (`web/embed.go`); por eso `web/dist/.gitkeep` está
versionado aunque el contenido del build no.

# Camino a web y móvil

Estado tras la Fase 9: la separación **funciona**. `cmd/server` (web) y
`cmd/desktop` (Wails) comparten **exactamente** el mismo conjunto de paquetes
`internal/` (verificado con `go list -deps`). Lo único que cambia entre
escritorio y web es el envoltorio en `cmd/`:

| | Escritorio | Web |
|---|---|---|
| Binario | `cmd/desktop` (Wails, tag `desktop`) | `cmd/server` |
| Ventana | webview nativo (`AssetServer.Handler`) | navegador → HTTP |
| Frontend | SPA embebido (`web/embed.go`) | SPA embebido **o** CDN estático |
| BD | SQLite en el datadir del usuario | SQLite en disco **o** (futuro) Postgres |
| Auth | ninguna (monousuario) | `Deps.AuthMiddleware` (hoy nil) |

El frontend ya llama a `/api` **relativo**, así que sirve igual en los dos
sitios sin cambios. Para móvil: el SPA es responsive → PWA, o envolverlo con
Capacitor apuntando `VITE_API_BASE_URL` al host de la API.

## Lo que falta para un despliegue web multiusuario

Ninguno de estos puntos toca el core (`internal/core`); son adaptadores o
middleware nuevos detrás de los puertos que ya existen.

### 1. Autenticación
- Implementar un `func(http.Handler) http.Handler` y pasarlo como
  `appboot.Options.AuthMiddleware` desde `cmd/server`.
- Valida el token/sesión y mete el `usuario`/`negocio` en el `context.Context`.
- El escritorio lo deja en `nil`.

### 2. Multi-tenant (`negocio_id`)
- Añadir `negocio_id` a las tablas raíz (`productos`, `areas`, `recetas`,
  `ventas`, `inventario_diario`, `modelo_ipv`, `historial_cambios`) en una
  migración `00002`.
- Los repositorios SQLite filtran por el `negocio_id` del contexto. La firma de
  los puertos no cambia si el `negocio_id` viaja en `ctx`; si se prefiere
  explícito, se añade un parámetro y se ajustan las 7 interfaces + adaptador.
- El core y los casos de uso no se enteran.

### 3. Adaptador Postgres
- Nuevo paquete `internal/adapters/postgres` que implemente `ports.Repos` y
  `ports.UnitOfWork` (mismo contrato que `sqlite.Store`).
- `appboot` elige el adaptador según config (`CONTROL_IPV_DB_DRIVER`).
- Migraciones goose con dialecto `postgres` (las de SQLite no sirven tal cual).
- `google/uuid` ya se usa para ids; encaja con `uuid` de Postgres.

### 4. Concurrencia / bloqueo optimista
- Hoy `IPVService.Guardar` hace *upsert* «último que escribe gana».
- Para varios usuarios sobre la misma hoja: añadir `version`/`updated_at` a
  `inventario_diario` y devolver 409 si cambió por debajo.
- Es un cambio en `InventarioDiarioRepository.GuardarTodos` + el caso de uso;
  el DTO gana un campo `version`.

### 5. CORS y cabeceras
- `CONTROL_IPV_CORS_ORIGINS` ya existe (lista blanca, se aplica solo si no está
  vacía). Revisar `Vary`, credenciales y `max-age` al montar el dominio real.

### 6. Operación
- `cmd/server` ya tiene apagado *graceful* y `/healthz` con versión de esquema.
- Falta: métricas, `/readyz`, límites de tamaño de cuerpo por ruta, rate-limit.

## Deuda conocida que conviene saldar antes de web

- El mensaje de error de `validator` sale crudo (Fase 5). Traducirlo.
- `grupos` / `recetas.grupo_id` existen en la BD pero el código no los expone
  (Fase 8). Si la versión web los necesita, añadirlos al dominio `Receta`,
  a los repos y a los DTO.

# Distribución de la app de escritorio

Control IPV es multiplataforma: Windows, macOS y Linux, con el mismo código. Cada
instalación es independiente y offline; por defecto los datos viven en una BD
SQLite propia en el directorio de configuración del usuario:

| SO | Ruta de datos |
|----|---------------|
| Windows | `%AppData%\ControlIPV\inventario.db` |
| macOS | `~/Library/Application Support/ControlIPV/inventario.db` |
| Linux | `~/.config/ControlIPV/inventario.db` |

## Modo portable (varios negocios con la misma app)

El `Control IPV.exe`/`Control IPV` "portable" que se publica junto al
instalador en cada release (sin instalador de por medio) detecta en tiempo de
arranque si puede escribir en su propia carpeta. Si puede, usa
`<carpeta del ejecutable>/data/inventario.db` en vez de AppData/Config — cada
copia que el usuario haga del ejecutable (una carpeta por negocio/restaurante)
tiene así su propia base de datos, sin variables de entorno ni scripts.

Esto se implementa en `internal/platform/portable.go`
(`EnsurePortableDataDir`), cableado desde `cmd/desktop/main.go` **antes** de
`appboot.New`, y solo si `CONTROL_IPV_DATA_DIR` no está fijada a mano. Nunca
se activa para el `.exe` instalado por el instalador NSIS (vive en
`Program Files`, de solo lectura para un usuario estándar) ni dentro de un
`.app` de macOS — ahí sigue usando AppData/Config como siempre, porque:

- escribir en `Program Files` sin ser administrador falla o Windows lo
  redirige en silencio a una carpeta oculta por usuario (File System
  Virtualization), y
- el desinstalador de Windows borra `Program Files\...\Control IPV` entero
  (`RMDir /r`) al desinstalar o reinstalar una actualización — si la BD
  viviera ahí, se perdería.

La detección prueba permiso de escritura real (crea y borra un fichero), no
solo mira la ruta, así que es segura por construcción incluso si algún día
cambia dónde vive el instalador.

Para separar negocios con la app **instalada** (no la portable), sigue
haciendo falta `CONTROL_IPV_DATA_DIR` por instancia (ver `docs/web-roadmap.md`
para el multi-tenant real a futuro).

## Varias instancias en paralelo (multi-negocio)

Cada ventana de escritorio es un proceso independiente con su propia BD. El
candado de instancia única de Wails (`SingleInstanceLock`, en
`cmd/desktop/main.go`) identifica el proceso con un hash corto de la ruta de
la BD en vez de un id fijo: así, dos negocios con datadirs distintos (uno
portable y otro no, o dos con `CONTROL_IPV_DATA_DIR` diferentes) pueden tener
cada uno su ventana abierta a la vez. Abrir el **mismo** negocio dos veces
sigue bloqueado como siempre (el segundo lanzamiento trae al frente la
ventana existente, no abre una nueva).

Para distinguir de un vistazo qué ventana corresponde a qué negocio cuando
hay varias abiertas, el nombre del negocio (editable en `/configuracion`) se
usa como título de la ventana ("Control IPV — <negocio>") y como etiqueta en
la barra de navegación.

## Copias de seguridad (backups)

Desde `/configuracion` el usuario puede crear backups manuales de la BD
("Backup ahora"), descargarlos, y restaurar uno (reemplaza los datos
actuales; la app pide reiniciarse tras restaurar porque el proceso cierra la
conexión a la BD a propósito en vez de intentar un hot-swap). Mecánica:

- `VACUUM INTO` a un `.tmp` + rename atómico (mismo patrón que la
  importación de la BD legada de Python).
- Se conservan como máximo 7 backups; al crear uno nuevo se borran
  automáticamente los más viejos (`internal/appboot/backup.go`,
  `backupsARetener`).
- La carpeta de destino es configurable (por defecto, `<datadir>/backups`);
  puede apuntarse a una carpeta ya sincronizada por Dropbox/Drive/similar
  para tener una copia fuera de la máquina. Se guarda en la tabla
  `configuracion` (columna `backup_dir`), no es una variable de entorno.
- En escritorio, la carpeta se puede elegir con el diálogo nativo del SO
  (`wruntime.OpenDirectoryDialog`, cableado solo en `cmd/desktop`) en vez de
  escribirla a mano; en el despliegue web ese selector no existe (un
  navegador no puede elegir una carpeta del disco del servidor) y el campo
  de texto sigue siendo la única forma de fijarla.

## Compilar

Wails **no** cross-compila de forma fiable: cada plataforma se compila en su
propio SO. El workflow `.github/workflows/release.yml` lo hace por ti en un tag
`vX.Y.Z` (matriz Windows + macOS + Linux) y adjunta los tres a un GitHub Release.

A mano, en cada SO, desde `cmd/desktop/`:

```sh
# Windows  -> Control IPV.exe + Control IPV-amd64-installer.exe
wails build -clean -nsis -webview2 download

# macOS    -> Control IPV.app  (recomendado empaquetar en .zip con `ditto -c -k --keepParent`)
wails build -clean -platform darwin/universal

# Linux    -> ./Control IPV  (binario; empaquetar en .tar.gz)
wails build -clean
```

Antes de tag: subir `info.productVersion` en `cmd/desktop/wails.json`.

## Compartir

- **Directo**: enviar el instalador/zip por Drive, correo o carpeta compartida.
- **GitHub Releases**: el workflow los sube solo; se comparte el enlace del asset
  (sirve con el repo privado).

## Requisitos en la máquina del usuario

| SO | Runtime |
|----|---------|
| Windows 10 20H2+/11 | WebView2 (ya incluido; `-webview2 download` lo instala si falta) |
| macOS 11+ | nada (WebKit del sistema) |
| Linux | `libgtk-3-0` y `libwebkit2gtk-4.1-0` (`sudo apt install ...` / equivalente) |

## Avisos de seguridad (sin firma de código)

- **Windows**: SmartScreen → *Más información → Ejecutar de todas formas*.
  Para quitarlo: certificado Authenticode.
- **macOS**: Gatekeeper bloquea el `.app` sin notarizar. El usuario:
  clic derecho → *Abrir*, o `xattr -dr com.apple.quarantine "Control IPV.app"`.
  Para quitarlo: cuenta Apple Developer + notarización.
- **Linux**: sin aviso; dar permiso de ejecución (`chmod +x`).

## Actualizaciones

No hay auto-update. Se envía el nuevo instalador/zip y el usuario lo ejecuta
encima. **La BD del `%AppData%`/`~/.config` no se toca**: los datos se conservan.

## Migrar datos de la versión Python

Al primer arranque, si el datadir aún no tiene BD, se busca una `inventario.db`
de la versión anterior en: `$CONTROL_IPV_IMPORT_DB`, el directorio actual,
`./backend/instance/`, y junto al ejecutable. Si la encuentra, la copia al
datadir (`VACUUM INTO`) y deja `inventario.db.pre-go.bak`. El origen no se toca.

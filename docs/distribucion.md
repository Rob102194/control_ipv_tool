# Distribución de la app de escritorio

Control IPV es multiplataforma: Windows, macOS y Linux, con el mismo código. Cada
instalación es independiente y offline; los datos viven en una BD SQLite propia
en el directorio de configuración del usuario:

| SO | Ruta de datos |
|----|---------------|
| Windows | `%AppData%\ControlIPV\inventario.db` |
| macOS | `~/Library/Application Support/ControlIPV/inventario.db` |
| Linux | `~/.config/ControlIPV/inventario.db` |

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

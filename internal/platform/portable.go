package platform

import (
	"os"
	"path/filepath"
	"strings"
)

// EnsurePortableDataDir decide si procede usar <carpeta del ejecutable>/data
// como directorio de datos, en vez del AppData/Config del usuario (ver
// EnsureDataDir en paths.go).
//
// Solo se activa cuando el usuario puede escribir en la carpeta del
// ejecutable: el caso real es el binario "portable" que se publica junto al
// instalador en cada release (ver docs/distribucion.md) — si el usuario lo
// copia a una carpeta propia (Escritorio, una carpeta por negocio, etc.),
// cada copia escribe su propia base de datos junto a sí misma, sin variables
// de entorno ni scripts (.bat) para separar negocios.
//
// Nunca se activa cuando el ejecutable vive en una ubicación de solo lectura
// para el usuario estándar — el caso típico es el instalador de Windows, que
// pone el .exe en Program Files: ahí escribir falla o Windows lo redirige en
// silencio a una carpeta oculta por usuario, y el desinstalador borra ese
// directorio entero al desinstalar/actualizar (se perdería la BD). Tampoco se
// activa dentro de un .app de macOS, por el mismo motivo (se perdería al
// reinstalar, y escribir ahí rompe la integridad del bundle firmado).
//
// Devuelve ("", false) si no corresponde modo portable: el llamador debe
// seguir con el comportamiento por defecto (EnsureDataDir con override vacío).
func EnsurePortableDataDir() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	// Resuelve symlinks (p. ej. un acceso directo de Linux o un enlace
	// simbólico) para operar sobre la ubicación real del binario.
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return portableDataDirFor(exe)
}

func portableDataDirFor(exePath string) (string, bool) {
	if isInsideMacAppBundle(exePath) {
		return "", false
	}
	dir := filepath.Join(filepath.Dir(exePath), "data")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false
	}
	if !DirIsWritable(dir) {
		return "", false
	}
	return dir, true
}

// isInsideMacAppBundle detecta si la ruta cae dentro de Contents/MacOS de un
// .app, el patrón de cualquier app empaquetada con Wails en macOS.
func isInsideMacAppBundle(p string) bool {
	sep := string(filepath.Separator)
	return strings.Contains(p, ".app"+sep+"Contents"+sep+"MacOS"+sep)
}

// DirIsWritable comprueba permiso de escritura real (no solo los bits de
// permisos, que en Windows no reflejan fielmente ACLs/Program Files):
// intenta crear y borrar un fichero temporal dentro de dir. La usa también
// backup.go al resolver un directorio de backups configurado a mano.
func DirIsWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

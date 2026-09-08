package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// appDirName es el nombre de la carpeta de la aplicación dentro del directorio
// de configuración del usuario del sistema operativo.
const appDirName = "ControlIPV"

// EnsureDataDir resuelve y crea (si no existe) el directorio de datos del usuario.
//
// Si override no está vacío se usa tal cual. En caso contrario se usa
// os.UserConfigDir()/ControlIPV, que resuelve a:
//   - macOS:   ~/Library/Application Support/ControlIPV
//   - Windows: %AppData%\ControlIPV
//   - Linux:   ~/.config/ControlIPV
func EnsureDataDir(override string) (string, error) {
	dir := override
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolviendo directorio de configuración del usuario: %w", err)
		}
		dir = filepath.Join(base, appDirName)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creando directorio de datos %q: %w", dir, err)
	}
	return dir, nil
}

// ResolveDBPath devuelve la ruta del fichero SQLite: dbPathOverride si se indicó,
// o <dataDir>/inventario.db en caso contrario.
func ResolveDBPath(dataDir, dbPathOverride string) string {
	if dbPathOverride != "" {
		return dbPathOverride
	}
	return filepath.Join(dataDir, "inventario.db")
}

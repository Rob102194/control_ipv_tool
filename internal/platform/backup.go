package platform

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// BackupInfo describe un backup de la base de datos ya creado en disco.
type BackupInfo struct {
	Nombre      string
	CreadoEn    time.Time
	TamanoBytes int64
}

const backupTimestampLayout = "20060102_150405"

// backupNombreRe es el único patrón de nombre que CrearBackup genera y que
// RutaBackup/ListarBackups aceptan. Sirve de defensa en profundidad contra
// path traversal: un nombre que no case con esto (p. ej. "../../etc/passwd")
// se rechaza antes de tocar el filesystem.
var backupNombreRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+\.db$`)

// ResolverDirectorioBackups decide el directorio efectivo de backups: el
// override del usuario (configuración del negocio) si no está vacío, o
// <dirBD>/backups por defecto. En ambos casos lo crea si no existe y
// comprueba que se pueda escribir de verdad — para fallar con un mensaje
// claro en el momento de usarlo (crear/listar/restaurar), no a mitad de un
// backup ya empezado.
func ResolverDirectorioBackups(dirBD, override string) (string, error) {
	dir := strings.TrimSpace(override)
	if dir == "" {
		dir = filepath.Join(dirBD, "backups")
	} else {
		dir = filepath.Clean(dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creando el directorio de backups %q: %w", dir, err)
	}
	if !DirIsWritable(dir) {
		return "", fmt.Errorf("no se puede escribir en el directorio de backups %q", dir)
	}
	return dir, nil
}

// CrearBackup saca una copia limpia de la BD a un fichero nuevo y timestamped
// dentro de destDir, usando VACUUM INTO (ver legacydb.go: válido incluso en
// modo WAL con escrituras recientes sin checkpoint, produce un único fichero
// sin los -wal/-shm sueltos). Escribe primero a un nombre temporal y renombra
// al final (atómico): un backup a medio escribir nunca aparece listado como
// si estuviera completo.
func CrearBackup(ctx context.Context, db *sql.DB, dbPath, destDir string) (BackupInfo, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return BackupInfo{}, fmt.Errorf("creando el directorio de backups %q: %w", destDir, err)
	}
	prefijo := strings.TrimSuffix(filepath.Base(dbPath), filepath.Ext(dbPath))
	if prefijo == "" {
		prefijo = "inventario"
	}
	ahora := time.Now()
	nombre := fmt.Sprintf("%s_%s.db", prefijo, ahora.Format(backupTimestampLayout))
	destino := filepath.Join(destDir, nombre)
	tmp := destino + ".tmp"
	_ = os.Remove(tmp)

	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", tmp); err != nil {
		_ = os.Remove(tmp)
		return BackupInfo{}, fmt.Errorf("generando el backup: %w", err)
	}
	if err := os.Rename(tmp, destino); err != nil {
		_ = os.Remove(tmp)
		return BackupInfo{}, fmt.Errorf("finalizando el backup: %w", err)
	}
	fi, err := os.Stat(destino)
	if err != nil {
		return BackupInfo{}, fmt.Errorf("leyendo el backup recién creado: %w", err)
	}
	return BackupInfo{Nombre: nombre, CreadoEn: ahora, TamanoBytes: fi.Size()}, nil
}

// ListarBackups lista los backups válidos de destDir, más reciente primero.
// Un directorio que aún no existe se trata como "sin backups todavía", no
// como error.
func ListarBackups(destDir string) ([]BackupInfo, error) {
	entradas, err := os.ReadDir(destDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []BackupInfo
	for _, e := range entradas {
		if e.IsDir() || !backupNombreRe.MatchString(e.Name()) {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, BackupInfo{Nombre: e.Name(), CreadoEn: fi.ModTime(), TamanoBytes: fi.Size()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreadoEn.After(out[j].CreadoEn) })
	return out, nil
}

// AplicarRetencion borra los backups más viejos de destDir, conservando como
// mucho `mantener` (los más recientes). No devuelve error: una falla aislada
// borrando un fichero viejo no debe tumbar el "crear backup" que la provocó.
func AplicarRetencion(destDir string, mantener int) {
	backups, err := ListarBackups(destDir)
	if err != nil || len(backups) <= mantener {
		return
	}
	for _, b := range backups[mantener:] {
		_ = os.Remove(filepath.Join(destDir, b.Nombre))
	}
}

// RutaBackup valida `nombre` (debe caer exactamente en el patrón que
// CrearBackup genera, sin separadores de ruta) y devuelve su ruta completa
// dentro de destDir, comprobando que el fichero exista.
func RutaBackup(destDir, nombre string) (string, error) {
	limpio := filepath.Base(nombre)
	if limpio != nombre || !backupNombreRe.MatchString(limpio) {
		return "", fmt.Errorf("nombre de backup inválido: %q", nombre)
	}
	ruta := filepath.Join(destDir, limpio)
	if _, err := os.Stat(ruta); err != nil {
		return "", err
	}
	return ruta, nil
}

// RestaurarArchivo reemplaza dbPath por una copia de backupPath. El llamador
// debe haber cerrado ya cualquier *sql.DB abierto sobre dbPath: esta función
// solo mueve archivos, no gestiona conexiones. Borra también -wal/-shm
// residuales de la BD anterior para que no se repliquen contra el fichero
// restaurado en el próximo arranque (backupPath, al venir de CrearBackup, ya
// es un fichero limpio sin WAL pendiente).
func RestaurarArchivo(dbPath, backupPath string) error {
	for _, sufijo := range []string{"", "-wal", "-shm"} {
		_ = os.Remove(dbPath + sufijo)
	}

	origen, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("abriendo el backup: %w", err)
	}
	defer origen.Close()

	destino, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("creando %q: %w", dbPath, err)
	}
	if _, err := io.Copy(destino, origen); err != nil {
		_ = destino.Close()
		return fmt.Errorf("copiando el backup: %w", err)
	}
	return destino.Close()
}

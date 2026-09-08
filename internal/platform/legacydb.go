package platform

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// ImportLegacyIfNeeded copia una BD de la versión Python al destino (dbPath) la
// primera vez que se arranca la versión Go, si aún no hay BD en el destino.
//
// No modifica ni borra el fichero de origen (es su propia copia de seguridad).
// Además deja un `<dbPath>.pre-go.bak` y un `<dbPath>.imported-from.txt`.
//
// Orden de búsqueda del origen:
//  1. $CONTROL_IPV_IMPORT_DB (ruta explícita)
//  2. ./inventario.db  (CWD; es donde la versión Python ponía sqlite:///inventario.db)
//  3. ./backend/instance/inventario.db
//  4. inventario.db junto al ejecutable
func ImportLegacyIfNeeded(dbPath string, logger *slog.Logger) error {
	if fileExists(dbPath) {
		return nil
	}

	src := findLegacyDB()
	if src == "" {
		return nil // instalación nueva: no hay nada que importar
	}
	if same, _ := sameFile(src, dbPath); same {
		return nil
	}

	logger.Info("importando base de datos de la versión anterior", "origen", src, "destino", dbPath)

	if err := vacuumInto(src, dbPath); err != nil {
		return fmt.Errorf("copiando %q -> %q: %w", src, dbPath, err)
	}
	if err := copyFile(dbPath, dbPath+".pre-go.bak"); err != nil {
		logger.Warn("no se pudo crear la copia .pre-go.bak", "err", err)
	}
	_ = os.WriteFile(dbPath+".imported-from.txt", []byte(src+"\n"), 0o644)

	logger.Info("base de datos importada", "destino", dbPath)
	return nil
}

func findLegacyDB() string {
	var cands []string
	if p := os.Getenv("CONTROL_IPV_IMPORT_DB"); p != "" {
		cands = append(cands, p)
	}
	cands = append(cands, "inventario.db", filepath.Join("backend", "instance", "inventario.db"))
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), "inventario.db"))
	}
	for _, c := range cands {
		if fileExists(c) {
			abs, err := filepath.Abs(c)
			if err != nil {
				return c
			}
			return abs
		}
	}
	return ""
}

// vacuumInto usa `VACUUM INTO` para producir una copia limpia de la BD de
// origen (integra el WAL, sin ficheros -wal/-shm sueltos).
func vacuumInto(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", "file:"+src+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec("VACUUM INTO ?", dst); err != nil {
		return err
	}
	return nil
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

func sameFile(a, b string) (bool, error) {
	fa, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	fb, err := os.Stat(b)
	if err != nil {
		return false, nil
	}
	return os.SameFile(fa, fb), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

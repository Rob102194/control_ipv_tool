package platform

import (
	"database/sql"
	"errors"
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

	logger.Info("importando base de datos de la versión anterior", "origen", src, "destino", dbPath)

	importada, err := vacuumInto(src, dbPath)
	if err != nil {
		return fmt.Errorf("copiando %q -> %q: %w", src, dbPath, err)
	}
	if !importada {
		// Otro proceso ganó la carrera y ya importó la BD (p. ej. escritorio y
		// servidor arrancando a la vez contra el mismo datadir): nada que hacer.
		logger.Info("la base de datos ya fue importada por otro proceso", "destino", dbPath)
		return nil
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
//
// Escribe primero a un fichero temporal exclusivo de este proceso y lo
// publica en dst con os.Link, que falla atómicamente si dst ya existe. Esto
// cierra la ventana de tiempo entre comprobar que dst no existe (en
// ImportLegacyIfNeeded) y crearlo: si dos procesos arrancan a la vez contra
// el mismo datadir, solo uno "gana" la importación y el otro recibe
// imported=false en vez de un error crudo de sqlite.
func vacuumInto(src, dst string) (imported bool, err error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, err
	}

	tmp := fmt.Sprintf("%s.importing-%d.tmp", dst, os.Getpid())
	_ = os.Remove(tmp) // por si quedó de un intento anterior con el mismo PID
	defer os.Remove(tmp)

	db, err := sql.Open("sqlite", "file:"+src+"?mode=ro")
	if err != nil {
		return false, err
	}
	defer db.Close()
	if _, err := db.Exec("VACUUM INTO ?", tmp); err != nil {
		return false, err
	}

	if err := os.Link(tmp, dst); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
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

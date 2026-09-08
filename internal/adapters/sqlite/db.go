// Package sqlite es el adaptador de persistencia sobre SQLite.
//
// Usa el driver pure-Go modernc.org/sqlite (sin CGO) para que el binario de
// escritorio (Wails) se compile sin toolchain de C. Las implementaciones de
// repositorio de este paquete devuelven objetos de dominio, nunca filas ni
// tipos generados: la frontera hexagonal se cierra aquí.
package sqlite

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

// Open abre (creando el fichero si no existe) la base de datos SQLite en path y
// devuelve un *sql.DB listo para usar, con los PRAGMA de la aplicación aplicados
// a cada conexión:
//
//   - journal_mode=WAL   lecturas concurrentes con un escritor
//   - foreign_keys=ON    integridad referencial (SQLite la desactiva por defecto)
//   - busy_timeout=5000  reintenta 5s ante "database is locked" en vez de fallar
//
// El pool se limita a una conexión: en un escritorio monousuario evita de raíz
// los bloqueos de escritura y simplifica el razonamiento sobre transacciones.
func Open(path string) (*sql.DB, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("ruta de base de datos vacía")
	}

	dsn := "file:" + path + "?" + url.Values{
		"_pragma": {
			"journal_mode(WAL)",
			"foreign_keys(ON)",
			"busy_timeout(5000)",
		},
	}.Encode()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("abriendo base de datos %q: %w", path, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("verificando conexión con %q: %w", path, err)
	}

	return db, nil
}

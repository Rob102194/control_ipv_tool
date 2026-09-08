package sqlite

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// datetimeLayout es el formato en que SQLite (y SQLAlchemy) guardan las columnas
// DATETIME: "YYYY-MM-DD HH:MM:SS", en UTC y sin zona.
const datetimeLayout = "2006-01-02 15:04:05"

// mapErr traduce errores del driver a errores de puerto/dominio.
//   - sql.ErrNoRows              -> ports.ErrNoEncontrado
//   - violación de UNIQUE        -> *domain.ConflictError
//
// El resto se devuelve tal cual (la capa HTTP lo tratará como 500).
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ports.ErrNoEncontrado
	}
	if isUniqueViolation(err) {
		return &domain.ConflictError{Msg: "ya existe un registro con ese valor único"}
	}
	return err
}

// isUniqueViolation detecta el error de restricción UNIQUE de modernc.org/sqlite
// sin acoplarse a su tipo concreto: su mensaje contiene "UNIQUE constraint failed".
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// --- conversión de valores ------------------------------------------------

// dbDate serializa una domain.Date al texto "YYYY-MM-DD" que espera la columna.
func dbDate(d domain.Date) string { return d.String() }

// parseDate interpreta el texto de una columna DATE.
func parseDate(s string) (domain.Date, error) {
	if s == "" {
		return domain.Date{}, nil
	}
	// SQLAlchemy puede haber guardado "2026-09-05" o, en algún caso, con hora.
	if len(s) > 10 {
		s = s[:10]
	}
	return domain.ParseDate(s)
}

// dbDateTime serializa un time.Time al texto DATETIME de SQLite (UTC).
func dbDateTime(t time.Time) string { return t.UTC().Format(datetimeLayout) }

// parseDateTime interpreta el texto de una columna DATETIME como UTC.
func parseDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(datetimeLayout, s); err == nil {
		return t.UTC(), nil
	}
	// Tolera el formato ISO con "T" por si acaso.
	return time.Parse(time.RFC3339, s)
}

// nullString convierte "" en NULL al escribir (la versión Python guardaba None).
func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// strFromNull devuelve "" cuando la columna es NULL.
func strFromNull(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// Migrate aplica todas las migraciones pendientes embebidas en el binario.
// Es idempotente: si la BD ya está al día no hace nada. Si detecta una BD
// preexistente de la versión Python (tiene las tablas pero no la de goose),
// sella la migración inicial como aplicada en vez de intentar recrearla.
func Migrate(db *sql.DB, logger *slog.Logger) error {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(gooseSlog{logger})

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("configurando dialecto goose: %w", err)
	}

	if err := baselineIfLegacy(db, logger); err != nil {
		return fmt.Errorf("sellando BD preexistente: %w", err)
	}

	before, _ := goose.GetDBVersion(db)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("aplicando migraciones: %w", err)
	}

	after, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("leyendo versión de esquema: %w", err)
	}

	if before == after {
		logger.Info("esquema al día", "version", after)
	} else {
		logger.Info("esquema migrado", "desde", before, "hasta", after)
	}
	return nil
}

// SchemaVersion devuelve la versión de migración actual de la BD.
func SchemaVersion(db *sql.DB) (int64, error) {
	return goose.GetDBVersion(db)
}

// baselineIfLegacy sella la migración 00001 como aplicada cuando la BD ya trae
// el esquema (viene de la versión Python) pero aún no la tabla de control de
// goose. Así no se intenta recrear tablas que ya existen y no se pierde nada.
func baselineIfLegacy(db *sql.DB, logger *slog.Logger) error {
	if tableExists(db, "goose_db_version") {
		return nil // ya la gestiona goose
	}
	if !tableExists(db, "productos") {
		return nil // BD nueva: goose.Up ejecutará 00001 con normalidad
	}
	logger.Info("BD preexistente detectada (esquema Python); sellando 00001 como baseline")
	if _, err := goose.EnsureDBVersion(db); err != nil {
		return err
	}
	_, err := db.Exec(
		`INSERT INTO goose_db_version (version_id, is_applied, tstamp) VALUES (1, 1, CURRENT_TIMESTAMP)`,
	)
	return err
}

func tableExists(db *sql.DB, name string) bool {
	var n string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name = ?`, name,
	).Scan(&n)
	return err == nil
}

// gooseSlog adapta el logger de goose a slog.
type gooseSlog struct{ l *slog.Logger }

func (g gooseSlog) Printf(format string, v ...any) {
	g.l.Info("goose: " + fmt.Sprintf(format, v...))
}

// Fatalf registra el error pero no termina el proceso: el error devuelto por
// goose.Up es quien decide el flujo en el llamador.
func (g gooseSlog) Fatalf(format string, v ...any) {
	g.l.Error("goose: " + fmt.Sprintf(format, v...))
}

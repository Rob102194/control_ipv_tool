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
// Es idempotente: si la BD ya está al día no hace nada.
func Migrate(db *sql.DB, logger *slog.Logger) error {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(gooseSlog{logger})

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("configurando dialecto goose: %w", err)
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

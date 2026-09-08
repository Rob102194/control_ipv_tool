package sqlite

import (
	"io"
	"log/slog"
	"path/filepath"
	"sort"
	"testing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestMigrateCreatesSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Migrate(db, newTestLogger()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Idempotencia: una segunda pasada no debe fallar.
	if err := Migrate(db, newTestLogger()); err != nil {
		t.Fatalf("Migrate (2ª vez): %v", err)
	}

	v, err := SchemaVersion(db)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v < 1 {
		t.Fatalf("versión de esquema = %d, se esperaba >= 1", v)
	}

	rows, err := db.Query(
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name <> 'goose_db_version'`,
	)
	if err != nil {
		t.Fatalf("consultando tablas: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	sort.Strings(got)

	want := []string{
		"areas", "historial_cambios", "ingredientes", "inventario_diario",
		"modelo_ipv", "movimientos", "productos", "recetas", "ventas",
	}
	if len(got) != len(want) {
		t.Fatalf("tablas = %v, se esperaba %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tablas = %v, se esperaba %v", got, want)
		}
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fk.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, se esperaba 1", fk)
	}
}

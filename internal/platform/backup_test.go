package platform

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// abrirDBDePrueba crea una BD SQLite mínima con una fila, usando el mismo
// driver que legacydb.go (blank-import ya hecho en ese archivo, misma
// package). Evita depender de internal/adapters/sqlite para no acoplar este
// paquete a una capa superior.
func abrirDBDePrueba(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE t(x INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO t VALUES (42)"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCrearBackupYListar(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "inventario.db")
	destDir := filepath.Join(dir, "backups")
	db := abrirDBDePrueba(t, dbPath)

	info, err := CrearBackup(context.Background(), db, dbPath, destDir)
	if err != nil {
		t.Fatalf("CrearBackup: %v", err)
	}
	if info.TamanoBytes == 0 {
		t.Error("el backup quedó con tamaño 0")
	}
	if !backupNombreRe.MatchString(info.Nombre) {
		t.Errorf("nombre de backup inesperado: %q", info.Nombre)
	}

	backups, err := ListarBackups(destDir)
	if err != nil {
		t.Fatalf("ListarBackups: %v", err)
	}
	if len(backups) != 1 || backups[0].Nombre != info.Nombre {
		t.Fatalf("ListarBackups = %+v, se esperaba 1 con nombre %q", backups, info.Nombre)
	}

	// El contenido del backup debe ser una copia real, legible por separado.
	db2, err := sql.Open("sqlite", filepath.Join(destDir, info.Nombre))
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	var x int
	if err := db2.QueryRow("SELECT x FROM t").Scan(&x); err != nil {
		t.Fatalf("leyendo el backup: %v", err)
	}
	if x != 42 {
		t.Errorf("x = %d, se esperaba 42", x)
	}
}

func TestListarBackups_DirectorioInexistente(t *testing.T) {
	backups, err := ListarBackups(filepath.Join(t.TempDir(), "no-existe"))
	if err != nil {
		t.Fatalf("se esperaba nil, no error: %v", err)
	}
	if backups != nil {
		t.Errorf("backups = %+v, se esperaba nil", backups)
	}
}

func TestAplicarRetencion(t *testing.T) {
	dir := t.TempDir()
	// 10 ficheros con nombres válidos, mtimes escalonados.
	base := time.Now().Add(-10 * time.Hour)
	var nombres []string
	for i := 0; i < 10; i++ {
		nombre := filepath.Join(dir, "inventario_2026010"+string(rune('0'+i))+"_000000.db")
		if err := os.WriteFile(nombre, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		mtime := base.Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(nombre, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		nombres = append(nombres, nombre)
	}

	AplicarRetencion(dir, 7)

	backups, err := ListarBackups(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 7 {
		t.Fatalf("quedaron %d backups, se esperaban 7", len(backups))
	}
	// Deben sobrevivir los 7 más RECIENTES (índices 3..9, los últimos creados).
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(nombres[i]); err == nil {
			t.Errorf("el backup más viejo %q debería haberse borrado", nombres[i])
		}
	}
}

func TestRutaBackup_RechazaTraversal(t *testing.T) {
	dir := t.TempDir()
	casos := []string{
		"../../etc/passwd",
		"../secreto.db",
		"sub/evil.db",
		"sinextension",
		"nombre con espacios y .db raro.txt",
	}
	for _, c := range casos {
		if _, err := RutaBackup(dir, c); err == nil {
			t.Errorf("RutaBackup(%q) debería haber fallado", c)
		}
	}
}

func TestRutaBackup_NoExistente(t *testing.T) {
	dir := t.TempDir()
	if _, err := RutaBackup(dir, "inventario_20260101_000000.db"); err == nil {
		t.Error("se esperaba error: el fichero no existe")
	}
}

func TestRestaurarArchivo(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "inventario.db")
	backupDir := filepath.Join(dir, "backups")

	// BD "actual" con un valor, y su backup con OTRO valor (simula que el
	// backup es de un estado anterior).
	dbActual := abrirDBDePrueba(t, dbPath)
	if _, err := dbActual.Exec("UPDATE t SET x = 99"); err != nil {
		t.Fatal(err)
	}
	// Deja basura -wal/-shm como la dejaría WAL real.
	_ = os.WriteFile(dbPath+"-wal", []byte("wal"), 0o644)
	_ = os.WriteFile(dbPath+"-shm", []byte("shm"), 0o644)

	backupInfo, err := CrearBackup(context.Background(), dbActual, dbPath, backupDir)
	if err != nil {
		t.Fatal(err)
	}
	// Ahora cambiamos la BD "actual" para que sea claramente distinta del backup.
	if _, err := dbActual.Exec("UPDATE t SET x = 1000"); err != nil {
		t.Fatal(err)
	}
	if err := dbActual.Close(); err != nil {
		t.Fatal(err)
	}

	backupPath := filepath.Join(backupDir, backupInfo.Nombre)
	if err := RestaurarArchivo(dbPath, backupPath); err != nil {
		t.Fatalf("RestaurarArchivo: %v", err)
	}

	// Los -wal/-shm viejos deben haber desaparecido.
	if _, err := os.Stat(dbPath + "-wal"); err == nil {
		t.Error("el -wal viejo debería haberse borrado")
	}
	if _, err := os.Stat(dbPath + "-shm"); err == nil {
		t.Error("el -shm viejo debería haberse borrado")
	}

	dbRestaurada, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbRestaurada.Close()
	var x int
	if err := dbRestaurada.QueryRow("SELECT x FROM t").Scan(&x); err != nil {
		t.Fatal(err)
	}
	if x != 99 {
		t.Errorf("x = %d tras restaurar, se esperaba 99 (el valor del backup, no 1000)", x)
	}
}

func TestResolverDirectorioBackups_PorDefecto(t *testing.T) {
	dirBD := t.TempDir()
	dir, err := ResolverDirectorioBackups(dirBD, "")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dirBD, "backups")
	if dir != want {
		t.Errorf("dir = %q, se esperaba %q", dir, want)
	}
}

func TestResolverDirectorioBackups_Override(t *testing.T) {
	override := filepath.Join(t.TempDir(), "mi-carpeta-de-drive")
	dir, err := ResolverDirectorioBackups(t.TempDir(), override)
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Clean(override) {
		t.Errorf("dir = %q, se esperaba %q", dir, override)
	}
}

func TestResolverDirectorioBackups_NoEscribible(t *testing.T) {
	// Mismo truco que portable_test.go: un segmento de la ruta es un
	// ARCHIVO, no un directorio, así que MkdirAll falla igual en cualquier SO.
	base := t.TempDir()
	bloqueador := filepath.Join(base, "bloqueador")
	if err := os.WriteFile(bloqueador, []byte("soy un archivo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolverDirectorioBackups(base, filepath.Join(bloqueador, "backups")); err == nil {
		t.Error("se esperaba error")
	}
}

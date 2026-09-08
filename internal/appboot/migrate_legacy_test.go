package appboot_test

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/Rob102194/control_ipv_tool/internal/appboot"
	"github.com/Rob102194/control_ipv_tool/internal/platform"
)

// legacySchema es un subconjunto realista del esquema de producción de la
// versión Python: SIN tabla de goose, CON grupos y recetas.grupo_id.
const legacySchema = `
CREATE TABLE productos (id VARCHAR(36) PRIMARY KEY, nombre VARCHAR(100) UNIQUE NOT NULL, unidad_medida VARCHAR(10) NOT NULL);
CREATE TABLE areas (id VARCHAR(36) PRIMARY KEY, nombre VARCHAR(50) UNIQUE NOT NULL, codigo VARCHAR(10));
CREATE TABLE grupos (id VARCHAR(36) PRIMARY KEY, nombre VARCHAR(50) UNIQUE NOT NULL);
CREATE TABLE recetas (id VARCHAR(36) PRIMARY KEY, nombre VARCHAR(100) UNIQUE NOT NULL, activa BOOLEAN, grupo_id VARCHAR(36));
CREATE TABLE ingredientes (id VARCHAR(36) PRIMARY KEY, receta_id VARCHAR(36) NOT NULL, producto_id VARCHAR(36) NOT NULL, area_id VARCHAR(36) NOT NULL, cantidad FLOAT NOT NULL);
CREATE TABLE ventas (id VARCHAR(36) PRIMARY KEY, receta_nombre VARCHAR(100) NOT NULL, cantidad INTEGER NOT NULL, fecha DATE NOT NULL);
CREATE TABLE movimientos (id VARCHAR(36) PRIMARY KEY, tipo VARCHAR(7) NOT NULL, producto_id VARCHAR(36) NOT NULL, area_id VARCHAR(36) NOT NULL, cantidad FLOAT NOT NULL, fecha DATETIME, motivo VARCHAR(13), comentarios TEXT);
CREATE TABLE inventario_diario (id VARCHAR(36) PRIMARY KEY, fecha DATE NOT NULL, area_id VARCHAR(36) NOT NULL, producto_id VARCHAR(36) NOT NULL, inicio FLOAT, entradas FLOAT, consumo FLOAT, merma FLOAT, otras_salidas FLOAT, final_fisico FLOAT, final_teorico FLOAT, diferencia FLOAT, comentario TEXT, CONSTRAINT _fecha_area_producto_uc UNIQUE (fecha, area_id, producto_id));
CREATE TABLE modelo_ipv (id VARCHAR(36) PRIMARY KEY, area_id VARCHAR(36) NOT NULL, producto_id VARCHAR(36) NOT NULL, orden INTEGER NOT NULL, CONSTRAINT _area_producto_uc UNIQUE (area_id, producto_id));
CREATE TABLE historial_cambios (id VARCHAR(36) PRIMARY KEY, entidad_tipo VARCHAR(50) NOT NULL, entidad_id VARCHAR(36) NOT NULL, campo_modificado VARCHAR(50) NOT NULL, valor_anterior VARCHAR(255), valor_nuevo VARCHAR(255), fecha_cambio DATETIME);
CREATE TABLE alembic_version (version_num VARCHAR(32) PRIMARY KEY);
INSERT INTO alembic_version VALUES ('65b880ef501d');
INSERT INTO grupos VALUES ('g1','PRINCIPALES');
INSERT INTO productos VALUES ('p1','ACEITE','L'),('p2','SAL','KG');
INSERT INTO areas VALUES ('a1','COCINA','COC');
INSERT INTO recetas VALUES ('r1','PASTA',1,'g1');
INSERT INTO ingredientes VALUES ('i1','r1','p1','a1',0.05);
INSERT INTO ventas VALUES ('v1','PASTA',10,'2026-09-05');
INSERT INTO inventario_diario (id,fecha,area_id,producto_id,inicio,final_fisico) VALUES ('d1','2026-09-04','a1','p1',5,5);
`

func writeLegacyDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy", "inventario.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("creando BD legada: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImportLegacyYBaseline(t *testing.T) {
	legacy := writeLegacyDB(t)
	dataDir := filepath.Join(t.TempDir(), "data")

	t.Setenv("CONTROL_IPV_IMPORT_DB", legacy)
	t.Setenv("CONTROL_IPV_DATA_DIR", dataDir)
	t.Setenv("CONTROL_IPV_DB_PATH", "")

	cfg, err := platform.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app, err := appboot.New(cfg, logger, nil)
	if err != nil {
		t.Fatalf("appboot.New: %v", err)
	}
	t.Cleanup(func() { _ = app.Close() })

	// El fichero de destino y su copia .pre-go.bak existen; el origen sigue ahí.
	if _, err := os.Stat(app.DBPath); err != nil {
		t.Fatalf("no se creó la BD de destino: %v", err)
	}
	if _, err := os.Stat(app.DBPath + ".pre-go.bak"); err != nil {
		t.Errorf("no se creó .pre-go.bak: %v", err)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("el origen no debería haberse tocado: %v", err)
	}

	// goose sellado en la versión 1 sin recrear tablas.
	var v int64
	if err := app.DB.QueryRow(
		"SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1").Scan(&v); err != nil {
		t.Fatalf("leyendo goose_db_version: %v", err)
	}
	if v != 1 {
		t.Fatalf("versión de goose = %d, se esperaba 1", v)
	}

	// Los datos se conservan y fluyen por la API Go.
	var (
		nProd  int
		nVenta int
		nInv   int
	)
	app.DB.QueryRow("SELECT COUNT(*) FROM productos").Scan(&nProd)
	app.DB.QueryRow("SELECT COUNT(*) FROM ventas").Scan(&nVenta)
	app.DB.QueryRow("SELECT COUNT(*) FROM inventario_diario").Scan(&nInv)
	if nProd != 2 || nVenta != 1 || nInv != 1 {
		t.Fatalf("datos no conservados: productos=%d ventas=%d inventario=%d", nProd, nVenta, nInv)
	}

	// La columna grupo_id (que el código Go no usa) sigue intacta.
	var grupoID sql.NullString
	if err := app.DB.QueryRow("SELECT grupo_id FROM recetas WHERE id='r1'").Scan(&grupoID); err != nil {
		t.Fatalf("leyendo recetas.grupo_id: %v", err)
	}
	if grupoID.String != "g1" {
		t.Errorf("recetas.grupo_id = %q, se esperaba 'g1'", grupoID.String)
	}

	// Segundo arranque: idempotente, no reimporta ni re-migra.
	app.Close()
	app2, err := appboot.New(cfg, logger, nil)
	if err != nil {
		t.Fatalf("segundo appboot.New: %v", err)
	}
	t.Cleanup(func() { _ = app2.Close() })
	var v2 int64
	app2.DB.QueryRow("SELECT MAX(version_id) FROM goose_db_version").Scan(&v2)
	if v2 != 1 {
		t.Fatalf("segunda pasada: versión = %d", v2)
	}
}

// via HTTP: el escenario real fluye por los handlers.
func TestImportLegacy_APIResponde(t *testing.T) {
	legacy := writeLegacyDB(t)
	t.Setenv("CONTROL_IPV_IMPORT_DB", legacy)
	t.Setenv("CONTROL_IPV_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	t.Setenv("CONTROL_IPV_DB_PATH", "")

	cfg, _ := platform.LoadConfig()
	app, err := appboot.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.Close() })

	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/productos/", nil))
	if rec.Code != 200 {
		t.Fatalf("GET /api/productos/ -> %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"ACEITE"`) || !strings.Contains(body, `"SAL"`) {
		t.Fatalf("cuerpo inesperado: %s", body)
	}
}

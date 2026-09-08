package excel

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// fixture localiza migration/goldens/fixtures/<name>, generado por la versión
// Python (backend/tests/make_excel_fixtures.py).
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "migration", "goldens", "fixtures", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no se encontró el fixture %s (ejecuta `make excel-fixtures`): %v", name, err)
	}
	return b
}

func TestParseProductos_FixturePython(t *testing.T) {
	rows, err := ParseProductos(bytes.NewReader(fixture(t, "productos.xlsx")))
	if err != nil {
		t.Fatalf("ParseProductos: %v", err)
	}
	want := []ProductoRow{
		{"ACEITE", "L"}, {"SAL", "KG"}, {"RON", "L"},
	}
	if len(rows) != len(want) {
		t.Fatalf("filas = %d, se esperaban %d: %+v", len(rows), len(want), rows)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("fila %d = %+v, se esperaba %+v", i, rows[i], want[i])
		}
	}
}

func TestParseRecetas_FixturePython(t *testing.T) {
	rows, err := ParseRecetas(bytes.NewReader(fixture(t, "recetas.xlsx")))
	if err != nil {
		t.Fatalf("ParseRecetas: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("filas = %d, se esperaban 4: %+v", len(rows), rows)
	}
	if rows[0] != (RecetaRow{"PASTA", "ACEITE", "L", "COCINA", 0.05}) {
		t.Errorf("fila 0 = %+v", rows[0])
	}
	// AGUA: receta sin ingredientes -> fila con campos vacíos y cantidad 0.
	if rows[3].RecetaNombre != "AGUA" || rows[3].ProductoNombre != "" || rows[3].Cantidad != 0 {
		t.Errorf("fila AGUA = %+v", rows[3])
	}
}

func TestParseVentas_FixturePython(t *testing.T) {
	rows, err := ParseVentas(bytes.NewReader(fixture(t, "ventas_import.xlsx")))
	if err != nil {
		t.Fatalf("ParseVentas: %v", err)
	}
	if len(rows) != 5 {
		t.Fatalf("filas = %d, se esperaban 5: %+v", len(rows), rows)
	}
	// El número de fila debe corresponder a la hoja (cabecera = 1).
	if rows[0].Fila != 2 || rows[4].Fila != 6 {
		t.Errorf("números de fila = %d..%d, se esperaba 2..6", rows[0].Fila, rows[4].Fila)
	}
	check := map[string]bool{ // Nombre -> CantidadValida esperada
		"PASTA": true, "MOJITO": true, "CAFE": false, "TE": false,
	}
	for _, r := range rows[:4] {
		if r.CantidadValida != check[r.Nombre] {
			t.Errorf("%s: CantidadValida = %v, se esperaba %v", r.Nombre, r.CantidadValida, check[r.Nombre])
		}
	}
	if !rows[4].CantidadValida || rows[4].Cantidad != 2.5 {
		t.Errorf("fila 6 (PASTA 2.5) = %+v", rows[4])
	}
}

func TestParseProductos_ColumnasFaltantes(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteRecetas(&buf, nil); err != nil { // hoja con columnas de recetas
		t.Fatal(err)
	}
	_, err := ParseProductos(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatal("se esperaba error por columnas faltantes")
	}
	if got := err.Error(); got != "El archivo Excel debe contener las columnas: nombre, unidad_medida" {
		t.Errorf("mensaje = %q", got)
	}
}

func TestRoundTripProductos(t *testing.T) {
	in := []ProductoRow{{"ACEITE", "L"}, {"HARINA 000", "KG"}}
	var buf bytes.Buffer
	if err := WriteProductos(&buf, in); err != nil {
		t.Fatalf("WriteProductos: %v", err)
	}
	out, err := ParseProductos(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ParseProductos: %v", err)
	}
	if len(out) != len(in) || out[0] != in[0] || out[1] != in[1] {
		t.Fatalf("round-trip = %+v, se esperaba %+v", out, in)
	}
}

func TestRoundTripRecetas(t *testing.T) {
	in := []RecetaRow{
		{"PASTA", "ACEITE", "L", "COCINA", 0.05},
		{"AGUA", "", "", "", 0}, // sin ingredientes
	}
	var buf bytes.Buffer
	if err := WriteRecetas(&buf, in); err != nil {
		t.Fatalf("WriteRecetas: %v", err)
	}
	out, err := ParseRecetas(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ParseRecetas: %v", err)
	}
	if len(out) != 2 || out[0] != in[0] || out[1].RecetaNombre != "AGUA" || out[1].Cantidad != 0 {
		t.Fatalf("round-trip = %+v", out)
	}
}

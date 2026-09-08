package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/adapters/sqlite"
	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
	"github.com/Rob102194/control_ipv_tool/internal/httpapi"
)

// Arnés de paridad: monta la API Go sobre el mismo escenario determinista que
// generó migration/goldens/, hace las mismas peticiones y compara el cuerpo JSON
// (estructural, ignorando los campos "id" que la versión Python genera al azar).

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time     { return c.t }
func (c fixedClock) Today() domain.Date { return domain.DateFromTime(c.t) }

type seqIDs struct{ n *int }

func (g seqIDs) New() string { *g.n++; return "gen-" + itoaTest(*g.n) }

func itoaTest(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func goldenDir() string { return filepath.Join("..", "..", "migration", "goldens") }

func loadGolden(t *testing.T, name string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(goldenDir(), name+".json"))
	if err != nil {
		t.Fatalf("golden %s: %v", name, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("golden %s inválido: %v", name, err)
	}
	return doc
}

// normalizeIDs sustituye recursivamente cualquier valor bajo la clave "id".
func normalizeIDs(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if k == "id" {
				x[k] = "<ID>"
			} else {
				x[k] = normalizeIDs(val)
			}
		}
		return x
	case []any:
		for i := range x {
			x[i] = normalizeIDs(x[i])
		}
		return x
	default:
		return v
	}
}

func parseBody(t *testing.T, r io.Reader) any {
	t.Helper()
	var v any
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		t.Fatalf("cuerpo no es JSON: %v", err)
	}
	return v
}

// seedParidad carga el escenario de conftest.py con ids fijos, directamente por
// el Store (sin pasar por los casos de uso, para conservar los ids).
func seedParidad(t *testing.T, st *sqlite.Store) {
	t.Helper()
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	mustR := func(_ domain.Receta, err error) { must(err) }

	for _, p := range []domain.Producto{
		{ID: "prod-aceite", Nombre: "ACEITE", UnidadMedida: "L"},
		{ID: "prod-sal", Nombre: "SAL", UnidadMedida: "KG"},
		{ID: "prod-ron", Nombre: "RON", UnidadMedida: "L"},
		{ID: "prod-limon", Nombre: "LIMON", UnidadMedida: "U"},
	} {
		_, err := st.Productos().Crear(ctx, p)
		must(err)
	}
	for _, a := range []domain.Area{
		{ID: "area-cocina", Nombre: "COCINA", Codigo: "COC"},
		{ID: "area-bar", Nombre: "BAR", Codigo: "BAR"},
	} {
		_, err := st.Areas().Crear(ctx, a)
		must(err)
	}
	mustR(st.Recetas().Crear(ctx, domain.Receta{ID: "rec-pasta", Nombre: "PASTA", Activa: true, Ingredientes: []domain.Ingrediente{
		{ID: "ing-1", RecetaID: "rec-pasta", ProductoID: "prod-aceite", AreaID: "area-cocina", Cantidad: 0.05},
		{ID: "ing-2", RecetaID: "rec-pasta", ProductoID: "prod-sal", AreaID: "area-cocina", Cantidad: 0.01},
	}}))
	mustR(st.Recetas().Crear(ctx, domain.Receta{ID: "rec-mojito", Nombre: "MOJITO", Activa: true, Ingredientes: []domain.Ingrediente{
		{ID: "ing-3", RecetaID: "rec-mojito", ProductoID: "prod-ron", AreaID: "area-bar", Cantidad: 0.05},
		{ID: "ing-4", RecetaID: "rec-mojito", ProductoID: "prod-limon", AreaID: "area-bar", Cantidad: 1},
	}}))

	must(st.ModelosIPV().GuardarModelo(ctx, "area-cocina", []domain.ModeloIPV{
		{ID: "m-1", AreaID: "area-cocina", ProductoID: "prod-aceite", Orden: 0},
		{ID: "m-2", AreaID: "area-cocina", ProductoID: "prod-sal", Orden: 1},
	}))
	must(st.ModelosIPV().GuardarModelo(ctx, "area-bar", []domain.ModeloIPV{
		{ID: "m-3", AreaID: "area-bar", ProductoID: "prod-ron", Orden: 0},
		{ID: "m-4", AreaID: "area-bar", ProductoID: "prod-limon", Orden: 1},
	}))

	_, err := st.Ventas().CrearMultiples(ctx, []domain.Venta{
		{ID: "v-1", RecetaNombre: "PASTA", Cantidad: 10, Fecha: domain.MustParseDate("2026-09-05")},
		{ID: "v-2", RecetaNombre: "MOJITO", Cantidad: 8, Fecha: domain.MustParseDate("2026-09-05")},
	})
	must(err)

	prev := []domain.InventarioDiario{
		{ID: "inv-p1", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-cocina", ProductoID: "prod-aceite", FinalFisico: 5, FinalTeorico: 5},
		{ID: "inv-p2", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-cocina", ProductoID: "prod-sal", FinalFisico: 2, FinalTeorico: 2},
		{ID: "inv-p3", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-bar", ProductoID: "prod-ron", FinalFisico: 3, FinalTeorico: 3},
		{ID: "inv-p4", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-bar", ProductoID: "prod-limon", FinalFisico: 20, FinalTeorico: 20},
	}
	must(st.InventarioDiario().GuardarTodos(ctx, prev))
}

func parityServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "parity.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := sqlite.Migrate(db, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st := sqlite.NewStore(db)
	seedParidad(t, st)

	n := 0
	svc := usecases.New(usecases.Deps{
		Repos: st, UOW: st,
		Clock: fixedClock{t: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)},
		IDs:   seqIDs{n: &n},
	})
	srv := httptest.NewServer(httpapi.NewRouter(httpapi.Deps{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: svc,
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParidad_Listados(t *testing.T) {
	srv := parityServer(t)
	for _, tc := range []struct{ golden, url string }{
		{"productos_list", "/api/productos/"},
		{"areas_list", "/api/areas/"},
		{"recetas_list", "/api/recetas/"},
		{"ventas_list", "/api/ventas/"},
		{"ipv_calcular_consumo", "/api/ipv/calcular-consumo?fecha=2026-09-05"},
		{"ipv_estado_plantilla", "/api/ipv/estado?fecha=2026-09-05"},
	} {
		t.Run(tc.golden, func(t *testing.T) {
			resp, err := http.Get(srv.URL + tc.url)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.url, err)
			}
			defer resp.Body.Close()
			got := normalizeIDs(parseBody(t, resp.Body))

			g := loadGolden(t, tc.golden)
			wantStatus := int(g["response"].(map[string]any)["status"].(float64))
			if resp.StatusCode != wantStatus {
				t.Fatalf("status = %d, golden = %d", resp.StatusCode, wantStatus)
			}
			want := normalizeIDs(g["response"].(map[string]any)["body"])
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("paridad rota\n golden: %s\n Go:     %s", mustJSON(want), mustJSON(got))
			}
		})
	}
}

func TestParidad_GuardarYReporte(t *testing.T) {
	srv := parityServer(t)

	// 1) POST /api/ipv/guardar con el cuerpo del golden.
	gg := loadGolden(t, "ipv_guardar")
	reqBody := gg["request"].(map[string]any)["body"]
	payload := mustJSON(reqBody)

	resp, err := http.Post(srv.URL+"/api/ipv/guardar", "application/json", bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatalf("POST guardar: %v", err)
	}
	got := normalizeIDs(parseBody(t, resp.Body))
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("status guardar = %d", resp.StatusCode)
	}
	want := normalizeIDs(gg["response"].(map[string]any)["body"])
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("guardar: paridad rota\n golden: %s\n Go:     %s", mustJSON(want), mustJSON(got))
	}

	// 2) GET /api/ipv/reporte tras guardar.
	gr := loadGolden(t, "ipv_reporte")
	resp2, err := http.Get(srv.URL + "/api/ipv/reporte?fecha=2026-09-05")
	if err != nil {
		t.Fatalf("GET reporte: %v", err)
	}
	got2 := normalizeIDs(parseBody(t, resp2.Body))
	resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("status reporte = %d", resp2.StatusCode)
	}
	want2 := normalizeIDs(gr["response"].(map[string]any)["body"])
	if !reflect.DeepEqual(got2, want2) {
		t.Fatalf("reporte: paridad rota\n golden: %s\n Go:     %s", mustJSON(want2), mustJSON(got2))
	}
}

func TestParidad_ReporteFechaVaciaEs404(t *testing.T) {
	srv := parityServer(t)
	resp, err := http.Get(srv.URL + "/api/ipv/reporte?fecha=2020-01-01")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404", resp.StatusCode)
	}
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["error"] != "No se encontraron registros para la fecha especificada." {
		t.Fatalf("mensaje = %q", body["error"])
	}
}

var _ = ports.ErrNoEncontrado

func mustJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

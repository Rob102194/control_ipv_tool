package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// newTestStore abre una BD temporal migrada y devuelve el Store.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "repos.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := Migrate(db, newTestLogger()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return NewStore(db)
}

// seedEscenario carga el mismo escenario determinista que los goldens de la
// Fase 0 (10×PASTA, 8×MOJITO; cierre del día anterior).
func seedEscenario(t *testing.T, s *Store) {
	t.Helper()
	ctx := context.Background()

	productos := []domain.Producto{
		{ID: "prod-aceite", Nombre: "ACEITE", UnidadMedida: "L"},
		{ID: "prod-sal", Nombre: "SAL", UnidadMedida: "KG"},
		{ID: "prod-ron", Nombre: "RON", UnidadMedida: "L"},
		{ID: "prod-limon", Nombre: "LIMON", UnidadMedida: "U"},
	}
	for _, p := range productos {
		if _, err := s.Productos().Crear(ctx, p); err != nil {
			t.Fatalf("crear producto %s: %v", p.ID, err)
		}
	}
	for _, a := range []domain.Area{
		{ID: "area-cocina", Nombre: "COCINA", Codigo: "COC"},
		{ID: "area-bar", Nombre: "BAR", Codigo: "BAR"},
	} {
		if _, err := s.Areas().Crear(ctx, a); err != nil {
			t.Fatalf("crear area %s: %v", a.ID, err)
		}
	}

	recetas := []domain.Receta{
		{ID: "rec-pasta", Nombre: "PASTA", Activa: true, Ingredientes: []domain.Ingrediente{
			{ID: "ing-1", ProductoID: "prod-aceite", AreaID: "area-cocina", Cantidad: 0.05},
			{ID: "ing-2", ProductoID: "prod-sal", AreaID: "area-cocina", Cantidad: 0.01},
		}},
		{ID: "rec-mojito", Nombre: "MOJITO", Activa: true, Ingredientes: []domain.Ingrediente{
			{ID: "ing-3", ProductoID: "prod-ron", AreaID: "area-bar", Cantidad: 0.05},
			{ID: "ing-4", ProductoID: "prod-limon", AreaID: "area-bar", Cantidad: 1},
		}},
	}
	for _, rec := range recetas {
		if _, err := s.Recetas().Crear(ctx, rec); err != nil {
			t.Fatalf("crear receta %s: %v", rec.ID, err)
		}
	}

	modelo := []domain.ModeloIPV{
		{ID: "m-1", ProductoID: "prod-aceite", Orden: 0},
		{ID: "m-2", ProductoID: "prod-sal", Orden: 1},
	}
	if err := s.ModelosIPV().GuardarModelo(ctx, "area-cocina", modelo); err != nil {
		t.Fatalf("guardar modelo cocina: %v", err)
	}
	if err := s.ModelosIPV().GuardarModelo(ctx, "area-bar", []domain.ModeloIPV{
		{ID: "m-3", ProductoID: "prod-ron", Orden: 0},
		{ID: "m-4", ProductoID: "prod-limon", Orden: 1},
	}); err != nil {
		t.Fatalf("guardar modelo bar: %v", err)
	}

	ventas := []domain.Venta{
		{ID: "v-1", RecetaNombre: "PASTA", Cantidad: 10, Fecha: domain.MustParseDate("2026-09-05")},
		{ID: "v-2", RecetaNombre: "MOJITO", Cantidad: 8, Fecha: domain.MustParseDate("2026-09-05")},
	}
	if _, err := s.Ventas().CrearMultiples(ctx, ventas); err != nil {
		t.Fatalf("crear ventas: %v", err)
	}

	// Cierre del día anterior (alimenta el `inicio` del 2026-09-05).
	prev := []domain.InventarioDiario{
		{ID: "inv-p1", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-cocina", ProductoID: "prod-aceite", FinalFisico: 5},
		{ID: "inv-p2", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-cocina", ProductoID: "prod-sal", FinalFisico: 2},
		{ID: "inv-p3", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-bar", ProductoID: "prod-ron", FinalFisico: 3},
		{ID: "inv-p4", Fecha: domain.MustParseDate("2026-09-04"), AreaID: "area-bar", ProductoID: "prod-limon", FinalFisico: 20},
	}
	if err := s.InventarioDiario().GuardarTodos(ctx, prev); err != nil {
		t.Fatalf("guardar inventario previo: %v", err)
	}
}

func TestListados_OrdenParidad(t *testing.T) {
	s := newTestStore(t)
	seedEscenario(t, s)
	ctx := context.Background()

	prods, err := s.Productos().Listar(ctx, ports.OrdenProductoNombre)
	if err != nil {
		t.Fatalf("listar productos: %v", err)
	}
	gotP := names(prods, func(p domain.Producto) string { return p.Nombre })
	wantP := []string{"ACEITE", "LIMON", "RON", "SAL"} // alfabético, como productos_list.json
	assertStrSlice(t, "productos", gotP, wantP)

	areas, err := s.Areas().Listar(ctx)
	if err != nil {
		t.Fatalf("listar areas: %v", err)
	}
	gotA := names(areas, func(a domain.Area) string { return a.Nombre })
	assertStrSlice(t, "areas", gotA, []string{"COCINA", "BAR"}) // orden de inserción

	recs, err := s.Recetas().Listar(ctx, ports.ListarRecetasOpts{})
	if err != nil {
		t.Fatalf("listar recetas: %v", err)
	}
	assertStrSlice(t, "recetas", names(recs, func(r domain.Receta) string { return r.Nombre }),
		[]string{"MOJITO", "PASTA"})
	// Ingredientes en orden de inserción.
	if len(recs[0].Ingredientes) != 2 || recs[0].Ingredientes[0].ProductoID != "prod-ron" {
		t.Fatalf("ingredientes de MOJITO = %+v", recs[0].Ingredientes)
	}
}

func TestInventario_ArrastreYUpsert(t *testing.T) {
	s := newTestStore(t)
	seedEscenario(t, s)
	ctx := context.Background()

	ff, err := s.InventarioDiario().FinalFisicoDiaAnterior(ctx, domain.MustParseDate("2026-09-05"), "area-cocina", "prod-aceite")
	if err != nil {
		t.Fatalf("FinalFisicoDiaAnterior: %v", err)
	}
	if ff != 5 {
		t.Fatalf("arrastre aceite = %v, se esperaba 5", ff)
	}
	// Producto sin registro previo -> 0.
	ff0, _ := s.InventarioDiario().FinalFisicoDiaAnterior(ctx, domain.MustParseDate("2026-09-05"), "area-bar", "prod-inexistente")
	if ff0 != 0 {
		t.Fatalf("arrastre inexistente = %v, se esperaba 0", ff0)
	}

	fila := domain.InventarioDiario{
		ID: "s-1", Fecha: domain.MustParseDate("2026-09-05"),
		AreaID: "area-cocina", ProductoID: "prod-aceite",
		Inicio: 5, Entradas: 2, Consumo: 0.5, Merma: 0.1, FinalFisico: 6.2,
	}
	fila.CalcularDiferencias()
	if err := s.InventarioDiario().GuardarTodos(ctx, []domain.InventarioDiario{fila}); err != nil {
		t.Fatalf("guardar (insert): %v", err)
	}
	// Segundo guardado con la misma clave (fecha, area, producto) -> UPDATE.
	fila.Entradas = 3
	fila.CalcularDiferencias()
	if err := s.InventarioDiario().GuardarTodos(ctx, []domain.InventarioDiario{fila}); err != nil {
		t.Fatalf("guardar (upsert): %v", err)
	}

	got, err := s.InventarioDiario().ListarPorFecha(ctx, domain.MustParseDate("2026-09-05"))
	if err != nil {
		t.Fatalf("listar por fecha: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("filas = %d, se esperaba 1 (upsert, no duplicado)", len(got))
	}
	if got[0].Entradas != 3 || got[0].FinalTeorico != 7.4 {
		t.Fatalf("fila tras upsert = %+v", got[0])
	}

	fechas, err := s.InventarioDiario().FechasConRegistros(ctx)
	if err != nil {
		t.Fatalf("FechasConRegistros: %v", err)
	}
	if len(fechas) != 2 || fechas[0].String() != "2026-09-05" || fechas[1].String() != "2026-09-04" {
		t.Fatalf("fechas = %v (se esperaba desc)", fechas)
	}
}

func TestProducto_EnUso_Conflicto_NoEncontrado(t *testing.T) {
	s := newTestStore(t)
	seedEscenario(t, s)
	ctx := context.Background()

	if enUso, _ := s.Productos().EnUso(ctx, "prod-aceite"); !enUso {
		t.Error("prod-aceite debería estar en uso (ingrediente + modelo)")
	}
	if _, err := s.Productos().Crear(ctx, domain.Producto{ID: "x", Nombre: "LIBRE", UnidadMedida: "U"}); err != nil {
		t.Fatalf("crear producto libre: %v", err)
	}
	if enUso, _ := s.Productos().EnUso(ctx, "x"); enUso {
		t.Error("LIBRE no debería estar en uso")
	}

	// Nombre duplicado -> ConflictError.
	_, err := s.Productos().Crear(ctx, domain.Producto{ID: "y", Nombre: "ACEITE", UnidadMedida: "L"})
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("nombre duplicado: se esperaba ConflictError, se obtuvo %v", err)
	}

	// No encontrado.
	_, err = s.Productos().ObtenerPorID(ctx, "no-existe")
	if !errors.Is(err, ports.ErrNoEncontrado) {
		t.Fatalf("ObtenerPorID inexistente: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
	if err := s.Productos().Eliminar(ctx, "no-existe"); !errors.Is(err, ports.ErrNoEncontrado) {
		t.Fatalf("Eliminar inexistente: se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

func TestReceta_ActualizarYEliminar(t *testing.T) {
	s := newTestStore(t)
	seedEscenario(t, s)
	ctx := context.Background()

	rec, err := s.Recetas().ObtenerPorID(ctx, "rec-pasta")
	if err != nil {
		t.Fatalf("obtener receta: %v", err)
	}
	rec.Ingredientes = []domain.Ingrediente{
		{ID: "ing-nuevo", ProductoID: "prod-aceite", AreaID: "area-cocina", Cantidad: 0.09},
	}
	if _, err := s.Recetas().Actualizar(ctx, rec); err != nil {
		t.Fatalf("actualizar receta: %v", err)
	}
	got, _ := s.Recetas().ObtenerPorID(ctx, "rec-pasta")
	if len(got.Ingredientes) != 1 || got.Ingredientes[0].ID != "ing-nuevo" || got.Ingredientes[0].Cantidad != 0.09 {
		t.Fatalf("ingredientes tras actualizar = %+v", got.Ingredientes)
	}

	if err := s.Recetas().Eliminar(ctx, "rec-pasta"); err != nil {
		t.Fatalf("eliminar receta: %v", err)
	}
	if _, err := s.Recetas().ObtenerPorID(ctx, "rec-pasta"); !errors.Is(err, ports.ErrNoEncontrado) {
		t.Fatalf("receta debería haber desaparecido, err = %v", err)
	}
	// Sus ingredientes también.
	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM ingredientes WHERE receta_id = 'rec-pasta'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("quedaron %d ingredientes huérfanos", n)
	}
}

func TestUnitOfWork_Rollback(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sentinela := errors.New("aborta")
	err := s.Do(ctx, func(r ports.Repos) error {
		if _, err := r.Productos().Crear(ctx, domain.Producto{ID: "tmp", Nombre: "TMP", UnidadMedida: "U"}); err != nil {
			return err
		}
		return sentinela
	})
	if !errors.Is(err, sentinela) {
		t.Fatalf("Do debería devolver el error de fn, devolvió %v", err)
	}
	if _, err := s.Productos().ObtenerPorID(ctx, "tmp"); !errors.Is(err, ports.ErrNoEncontrado) {
		t.Fatalf("el rollback no revirtió la inserción, err = %v", err)
	}
}

func TestHistorial_Orden(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	base := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)

	for i, campo := range []string{"nombre", "unidad_medida", "Creación"} {
		h := domain.HistorialCambios{
			ID: "h-" + campo, EntidadTipo: domain.EntidadProducto, EntidadID: "p1",
			CampoModificado: campo, FechaCambio: base.Add(time.Duration(i) * time.Hour),
		}
		if err := s.Historial().Registrar(ctx, h); err != nil {
			t.Fatalf("registrar %s: %v", campo, err)
		}
	}
	got, err := s.Historial().PorEntidad(ctx, domain.EntidadProducto)
	if err != nil {
		t.Fatalf("PorEntidad: %v", err)
	}
	if len(got) != 3 || got[0].CampoModificado != "Creación" || got[2].CampoModificado != "nombre" {
		t.Fatalf("orden del historial = %v (se esperaba fecha desc)", names(got, func(h domain.HistorialCambios) string { return h.CampoModificado }))
	}
	if !got[0].FechaCambio.Equal(base.Add(2 * time.Hour)) {
		t.Fatalf("fecha_cambio no round-trippeó: %v", got[0].FechaCambio)
	}
}

// --- helpers ---

func names[T any](xs []T, f func(T) string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

func assertStrSlice(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, se esperaba %v", what, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, se esperaba %v", what, got, want)
		}
	}
}

package usecases_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/adapters/excel"
	"github.com/Rob102194/control_ipv_tool/internal/adapters/sqlite"
	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// --- dobles deterministas -------------------------------------------------

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time     { c.t = c.t.Add(time.Second); return c.t }
func (c *fakeClock) Today() domain.Date { return domain.DateFromTime(c.t) }

type seqIDs struct{ n int }

func (g *seqIDs) New() string { g.n++; return fmt.Sprintf("id-%04d", g.n) }

func newServices(t *testing.T) (*usecases.Services, *sqlite.Store) {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "uc.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := sqlite.Migrate(db, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st := sqlite.NewStore(db)
	svc := usecases.New(usecases.Deps{
		Repos: st, UOW: st,
		Clock: &fakeClock{t: time.Date(2026, 9, 5, 8, 0, 0, 0, time.UTC)},
		IDs:   &seqIDs{},
	})
	return svc, st
}

func fecha(s string) domain.Date { return domain.MustParseDate(s) }

func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "migration", "goldens", "fixtures", name))
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return b
}

// seedIPV carga el escenario determinista de los goldens de la Fase 0.
func seedIPV(t *testing.T, svc *usecases.Services) {
	t.Helper()
	ctx := context.Background()
	crearP := func(n, um string) string {
		p, err := svc.Productos.Crear(ctx, usecases.ProductoInput{Nombre: n, UnidadMedida: um})
		if err != nil {
			t.Fatalf("crear producto %s: %v", n, err)
		}
		return p.ID
	}
	crearA := func(n string) string {
		a, err := svc.Areas.Crear(ctx, usecases.AreaInput{Nombre: n})
		if err != nil {
			t.Fatalf("crear area %s: %v", n, err)
		}
		return a.ID
	}
	aceite, sal := crearP("ACEITE", "L"), crearP("SAL", "KG")
	ron, limon := crearP("RON", "L"), crearP("LIMON", "U")
	cocina, bar := crearA("COCINA"), crearA("BAR")

	if _, err := svc.Recetas.Crear(ctx, usecases.RecetaInput{Nombre: "PASTA", Activa: true, Ingredientes: []usecases.IngredienteInput{
		{ProductoID: aceite, AreaID: cocina, Cantidad: 0.05},
		{ProductoID: sal, AreaID: cocina, Cantidad: 0.01},
	}}); err != nil {
		t.Fatalf("crear PASTA: %v", err)
	}
	if _, err := svc.Recetas.Crear(ctx, usecases.RecetaInput{Nombre: "MOJITO", Activa: true, Ingredientes: []usecases.IngredienteInput{
		{ProductoID: ron, AreaID: bar, Cantidad: 0.05},
		{ProductoID: limon, AreaID: bar, Cantidad: 1},
	}}); err != nil {
		t.Fatalf("crear MOJITO: %v", err)
	}

	if err := svc.IPV.GuardarModelo(ctx, cocina, []usecases.ModeloItem{{ProductoID: aceite, Orden: 0}, {ProductoID: sal, Orden: 1}}); err != nil {
		t.Fatalf("modelo cocina: %v", err)
	}
	if err := svc.IPV.GuardarModelo(ctx, bar, []usecases.ModeloItem{{ProductoID: ron, Orden: 0}, {ProductoID: limon, Orden: 1}}); err != nil {
		t.Fatalf("modelo bar: %v", err)
	}

	for _, v := range []usecases.VentaInput{
		{RecetaNombre: "PASTA", Cantidad: 10, Fecha: fecha("2026-09-05")},
		{RecetaNombre: "MOJITO", Cantidad: 8, Fecha: fecha("2026-09-05")},
	} {
		if _, err := svc.Ventas.Crear(ctx, v); err != nil {
			t.Fatalf("crear venta: %v", err)
		}
	}

	// Cierre del día anterior (alimenta el `inicio` del 2026-09-05).
	prev := []usecases.InventarioFilaView{
		{Fecha: fecha("2026-09-04"), AreaID: cocina, ProductoID: aceite, FinalFisico: 5},
		{Fecha: fecha("2026-09-04"), AreaID: cocina, ProductoID: sal, FinalFisico: 2},
		{Fecha: fecha("2026-09-04"), AreaID: bar, ProductoID: ron, FinalFisico: 3},
		{Fecha: fecha("2026-09-04"), AreaID: bar, ProductoID: limon, FinalFisico: 20},
	}
	if _, err := svc.IPV.Guardar(ctx, prev); err != nil {
		t.Fatalf("guardar IPV previo: %v", err)
	}
}

// --- productos ----------------------------------------------------------

func TestProducto_CrearNormalizaYRegistraHistorial(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()

	p, err := svc.Productos.Crear(ctx, usecases.ProductoInput{Nombre: "  aceite ", UnidadMedida: " l "})
	if err != nil {
		t.Fatalf("Crear: %v", err)
	}
	if p.Nombre != "ACEITE" || p.UnidadMedida != "L" {
		t.Fatalf("normalización: %+v", p)
	}

	hist, err := svc.Historial.Listar(ctx, domain.EntidadProducto)
	if err != nil {
		t.Fatalf("historial: %v", err)
	}
	if len(hist) != 1 || hist[0].CampoModificado != "Creación" || hist[0].ValorNuevo != "Producto 'ACEITE' creado" {
		t.Fatalf("historial de creación = %+v", hist)
	}
}

func TestProducto_DuplicadoEsConflicto(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()
	if _, err := svc.Productos.Crear(ctx, usecases.ProductoInput{Nombre: "SAL", UnidadMedida: "KG"}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Productos.Crear(ctx, usecases.ProductoInput{Nombre: "sal", UnidadMedida: "g"})
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("se esperaba ConflictError, se obtuvo %v", err)
	}
}

func TestProducto_EliminarEnUsoEsConflicto(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()
	seedIPV(t, svc)

	prods, _ := svc.Productos.Listar(ctx, ports.OrdenProductoNombre)
	var aceiteID string
	for _, p := range prods {
		if p.Nombre == "ACEITE" {
			aceiteID = p.ID
		}
	}
	err := svc.Productos.Eliminar(ctx, aceiteID)
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("eliminar producto en uso: se esperaba ConflictError, se obtuvo %v", err)
	}
	// Sigue existiendo.
	if _, err := svc.Productos.Obtener(ctx, aceiteID); err != nil {
		t.Fatalf("el producto no debería haberse borrado: %v", err)
	}
}

// --- IPV --------------------------------------------------------------

func TestIPV_ObtenerEstado_Plantilla(t *testing.T) {
	svc, _ := newServices(t)
	seedIPV(t, svc)
	ctx := context.Background()

	estado, err := svc.IPV.ObtenerEstado(ctx, fecha("2026-09-05"))
	if err != nil {
		t.Fatalf("ObtenerEstado: %v", err)
	}
	inicioDe := map[string]float64{}
	for _, filas := range estado {
		for _, f := range filas {
			inicioDe[f.ProductoNombre] = f.Inicio
			if f.Consumo != 0 || f.FinalTeorico != 0 {
				t.Errorf("plantilla no debería traer consumo/final_teorico: %+v", f)
			}
		}
	}
	want := map[string]float64{"ACEITE": 5, "SAL": 2, "RON": 3, "LIMON": 20}
	for k, v := range want {
		if inicioDe[k] != v {
			t.Errorf("inicio[%s] = %v, se esperaba %v", k, inicioDe[k], v)
		}
	}
	if names := namesOf(estado["COCINA"]); names[0] != "ACEITE" || names[1] != "SAL" {
		t.Errorf("orden COCINA = %v (se esperaba modelo: ACEITE, SAL)", names)
	}
}

func TestIPV_CalcularConsumo(t *testing.T) {
	svc, _ := newServices(t)
	seedIPV(t, svc)
	ctx := context.Background()

	consumo, err := svc.IPV.CalcularConsumo(ctx, fecha("2026-09-05"))
	if err != nil {
		t.Fatalf("CalcularConsumo: %v", err)
	}
	if len(consumo) != 4 {
		t.Fatalf("consumo = %v", consumo)
	}
	// Las claves llevan ids "id-XXXX"; comprobamos por valor agregado.
	var total float64
	for _, v := range consumo {
		total += v
	}
	if math.Abs(total-(0.5+0.1+0.4+8.0)) > 1e-9 {
		t.Fatalf("suma de consumo = %v, se esperaba 9.0", total)
	}
}

func TestIPV_GuardarYReporte(t *testing.T) {
	svc, _ := newServices(t)
	seedIPV(t, svc)
	ctx := context.Background()

	e, _ := svc.IPV.ObtenerEstado(ctx, fecha("2026-09-05"))
	// Rellenamos como en ipv_guardar.json.
	set := func(nombre string, entradas, consumo, merma, fisico float64) {
		for area, filas := range e {
			for i := range filas {
				if filas[i].ProductoNombre == nombre {
					filas[i].Entradas, filas[i].Consumo = entradas, consumo
					filas[i].Merma, filas[i].FinalFisico = merma, fisico
				}
			}
			e[area] = filas
		}
	}
	set("ACEITE", 2, 0.5, 0.1, 6.2)
	set("SAL", 0, 0.1, 0, 1.85)
	set("RON", 1, 0.4, 0, 3.6)
	set("LIMON", 0, 8, 2, 10)
	// Comentario en la fila de SAL, como en ipv_guardar.json (s-2).
	for i := range e["COCINA"] {
		if e["COCINA"][i].ProductoNombre == "SAL" {
			e["COCINA"][i].Comentario = `{"merma": "", "diferencia": "faltan 50g"}`
		}
	}

	var todas []usecases.InventarioFilaView
	for _, filas := range e {
		todas = append(todas, filas...)
	}
	guardadas, err := svc.IPV.Guardar(ctx, todas)
	if err != nil {
		t.Fatalf("Guardar: %v", err)
	}
	for _, f := range guardadas {
		if f.ProductoNombre == "ACEITE" {
			if math.Abs(f.FinalTeorico-6.4) > 1e-9 || math.Abs(f.Diferencia-(-0.2)) > 1e-9 {
				t.Fatalf("ACEITE tras guardar = %+v", f)
			}
		}
	}

	rep, err := svc.IPV.GenerarReporte(ctx, fecha("2026-09-05"))
	if err != nil {
		t.Fatalf("GenerarReporte: %v", err)
	}
	if _, ok := rep.Areas["COCINA"]; !ok {
		t.Fatalf("reporte sin COCINA: %+v", rep.Areas)
	}
	falt := rep.Resumen["COCINA"].Faltantes
	if len(falt) == 0 || falt[0].Producto != "ACEITE" || math.Abs(falt[0].Cantidad-0.2) > 1e-9 {
		t.Fatalf("faltantes COCINA = %+v", falt)
	}
	var notaSal *usecases.ReporteNota
	for i := range rep.Notas {
		if rep.Notas[i].Producto == "SAL" {
			notaSal = &rep.Notas[i]
		}
	}
	if notaSal == nil || notaSal.Campo != "diferencia" || notaSal.Texto != "faltan 50g" {
		t.Fatalf("nota SAL = %+v", rep.Notas)
	}
}

func TestIPV_ReporteSinRegistrosEs404(t *testing.T) {
	svc, _ := newServices(t)
	seedIPV(t, svc)
	_, err := svc.IPV.GenerarReporte(context.Background(), fecha("2020-01-01"))
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("se esperaba NotFoundError, se obtuvo %v", err)
	}
}

func TestIPV_Recalcular(t *testing.T) {
	svc, _ := newServices(t)
	out := svc.IPV.Recalcular([]usecases.InventarioFilaView{
		{Inicio: 5, Entradas: 2, Consumo: 0.5, Merma: 0.1, FinalFisico: 6.2},
	})
	if math.Abs(out[0].FinalTeorico-6.4) > 1e-9 || math.Abs(out[0].Diferencia-(-0.2)) > 1e-9 {
		t.Fatalf("Recalcular = %+v", out[0])
	}
}

// --- importaciones ----------------------------------------------------

func TestImportarVentas_FixturePython(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()

	rows, err := excel.ParseVentas(bytes.NewReader(fixtureBytes(t, "ventas_import.xlsx")))
	if err != nil {
		t.Fatalf("ParseVentas: %v", err)
	}
	var items []usecases.ImportVentaItem
	for _, r := range rows {
		items = append(items, usecases.ImportVentaItem{
			Fila: r.Fila, Nombre: r.Nombre, Cantidad: r.Cantidad,
			CantidadValida: r.CantidadValida, RawCantidad: r.RawCantidad,
		})
	}

	// El fixture tiene 2 filas inválidas (CAFE=0, TE=abc) -> ValidationError, nada insertado.
	_, err = svc.Ventas.Importar(ctx, items, nil)
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("se esperaba ValidationError, se obtuvo %v", err)
	}
	if ventas, _ := svc.Ventas.Listar(ctx); len(ventas) != 0 {
		t.Fatalf("no debería haber ventas tras el error (rollback), hay %d", len(ventas))
	}

	// Solo las filas válidas -> se importan y se crean las recetas nuevas.
	var validas []usecases.ImportVentaItem
	for _, it := range items {
		if it.CantidadValida {
			validas = append(validas, it)
		}
	}
	f := fecha("2026-09-05")
	res, err := svc.Ventas.Importar(ctx, validas, &f)
	if err != nil {
		t.Fatalf("Importar (válidas): %v", err)
	}
	if len(res.Ventas) != 3 { // PASTA, MOJITO, PASTA(2.5->2)
		t.Fatalf("ventas importadas = %d", len(res.Ventas))
	}
	if len(res.NuevasRecetas) != 2 { // PASTA, MOJITO (no existían)
		t.Fatalf("recetas nuevas = %d (%v)", len(res.NuevasRecetas), res.NuevasRecetas)
	}
	for _, v := range res.Ventas {
		if v.RecetaNombre == "PASTA" && v.Cantidad == 2 {
			return // el 2.5 se truncó a 2, como int() en Python
		}
	}
	t.Fatalf("no se encontró la venta PASTA con cantidad 2 (truncada): %+v", res.Ventas)
}

func TestImportarProductosYRecetas_FixturePython(t *testing.T) {
	svc, _ := newServices(t)
	ctx := context.Background()

	prodRows, _ := excel.ParseProductos(bytes.NewReader(fixtureBytes(t, "productos.xlsx")))
	var pItems []usecases.ImportProductoItem
	for _, r := range prodRows {
		pItems = append(pItems, usecases.ImportProductoItem{Nombre: r.Nombre, UnidadMedida: r.UnidadMedida})
	}
	res, err := svc.Productos.Importar(ctx, pItems)
	if err != nil || res.Creados != 3 {
		t.Fatalf("importar productos: creados=%d err=%v", res.Creados, err)
	}
	// Reimportar -> todos omitidos.
	res2, _ := svc.Productos.Importar(ctx, pItems)
	if res2.Creados != 0 || res2.Omitidos != 3 {
		t.Fatalf("reimport productos = %+v", res2)
	}

	recRows, _ := excel.ParseRecetas(bytes.NewReader(fixtureBytes(t, "recetas.xlsx")))
	var rItems []usecases.ImportRecetaItem
	for _, r := range recRows {
		rItems = append(rItems, usecases.ImportRecetaItem{
			RecetaNombre: r.RecetaNombre, ProductoNombre: r.ProductoNombre,
			UnidadMedida: r.UnidadMedida, AreaNombre: r.AreaNombre, Cantidad: r.Cantidad,
		})
	}
	rres, err := svc.Recetas.Importar(ctx, rItems)
	if err != nil {
		t.Fatalf("importar recetas: %v", err)
	}
	if rres.Creados != 3 { // PASTA, MOJITO, AGUA
		t.Fatalf("recetas creadas = %d", rres.Creados)
	}
	// Se creó el área BAR y el producto RON al vuelo.
	if _, err := findArea(ctx, svc, "BAR"); err != nil {
		t.Fatalf("no se creó el área BAR: %v", err)
	}
	recs, _ := svc.Recetas.Listar(ctx, ports.ListarRecetasOpts{})
	for _, rc := range recs {
		if rc.Nombre == "AGUA" && len(rc.Ingredientes) != 0 {
			t.Fatalf("AGUA debería no tener ingredientes: %+v", rc.Ingredientes)
		}
		if rc.Nombre == "PASTA" && len(rc.Ingredientes) != 2 {
			t.Fatalf("PASTA debería tener 2 ingredientes: %+v", rc.Ingredientes)
		}
	}
}

// --- helpers de test -------------------------------------------------

func namesOf(filas []usecases.InventarioFilaView) []string {
	out := make([]string, len(filas))
	for i, f := range filas {
		out[i] = f.ProductoNombre
	}
	return out
}

func findArea(ctx context.Context, svc *usecases.Services, nombre string) (domain.Area, error) {
	areas, err := svc.Areas.Listar(ctx)
	if err != nil {
		return domain.Area{}, err
	}
	for _, a := range areas {
		if a.Nombre == nombre {
			return a, nil
		}
	}
	return domain.Area{}, errors.New("no encontrada")
}

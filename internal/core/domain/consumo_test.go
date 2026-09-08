package domain

import (
	"math"
	"sort"
	"testing"
)

// Escenario y valores esperados de migration/goldens/ipv_calcular_consumo.json:
// 10 × PASTA (ACEITE 0.05 + SAL 0.01 en COCINA) y 8 × MOJITO (RON 0.05 +
// LIMON 1 en BAR).
func escenarioSeed() ([]Venta, map[string]Receta) {
	recetas := map[string]Receta{
		"PASTA": {ID: "rec-pasta", Nombre: "PASTA", Activa: true, Ingredientes: []Ingrediente{
			{ProductoID: "prod-aceite", AreaID: "area-cocina", Cantidad: 0.05},
			{ProductoID: "prod-sal", AreaID: "area-cocina", Cantidad: 0.01},
		}},
		"MOJITO": {ID: "rec-mojito", Nombre: "MOJITO", Activa: true, Ingredientes: []Ingrediente{
			{ProductoID: "prod-ron", AreaID: "area-bar", Cantidad: 0.05},
			{ProductoID: "prod-limon", AreaID: "area-bar", Cantidad: 1},
		}},
	}
	ventas := []Venta{
		{ID: "v-1", RecetaNombre: "PASTA", Cantidad: 10, Fecha: MustParseDate("2026-09-05")},
		{ID: "v-2", RecetaNombre: "MOJITO", Cantidad: 8, Fecha: MustParseDate("2026-09-05")},
	}
	return ventas, recetas
}

func TestCalcularConsumo_Seed(t *testing.T) {
	ventas, recetas := escenarioSeed()
	got := CalcularConsumo(ventas, recetas)

	want := map[ConsumoKey]float64{
		{ProductoID: "prod-aceite", AreaID: "area-cocina"}: 0.5,
		{ProductoID: "prod-sal", AreaID: "area-cocina"}:    0.1,
		{ProductoID: "prod-ron", AreaID: "area-bar"}:       0.4,
		{ProductoID: "prod-limon", AreaID: "area-bar"}:     8.0,
	}
	if len(got) != len(want) {
		t.Fatalf("consumo = %v, se esperaban %d claves", got, len(want))
	}
	for k, w := range want {
		if math.Abs(got[k]-w) > 1e-9 {
			t.Errorf("consumo[%v] = %v, se esperaba %v", k, got[k], w)
		}
	}

	// La clave serializada "producto|area" debe reproducir el golden.
	var claves []string
	for k := range got {
		claves = append(claves, k.ProductoID+"|"+k.AreaID)
	}
	sort.Strings(claves)
	wantClaves := []string{
		"prod-aceite|area-cocina", "prod-limon|area-bar",
		"prod-ron|area-bar", "prod-sal|area-cocina",
	}
	for i := range wantClaves {
		if claves[i] != wantClaves[i] {
			t.Errorf("claves serializadas = %v, se esperaba %v", claves, wantClaves)
			break
		}
	}
}

func TestCalcularConsumo_RecetaDesconocidaSeIgnora(t *testing.T) {
	_, recetas := escenarioSeed()
	ventas := []Venta{
		{RecetaNombre: "PASTA", Cantidad: 1, Fecha: MustParseDate("2026-09-05")},
		{RecetaNombre: "PLATO FANTASMA", Cantidad: 99, Fecha: MustParseDate("2026-09-05")},
	}
	got := CalcularConsumo(ventas, recetas)
	if len(got) != 2 {
		t.Fatalf("consumo = %v, se esperaban solo las claves de PASTA", got)
	}
	if math.Abs(got[ConsumoKey{"prod-aceite", "area-cocina"}]-0.05) > 1e-9 {
		t.Errorf("aceite = %v, se esperaba 0.05", got[ConsumoKey{"prod-aceite", "area-cocina"}])
	}
}

func TestCalcularConsumo_Acumula(t *testing.T) {
	recetas := map[string]Receta{
		"CAFE": {Nombre: "CAFE", Ingredientes: []Ingrediente{
			{ProductoID: "p-cafe", AreaID: "a-bar", Cantidad: 0.007},
		}},
	}
	ventas := []Venta{
		{RecetaNombre: "CAFE", Cantidad: 3},
		{RecetaNombre: "CAFE", Cantidad: 2},
	}
	got := CalcularConsumo(ventas, recetas)
	if math.Abs(got[ConsumoKey{"p-cafe", "a-bar"}]-0.035) > 1e-9 {
		t.Errorf("cafe = %v, se esperaba 0.035", got[ConsumoKey{"p-cafe", "a-bar"}])
	}
}

func TestCalcularConsumo_EntradaVacia(t *testing.T) {
	if got := CalcularConsumo(nil, nil); len(got) != 0 {
		t.Errorf("consumo con entrada vacía = %v, se esperaba mapa vacío", got)
	}
}

package domain

import (
	"errors"
	"testing"
)

func TestProductoNormalizarYValidar(t *testing.T) {
	p := Producto{Nombre: "  aceite de oliva ", UnidadMedida: " l "}
	p.Normalizar()
	if p.Nombre != "ACEITE DE OLIVA" || p.UnidadMedida != "L" {
		t.Fatalf("Normalizar -> %+v", p)
	}
	if err := p.Validar(); err != nil {
		t.Fatalf("Validar: %v", err)
	}

	var ve *ValidationError
	if err := (Producto{Nombre: "", UnidadMedida: "L"}).Validar(); !errors.As(err, &ve) || ve.Campo != "nombre" {
		t.Fatalf("se esperaba ValidationError en 'nombre', se obtuvo %v", err)
	}
	if err := (Producto{Nombre: "SAL", UnidadMedida: ""}).Validar(); !errors.As(err, &ve) || ve.Campo != "unidad_medida" {
		t.Fatalf("se esperaba ValidationError en 'unidad_medida', se obtuvo %v", err)
	}
}

func TestAreaNormalizar(t *testing.T) {
	a := Area{Nombre: " bar ", Codigo: " b1 "}
	a.Normalizar()
	if a.Nombre != "BAR" || a.Codigo != "b1" {
		t.Fatalf("Normalizar -> %+v", a)
	}
}

func TestRecetaValidar(t *testing.T) {
	ok := Receta{Nombre: "PASTA", Ingredientes: []Ingrediente{
		{ProductoID: "p", AreaID: "a", Cantidad: 0.05},
	}}
	if err := ok.Validar(); err != nil {
		t.Fatalf("Validar receta válida: %v", err)
	}

	var ve *ValidationError
	bad := Receta{Nombre: "PASTA", Ingredientes: []Ingrediente{
		{ProductoID: "p", AreaID: "a", Cantidad: 0},
	}}
	if err := bad.Validar(); !errors.As(err, &ve) || ve.Campo != "ingredientes" {
		t.Fatalf("se esperaba ValidationError en 'ingredientes', se obtuvo %v", err)
	}

	sinArea := Receta{Nombre: "X", Ingredientes: []Ingrediente{{ProductoID: "p", Cantidad: 1}}}
	if err := sinArea.Validar(); !errors.As(err, &ve) {
		t.Fatalf("ingrediente sin área debería fallar, se obtuvo %v", err)
	}
}

func TestModeloIPVOrdenarYReindexar(t *testing.T) {
	filas := []ModeloIPV{
		{ProductoID: "c", Orden: 5},
		{ProductoID: "a", Orden: 1},
		{ProductoID: "b", Orden: 3},
	}
	OrdenarModelo(filas)
	if filas[0].ProductoID != "a" || filas[1].ProductoID != "b" || filas[2].ProductoID != "c" {
		t.Fatalf("OrdenarModelo -> %+v", filas)
	}
	ReindexarModelo(filas)
	for i, f := range filas {
		if f.Orden != i {
			t.Fatalf("ReindexarModelo: fila %d tiene Orden %d", i, f.Orden)
		}
	}
}

func TestVentaValidar(t *testing.T) {
	var ve *ValidationError
	if err := (Venta{RecetaNombre: "", Cantidad: 1, Fecha: MustParseDate("2026-09-05")}).Validar(); !errors.As(err, &ve) {
		t.Errorf("receta vacía debería fallar")
	}
	if err := (Venta{RecetaNombre: "X", Cantidad: 0, Fecha: MustParseDate("2026-09-05")}).Validar(); !errors.As(err, &ve) {
		t.Errorf("cantidad 0 debería fallar")
	}
	if err := (Venta{RecetaNombre: "X", Cantidad: 1}).Validar(); !errors.As(err, &ve) {
		t.Errorf("fecha cero debería fallar")
	}
	if err := (Venta{RecetaNombre: "X", Cantidad: 1, Fecha: MustParseDate("2026-09-05")}).Validar(); err != nil {
		t.Errorf("venta válida no debería fallar: %v", err)
	}
}

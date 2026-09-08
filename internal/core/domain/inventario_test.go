package domain

import (
	"math"
	"strconv"
	"testing"
)

// Valores verificados contra migration/goldens/ipv_guardar.json (Fase 0).
func TestCalcularDiferencias(t *testing.T) {
	casos := []struct {
		nombre               string
		in                   InventarioDiario
		wantTeorico, wantDif float64
	}{
		{
			nombre: "aceite: faltante",
			in: InventarioDiario{
				Inicio: 5, Entradas: 2, Consumo: 0.5, Merma: 0.1, OtrasSalidas: 0, FinalFisico: 6.2,
			},
			wantTeorico: 6.4,
			wantDif:     -0.2,
		},
		{
			nombre: "sal: faltante pequeño",
			in: InventarioDiario{
				Inicio: 2, Entradas: 0, Consumo: 0.1, Merma: 0, OtrasSalidas: 0, FinalFisico: 1.85,
			},
			wantTeorico: 1.9,
			wantDif:     -0.05,
		},
		{
			nombre: "ron: cuadra",
			in: InventarioDiario{
				Inicio: 3, Entradas: 1, Consumo: 0.4, Merma: 0, OtrasSalidas: 0, FinalFisico: 3.6,
			},
			wantTeorico: 3.6,
			wantDif:     0,
		},
		{
			nombre: "limon: cuadra con merma",
			in: InventarioDiario{
				Inicio: 20, Entradas: 0, Consumo: 8, Merma: 2, OtrasSalidas: 0, FinalFisico: 10,
			},
			wantTeorico: 10,
			wantDif:     0,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := c.in
			got.CalcularDiferencias()
			if math.Abs(got.FinalTeorico-c.wantTeorico) > 1e-9 {
				t.Errorf("final_teorico = %v, se esperaba %v", got.FinalTeorico, c.wantTeorico)
			}
			if math.Abs(got.Diferencia-c.wantDif) > 1e-9 {
				t.Errorf("diferencia = %v, se esperaba %v", got.Diferencia, c.wantDif)
			}
			if (got.Diferencia < 0) != got.EsFaltante() {
				t.Errorf("EsFaltante() incoherente con diferencia %v", got.Diferencia)
			}
		})
	}
}

// La representación en coma flotante debe coincidir con la de Python para que el
// arnés de paridad de la Fase 5 pueda comparar cuerpos JSON. Ver la nota en
// MIGRATION.md y migration/goldens/ipv_guardar.json.
func TestCalcularDiferencias_ReprCoincideConPython(t *testing.T) {
	casos := []struct {
		in       InventarioDiario
		wantRepr string
	}{
		{
			in:       InventarioDiario{Inicio: 5, Entradas: 2, Consumo: 0.5, Merma: 0.1, FinalFisico: 6.2},
			wantRepr: "-0.20000000000000018",
		},
		{
			in:       InventarioDiario{Inicio: 2, Consumo: 0.1, FinalFisico: 1.85},
			wantRepr: "-0.04999999999999982",
		},
	}
	for _, c := range casos {
		got := c.in
		got.CalcularDiferencias()
		repr := strconv.FormatFloat(got.Diferencia, 'g', -1, 64)
		if repr != c.wantRepr {
			t.Errorf("repr(diferencia) = %s, se esperaba %s (paridad con Python)", repr, c.wantRepr)
		}
	}
}

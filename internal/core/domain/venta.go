package domain

import "strings"

// Venta es una unidad (o varias) de una receta vendida en una fecha.
type Venta struct {
	ID           string
	RecetaNombre string
	Cantidad     int
	Fecha        Date
}

// Validar comprueba las invariantes mínimas.
func (v Venta) Validar() error {
	if strings.TrimSpace(v.RecetaNombre) == "" {
		return &ValidationError{Campo: "receta_nombre", Msg: "el nombre de la receta es obligatorio"}
	}
	if v.Cantidad <= 0 {
		return &ValidationError{Campo: "cantidad", Msg: "la cantidad debe ser mayor que cero"}
	}
	if v.Fecha.IsZero() {
		return &ValidationError{Campo: "fecha", Msg: "la fecha es obligatoria"}
	}
	return nil
}

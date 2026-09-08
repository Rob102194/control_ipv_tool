package domain

import "strings"

// Producto es un artículo de inventario.
type Producto struct {
	ID           string
	Nombre       string
	UnidadMedida string // p. ej. KG, L, U
}

// NormalizarProducto aplica las reglas de saneamiento que la versión Python hacía
// en el caso de uso: nombre y unidad de medida en MAYÚSCULAS y sin espacios
// sobrantes.
func (p *Producto) Normalizar() {
	p.Nombre = strings.ToUpper(strings.TrimSpace(p.Nombre))
	p.UnidadMedida = strings.ToUpper(strings.TrimSpace(p.UnidadMedida))
}

// Validar comprueba las invariantes mínimas.
func (p Producto) Validar() error {
	if strings.TrimSpace(p.Nombre) == "" {
		return &ValidationError{Campo: "nombre", Msg: "el nombre es obligatorio"}
	}
	if strings.TrimSpace(p.UnidadMedida) == "" {
		return &ValidationError{Campo: "unidad_medida", Msg: "la unidad de medida es obligatoria"}
	}
	return nil
}

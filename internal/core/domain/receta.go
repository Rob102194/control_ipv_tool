package domain

import (
	"strconv"
	"strings"
)

// Receta es un plato que se vende. Su lista de ingredientes define cuánto
// producto se descuenta de cada área por unidad vendida.
type Receta struct {
	ID           string
	Nombre       string
	Activa       bool
	Ingredientes []Ingrediente
}

// Ingrediente es una línea de receta: qué producto, de qué área y en qué cantidad.
type Ingrediente struct {
	ID         string
	RecetaID   string
	ProductoID string
	AreaID     string
	Cantidad   float64
}

// Normalizar pone el nombre de la receta en MAYÚSCULAS.
func (r *Receta) Normalizar() {
	r.Nombre = strings.ToUpper(strings.TrimSpace(r.Nombre))
}

// Validar comprueba las invariantes mínimas de la receta y sus ingredientes.
func (r Receta) Validar() error {
	if strings.TrimSpace(r.Nombre) == "" {
		return &ValidationError{Campo: "nombre", Msg: "el nombre es obligatorio"}
	}
	for i, ing := range r.Ingredientes {
		if ing.ProductoID == "" {
			return &ValidationError{Campo: "ingredientes", Msg: "ingrediente sin producto"}
		}
		if ing.AreaID == "" {
			return &ValidationError{Campo: "ingredientes", Msg: "ingrediente sin área"}
		}
		if ing.Cantidad <= 0 {
			return &ValidationError{
				Campo: "ingredientes",
				Msg:   "la cantidad del ingrediente " + strconv.Itoa(i+1) + " debe ser mayor que cero",
			}
		}
	}
	return nil
}

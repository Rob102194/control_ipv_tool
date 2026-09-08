package domain

import "strings"

// Area es una zona del negocio de la que se descuenta inventario (COCINA, BAR…).
type Area struct {
	ID     string
	Nombre string
	Codigo string // opcional; "" significa sin código (se serializa como null)
}

// Normalizar pone el nombre en MAYÚSCULAS (regla heredada de la versión Python).
func (a *Area) Normalizar() {
	a.Nombre = strings.ToUpper(strings.TrimSpace(a.Nombre))
	a.Codigo = strings.TrimSpace(a.Codigo)
}

// Validar comprueba las invariantes mínimas.
func (a Area) Validar() error {
	if strings.TrimSpace(a.Nombre) == "" {
		return &ValidationError{Campo: "nombre", Msg: "el nombre es obligatorio"}
	}
	return nil
}

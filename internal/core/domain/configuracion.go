package domain

import "strings"

// Configuracion agrupa los ajustes del negocio. Cada base de datos pertenece
// a un único negocio (no hay multi-tenant todavía, ver docs/web-roadmap.md),
// así que es una sola fila sin id de negocio.
type Configuracion struct {
	NombreNegocio string
}

// Normalizar recorta espacios sobrantes. No fuerza mayúsculas: es un nombre
// propio (el del restaurante), no un código interno.
func (c *Configuracion) Normalizar() {
	c.NombreNegocio = strings.TrimSpace(c.NombreNegocio)
}

// Validar comprueba las invariantes mínimas. El nombre es opcional (un
// negocio recién instalado aún no lo configuró), pero no puede ser
// desproporcionadamente largo.
func (c Configuracion) Validar() error {
	if len(c.NombreNegocio) > 80 {
		return Invalid("nombre_negocio", "no puede superar los 80 caracteres")
	}
	return nil
}

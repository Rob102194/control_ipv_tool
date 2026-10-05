package domain

import "strings"

// Configuracion agrupa los ajustes del negocio. Cada base de datos pertenece
// a un único negocio (no hay multi-tenant todavía, ver docs/web-roadmap.md),
// así que es una sola fila sin id de negocio.
type Configuracion struct {
	NombreNegocio string
	// BackupDir es la carpeta donde se guardan los backups de la BD; "" usa
	// el valor por defecto (junto a la propia BD). No se valida que exista o
	// sea escribible aquí: eso es un chequeo de filesystem, se hace en el
	// momento de usarla (crear/listar/restaurar un backup), no al guardar
	// esta preferencia.
	BackupDir string
}

// Normalizar recorta espacios sobrantes. No fuerza mayúsculas en el nombre:
// es un nombre propio (el del restaurante), no un código interno.
func (c *Configuracion) Normalizar() {
	c.NombreNegocio = strings.TrimSpace(c.NombreNegocio)
	c.BackupDir = strings.TrimSpace(c.BackupDir)
}

// Validar comprueba las invariantes mínimas. Ambos campos son opcionales (un
// negocio recién instalado aún no los configuró), pero no pueden ser
// desproporcionadamente largos.
func (c Configuracion) Validar() error {
	if len(c.NombreNegocio) > 80 {
		return Invalid("nombre_negocio", "no puede superar los 80 caracteres")
	}
	if len(c.BackupDir) > 500 {
		return Invalid("backup_dir", "la ruta no puede superar los 500 caracteres")
	}
	return nil
}

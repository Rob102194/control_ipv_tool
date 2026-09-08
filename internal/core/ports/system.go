package ports

import (
	"time"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// Clock abstrae el tiempo para poder fijarlo en los tests.
type Clock interface {
	// Now devuelve el instante actual (para marcas de tiempo del historial).
	Now() time.Time
	// Today devuelve la fecha de calendario actual (para importar ventas sin fecha).
	Today() domain.Date
}

// IDGen genera identificadores únicos para las entidades nuevas. La versión
// Python usaba UUID v4 en cadena; la implementación por defecto hará lo mismo.
type IDGen interface {
	New() string
}

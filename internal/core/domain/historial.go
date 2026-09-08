package domain

import "time"

// TipoEntidad identifica a qué clase de entidad pertenece un cambio registrado.
type TipoEntidad string

const (
	EntidadProducto TipoEntidad = "Producto"
	EntidadReceta   TipoEntidad = "Receta"
)

// HistorialCambios es una entrada de auditoría: un campo de una entidad cambió
// de un valor a otro en un instante.
type HistorialCambios struct {
	ID              string
	EntidadTipo     TipoEntidad
	EntidadID       string
	CampoModificado string
	ValorAnterior   string
	ValorNuevo      string
	FechaCambio     time.Time
}

// NuevoCambio construye una entrada de historial. El id y la marca de tiempo los
// asigna la capa de aplicación (con sus puertos IDGen y Clock).
func NuevoCambio(tipo TipoEntidad, entidadID, campo, anterior, nuevo string) HistorialCambios {
	return HistorialCambios{
		EntidadTipo:     tipo,
		EntidadID:       entidadID,
		CampoModificado: campo,
		ValorAnterior:   anterior,
		ValorNuevo:      nuevo,
	}
}

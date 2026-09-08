package usecases

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// HistorialService expone la consulta del historial de cambios.
type HistorialService struct{ Deps }

// Listar devuelve el historial de un tipo de entidad ("Producto", "Receta"),
// de más reciente a más antiguo.
func (s *HistorialService) Listar(ctx context.Context, tipo domain.TipoEntidad) ([]domain.HistorialCambios, error) {
	return s.Repos.Historial().PorEntidad(ctx, tipo)
}

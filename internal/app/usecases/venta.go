package usecases

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// VentaService reúne los casos de uso de ventas (CRUD, sin historial).
type VentaService struct{ Deps }

// VentaInput son los datos de entrada para crear o actualizar una venta.
// Si Fecha es cero, se usa la fecha actual (como la versión Python).
type VentaInput struct {
	RecetaNombre string
	Cantidad     int
	Fecha        domain.Date
}

func (s *VentaService) Listar(ctx context.Context) ([]domain.Venta, error) {
	return s.Repos.Ventas().Listar(ctx)
}

func (s *VentaService) Obtener(ctx context.Context, id string) (domain.Venta, error) {
	return s.Repos.Ventas().ObtenerPorID(ctx, id)
}

func (s *VentaService) Crear(ctx context.Context, in VentaInput) (domain.Venta, error) {
	v := s.aDominio("", in)
	if err := v.Validar(); err != nil {
		return domain.Venta{}, err
	}
	v.ID = s.IDs.New()
	return s.Repos.Ventas().Crear(ctx, v)
}

func (s *VentaService) Actualizar(ctx context.Context, id string, in VentaInput) (domain.Venta, error) {
	v := s.aDominio(id, in)
	if err := v.Validar(); err != nil {
		return domain.Venta{}, err
	}
	return s.Repos.Ventas().Actualizar(ctx, v)
}

func (s *VentaService) Eliminar(ctx context.Context, id string) error {
	return s.Repos.Ventas().Eliminar(ctx, id)
}

func (s *VentaService) EliminarMultiples(ctx context.Context, ids []string) error {
	return s.Repos.Ventas().EliminarMultiples(ctx, ids)
}

func (s *VentaService) aDominio(id string, in VentaInput) domain.Venta {
	fecha := in.Fecha
	if fecha.IsZero() {
		fecha = s.Clock.Today()
	}
	return domain.Venta{
		ID:           id,
		RecetaNombre: in.RecetaNombre,
		Cantidad:     in.Cantidad,
		Fecha:        fecha,
	}
}

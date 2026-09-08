package usecases

import (
	"context"
	"errors"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// AreaService reúne los casos de uso de áreas. La versión Python no registra
// historial para áreas.
type AreaService struct{ Deps }

// AreaInput son los datos de entrada para crear o actualizar un área.
type AreaInput struct {
	Nombre string
	Codigo string
}

func (s *AreaService) Listar(ctx context.Context) ([]domain.Area, error) {
	return s.Repos.Areas().Listar(ctx)
}

func (s *AreaService) Obtener(ctx context.Context, id string) (domain.Area, error) {
	return s.Repos.Areas().ObtenerPorID(ctx, id)
}

func (s *AreaService) Crear(ctx context.Context, in AreaInput) (domain.Area, error) {
	a := domain.Area{Nombre: in.Nombre, Codigo: in.Codigo}
	a.Normalizar()
	if err := a.Validar(); err != nil {
		return domain.Area{}, err
	}
	if _, err := s.Repos.Areas().BuscarPorNombre(ctx, a.Nombre); err == nil {
		return domain.Area{}, domain.Conflictf("El área con el nombre '%s' ya existe.", a.Nombre)
	} else if !errors.Is(err, ports.ErrNoEncontrado) {
		return domain.Area{}, err
	}
	a.ID = s.IDs.New()
	return s.Repos.Areas().Crear(ctx, a)
}

func (s *AreaService) Actualizar(ctx context.Context, id string, in AreaInput) (domain.Area, error) {
	a := domain.Area{ID: id, Nombre: in.Nombre, Codigo: in.Codigo}
	a.Normalizar()
	if err := a.Validar(); err != nil {
		return domain.Area{}, err
	}
	return s.Repos.Areas().Actualizar(ctx, a)
}

func (s *AreaService) Eliminar(ctx context.Context, id string) error {
	return s.Repos.Areas().Eliminar(ctx, id)
}

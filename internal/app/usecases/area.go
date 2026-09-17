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

	var creada domain.Area
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		if _, err := r.Areas().BuscarPorNombre(ctx, a.Nombre); err == nil {
			return domain.Conflictf("El área con el nombre '%s' ya existe.", a.Nombre)
		} else if !errors.Is(err, ports.ErrNoEncontrado) {
			return err
		}
		a.ID = s.IDs.New()
		var err error
		creada, err = r.Areas().Crear(ctx, a)
		return err
	})
	return creada, err
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
	return s.UOW.Do(ctx, func(r ports.Repos) error {
		enUso, err := r.Areas().EnUso(ctx, id)
		if err != nil {
			return err
		}
		if enUso {
			return domain.Conflictf("El área no se puede eliminar porque está siendo utilizada en una o más recetas, inventarios o modelos.")
		}
		return r.Areas().Eliminar(ctx, id)
	})
}

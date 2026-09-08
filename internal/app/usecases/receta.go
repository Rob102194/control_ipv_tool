package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// RecetaService reúne los casos de uso de recetas.
type RecetaService struct{ Deps }

// RecetaInput son los datos de entrada para crear o actualizar una receta.
type RecetaInput struct {
	Nombre       string
	Activa       bool
	Ingredientes []IngredienteInput
}

// IngredienteInput es una línea de receta de entrada (sin id).
type IngredienteInput struct {
	ProductoID string
	AreaID     string
	Cantidad   float64
}

func (s *RecetaService) Listar(ctx context.Context, opts ports.ListarRecetasOpts) ([]domain.Receta, error) {
	return s.Repos.Recetas().Listar(ctx, opts)
}

func (s *RecetaService) Obtener(ctx context.Context, id string) (domain.Receta, error) {
	return s.Repos.Recetas().ObtenerPorID(ctx, id)
}

// Crear normaliza el nombre a MAYÚSCULAS, rechaza duplicados y registra el
// cambio. Genera los ids de la receta y de sus ingredientes.
func (s *RecetaService) Crear(ctx context.Context, in RecetaInput) (domain.Receta, error) {
	rec := s.aDominio("", in)
	rec.Normalizar()
	if err := rec.Validar(); err != nil {
		return domain.Receta{}, err
	}

	var creada domain.Receta
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		if _, err := r.Recetas().BuscarPorNombre(ctx, rec.Nombre); err == nil {
			return domain.Conflictf("La receta con el nombre '%s' ya existe.", rec.Nombre)
		} else if !errors.Is(err, ports.ErrNoEncontrado) {
			return err
		}
		rec.ID = s.IDs.New()
		for i := range rec.Ingredientes {
			rec.Ingredientes[i].ID = s.IDs.New()
			rec.Ingredientes[i].RecetaID = rec.ID
		}
		var err error
		if creada, err = r.Recetas().Crear(ctx, rec); err != nil {
			return err
		}
		return s.registrarCambio(ctx, r, domain.EntidadReceta, rec.ID, "Creación", "",
			fmt.Sprintf("Receta '%s' creada", rec.Nombre))
	})
	return creada, err
}

// Actualizar reemplaza la receta y sus ingredientes. Registra el cambio de
// nombre y el cambio en el número de ingredientes (misma lógica que Python).
func (s *RecetaService) Actualizar(ctx context.Context, id string, in RecetaInput) (domain.Receta, error) {
	nueva := s.aDominio(id, in)
	nueva.Normalizar()
	if err := nueva.Validar(); err != nil {
		return domain.Receta{}, err
	}

	var actualizada domain.Receta
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		actual, err := r.Recetas().ObtenerPorID(ctx, id)
		if err != nil {
			return err
		}
		for i := range nueva.Ingredientes {
			nueva.Ingredientes[i].ID = s.IDs.New()
			nueva.Ingredientes[i].RecetaID = id
		}
		if actual.Nombre != nueva.Nombre {
			if err := s.registrarCambio(ctx, r, domain.EntidadReceta, id, "nombre", actual.Nombre, nueva.Nombre); err != nil {
				return err
			}
		}
		if len(actual.Ingredientes) != len(nueva.Ingredientes) {
			if err := s.registrarCambio(ctx, r, domain.EntidadReceta, id, "ingredientes",
				fmt.Sprintf("Receta '%s' tenía %d ingredientes", actual.Nombre, len(actual.Ingredientes)),
				fmt.Sprintf("Receta '%s' ahora tiene %d ingredientes", nueva.Nombre, len(nueva.Ingredientes)),
			); err != nil {
				return err
			}
		}
		actualizada, err = r.Recetas().Actualizar(ctx, nueva)
		return err
	})
	return actualizada, err
}

// Eliminar comprueba que exista y registra el cambio.
func (s *RecetaService) Eliminar(ctx context.Context, id string) error {
	return s.UOW.Do(ctx, func(r ports.Repos) error {
		actual, err := r.Recetas().ObtenerPorID(ctx, id)
		if err != nil {
			return err
		}
		if err := s.registrarCambio(ctx, r, domain.EntidadReceta, id, "Eliminación",
			fmt.Sprintf("Receta '%s' eliminada", actual.Nombre), ""); err != nil {
			return err
		}
		return r.Recetas().Eliminar(ctx, id)
	})
}

func (s *RecetaService) aDominio(id string, in RecetaInput) domain.Receta {
	rec := domain.Receta{ID: id, Nombre: in.Nombre, Activa: in.Activa}
	for _, ing := range in.Ingredientes {
		rec.Ingredientes = append(rec.Ingredientes, domain.Ingrediente{
			ProductoID: ing.ProductoID,
			AreaID:     ing.AreaID,
			Cantidad:   ing.Cantidad,
		})
	}
	return rec
}

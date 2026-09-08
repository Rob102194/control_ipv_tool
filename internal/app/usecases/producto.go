package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// ProductoService reúne los casos de uso de productos.
type ProductoService struct{ Deps }

// ProductoInput son los datos de entrada para crear o actualizar un producto.
type ProductoInput struct {
	Nombre       string
	UnidadMedida string
}

// Listar devuelve los productos en el orden indicado.
func (s *ProductoService) Listar(ctx context.Context, orden ports.OrdenProducto) ([]domain.Producto, error) {
	return s.Repos.Productos().Listar(ctx, orden)
}

// Obtener devuelve un producto por id (ports.ErrNoEncontrado si no existe).
func (s *ProductoService) Obtener(ctx context.Context, id string) (domain.Producto, error) {
	return s.Repos.Productos().ObtenerPorID(ctx, id)
}

// Crear normaliza, comprueba que el nombre no exista y registra el cambio.
func (s *ProductoService) Crear(ctx context.Context, in ProductoInput) (domain.Producto, error) {
	p := domain.Producto{Nombre: in.Nombre, UnidadMedida: in.UnidadMedida}
	p.Normalizar()
	if err := p.Validar(); err != nil {
		return domain.Producto{}, err
	}

	var creado domain.Producto
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		if _, err := r.Productos().BuscarPorNombre(ctx, p.Nombre); err == nil {
			return domain.Conflictf("El producto con el nombre '%s' ya existe.", p.Nombre)
		} else if !errors.Is(err, ports.ErrNoEncontrado) {
			return err
		}
		p.ID = s.IDs.New()
		var err error
		if creado, err = r.Productos().Crear(ctx, p); err != nil {
			return err
		}
		return s.registrarCambio(ctx, r, domain.EntidadProducto, p.ID, "Creación", "",
			fmt.Sprintf("Producto '%s' creado", p.Nombre))
	})
	return creado, err
}

// Actualizar normaliza, comprueba que exista y registra los campos cambiados.
func (s *ProductoService) Actualizar(ctx context.Context, id string, in ProductoInput) (domain.Producto, error) {
	nuevo := domain.Producto{ID: id, Nombre: in.Nombre, UnidadMedida: in.UnidadMedida}
	nuevo.Normalizar()
	if err := nuevo.Validar(); err != nil {
		return domain.Producto{}, err
	}

	var actualizado domain.Producto
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		actual, err := r.Productos().ObtenerPorID(ctx, id)
		if err != nil {
			return err
		}
		if actual.Nombre != nuevo.Nombre {
			if err := s.registrarCambio(ctx, r, domain.EntidadProducto, id, "nombre", actual.Nombre, nuevo.Nombre); err != nil {
				return err
			}
		}
		if actual.UnidadMedida != nuevo.UnidadMedida {
			if err := s.registrarCambio(ctx, r, domain.EntidadProducto, id, "unidad_medida", actual.UnidadMedida, nuevo.UnidadMedida); err != nil {
				return err
			}
		}
		actualizado, err = r.Productos().Actualizar(ctx, nuevo)
		return err
	})
	return actualizado, err
}

// Eliminar comprueba que exista, que no esté en uso, y registra el cambio.
func (s *ProductoService) Eliminar(ctx context.Context, id string) error {
	return s.UOW.Do(ctx, func(r ports.Repos) error {
		actual, err := r.Productos().ObtenerPorID(ctx, id)
		if err != nil {
			return err
		}
		enUso, err := r.Productos().EnUso(ctx, id)
		if err != nil {
			return err
		}
		if enUso {
			return domain.Conflictf("El producto no se puede eliminar porque está siendo utilizado en una o más recetas, inventarios o modelos.")
		}
		if err := s.registrarCambio(ctx, r, domain.EntidadProducto, id, "Eliminación",
			fmt.Sprintf("Producto '%s' eliminado", actual.Nombre), ""); err != nil {
			return err
		}
		return r.Productos().Eliminar(ctx, id)
	})
}

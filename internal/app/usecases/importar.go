package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// ImportProductoItem es una fila de importación de productos.
type ImportProductoItem struct {
	Nombre       string
	UnidadMedida string
}

// ImportRecetaItem es una fila de importación de recetas (una por ingrediente;
// sin ingredientes -> ProductoNombre == "").
type ImportRecetaItem struct {
	RecetaNombre   string
	ProductoNombre string
	UnidadMedida   string
	AreaNombre     string
	Cantidad       float64
}

// ImportVentaItem es una fila de importación de ventas ya parseada.
type ImportVentaItem struct {
	Fila           int
	Nombre         string
	Cantidad       float64
	CantidadValida bool
	RawCantidad    string
}

// Importar crea los productos cuyo nombre no exista. NO normaliza los nombres y
// NO registra historial (igual que ImportProductosExcel en Python).
func (s *ProductoService) Importar(ctx context.Context, items []ImportProductoItem) (ImportResult, error) {
	var res ImportResult
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		for _, it := range items {
			if strings.TrimSpace(it.Nombre) == "" {
				res.Omitidos++
				continue
			}
			yaExiste, err := existe(r.Productos().BuscarPorNombre(ctx, it.Nombre))
			if err != nil {
				return err
			}
			if yaExiste {
				res.Omitidos++
				continue
			}
			if _, err := r.Productos().Crear(ctx, domain.Producto{
				ID: s.IDs.New(), Nombre: it.Nombre, UnidadMedida: it.UnidadMedida,
			}); err != nil {
				return err
			}
			res.Creados++
		}
		return nil
	})
	return res, err
}

// Importar agrupa las filas por receta, crea los productos y áreas que falten
// (por nombre, sin normalizar) y crea las recetas que no existan. Omite las
// recetas ya existentes. Misma lógica que ImportRecetasExcel en Python.
func (s *RecetaService) Importar(ctx context.Context, rows []ImportRecetaItem) (ImportResult, error) {
	var res ImportResult
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		prodCache := map[string]string{}
		areaCache := map[string]string{}

		type acumulada struct {
			nombre string
			ings   []domain.Ingrediente
			skip   bool
		}
		recetas := map[string]*acumulada{}
		var orden []string

		for _, row := range rows {
			rn := strings.TrimSpace(row.RecetaNombre)
			if rn == "" {
				continue
			}
			a, vista := recetas[rn]
			if !vista {
				yaExiste, err := existe(r.Recetas().BuscarPorNombre(ctx, rn))
				if err != nil {
					return err
				}
				a = &acumulada{nombre: rn, skip: yaExiste}
				recetas[rn] = a
				if yaExiste {
					res.Omitidos++
				} else {
					orden = append(orden, rn)
				}
			}
			if a.skip || strings.TrimSpace(row.ProductoNombre) == "" {
				continue
			}
			prodID, err := s.resolverProducto(ctx, r, prodCache, row.ProductoNombre, row.UnidadMedida)
			if err != nil {
				return err
			}
			areaID, err := s.resolverArea(ctx, r, areaCache, row.AreaNombre)
			if err != nil {
				return err
			}
			a.ings = append(a.ings, domain.Ingrediente{ProductoID: prodID, AreaID: areaID, Cantidad: row.Cantidad})
		}

		for _, rn := range orden {
			a := recetas[rn]
			rec := domain.Receta{ID: s.IDs.New(), Nombre: a.nombre, Activa: true, Ingredientes: a.ings}
			for i := range rec.Ingredientes {
				rec.Ingredientes[i].ID = s.IDs.New()
				rec.Ingredientes[i].RecetaID = rec.ID
			}
			if _, err := r.Recetas().Crear(ctx, rec); err != nil {
				return err
			}
			res.Creados++
		}
		return nil
	})
	return res, err
}

func (s *RecetaService) resolverProducto(ctx context.Context, r ports.Repos, cache map[string]string, nombre, unidad string) (string, error) {
	if id, ok := cache[nombre]; ok {
		return id, nil
	}
	p, err := r.Productos().BuscarPorNombre(ctx, nombre)
	if err == nil {
		cache[nombre] = p.ID
		return p.ID, nil
	}
	if !errors.Is(err, ports.ErrNoEncontrado) {
		return "", err
	}
	nuevo, err := r.Productos().Crear(ctx, domain.Producto{ID: s.IDs.New(), Nombre: nombre, UnidadMedida: unidad})
	if err != nil {
		return "", err
	}
	cache[nombre] = nuevo.ID
	return nuevo.ID, nil
}

func (s *RecetaService) resolverArea(ctx context.Context, r ports.Repos, cache map[string]string, nombre string) (string, error) {
	if id, ok := cache[nombre]; ok {
		return id, nil
	}
	a, err := r.Areas().BuscarPorNombre(ctx, nombre)
	if err == nil {
		cache[nombre] = a.ID
		return a.ID, nil
	}
	if !errors.Is(err, ports.ErrNoEncontrado) {
		return "", err
	}
	nueva, err := r.Areas().Crear(ctx, domain.Area{ID: s.IDs.New(), Nombre: nombre})
	if err != nil {
		return "", err
	}
	cache[nombre] = nueva.ID
	return nueva.ID, nil
}

// Importar crea automáticamente las recetas cuyo nombre no exista (sin
// ingredientes) y luego inserta las ventas. Si alguna fila tiene una cantidad
// inválida, no inserta nada y devuelve un ValidationError con todas las líneas
// de error (mismo texto que la versión Python).
func (s *VentaService) Importar(ctx context.Context, rows []ImportVentaItem, fecha *domain.Date) (ImportVentasResult, error) {
	f := s.Clock.Today()
	if fecha != nil && !fecha.IsZero() {
		f = *fecha
	}

	var result ImportVentasResult
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		existentes, err := r.Recetas().Listar(ctx, ports.ListarRecetasOpts{})
		if err != nil {
			return err
		}
		enDB := make(map[string]bool, len(existentes))
		for _, rec := range existentes {
			enDB[rec.Nombre] = true
		}

		var nuevas []domain.Receta
		vistas := map[string]bool{}
		for _, row := range rows {
			if enDB[row.Nombre] || vistas[row.Nombre] {
				continue
			}
			vistas[row.Nombre] = true
			nuevas = append(nuevas, domain.Receta{ID: s.IDs.New(), Nombre: row.Nombre, Activa: true})
		}
		if len(nuevas) > 0 {
			if _, err := r.Recetas().CrearMultiples(ctx, nuevas); err != nil {
				return err
			}
		}
		result.NuevasRecetas = nuevas

		var errsFila []string
		var aCrear []domain.Venta
		for _, row := range rows {
			if !row.CantidadValida {
				errsFila = append(errsFila, fmt.Sprintf(
					"Fila %d: Cantidad inválida (%s) para la receta '%s'", row.Fila, row.RawCantidad, row.Nombre))
				continue
			}
			aCrear = append(aCrear, domain.Venta{
				ID: s.IDs.New(), RecetaNombre: row.Nombre, Cantidad: int(row.Cantidad), Fecha: f,
			})
		}
		if len(errsFila) > 0 {
			return &domain.ValidationError{Msg: strings.Join(errsFila, "\n")}
		}
		if len(aCrear) > 0 {
			if _, err := r.Ventas().CrearMultiples(ctx, aCrear); err != nil {
				return err
			}
		}
		result.Ventas = aCrear
		return nil
	})
	if err != nil {
		return ImportVentasResult{}, err
	}
	return result, nil
}

// existe traduce el resultado de una búsqueda por identidad a un booleano:
// nil -> true, ports.ErrNoEncontrado -> false, otro error -> se propaga.
// Se llama pasándole directamente el (valor, error) de la búsqueda.
func existe[T any](_ T, err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ports.ErrNoEncontrado) {
		return false, nil
	}
	return false, err
}

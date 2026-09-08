package usecases

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// IPVService reúne los casos de uso del inventario diario (IPV).
type IPVService struct{ Deps }

// ObtenerEstado devuelve el estado del IPV de una fecha, indexado por nombre de
// área. Si no hay registros guardados, construye una plantilla a partir del
// modelo de cada área, con `inicio` arrastrado del final físico del día anterior.
func (s *IPVService) ObtenerEstado(ctx context.Context, fecha domain.Date) (map[string][]InventarioFilaView, error) {
	areas, err := s.Repos.Areas().Listar(ctx)
	if err != nil {
		return nil, err
	}
	existentes, err := s.Repos.InventarioDiario().ListarPorFecha(ctx, fecha)
	if err != nil {
		return nil, err
	}
	productos, err := s.Repos.Productos().Listar(ctx, ports.OrdenProductoNombre)
	if err != nil {
		return nil, err
	}
	prodByID := make(map[string]domain.Producto, len(productos))
	for _, p := range productos {
		prodByID[p.ID] = p
	}
	areaByID := make(map[string]domain.Area, len(areas))
	for _, a := range areas {
		areaByID[a.ID] = a
	}

	result := make(map[string][]InventarioFilaView, len(areas))
	for _, a := range areas {
		result[a.Nombre] = []InventarioFilaView{}
	}

	if len(existentes) == 0 {
		modelo, err := s.Repos.ModelosIPV().Listar(ctx)
		if err != nil {
			return nil, err
		}
		modeloByArea := make(map[string][]domain.ModeloIPV)
		for _, m := range modelo {
			modeloByArea[m.AreaID] = append(modeloByArea[m.AreaID], m)
		}
		for _, a := range areas {
			for _, m := range modeloByArea[a.ID] {
				p, ok := prodByID[m.ProductoID]
				if !ok {
					continue
				}
				inicio, err := s.Repos.InventarioDiario().FinalFisicoDiaAnterior(ctx, fecha, a.ID, p.ID)
				if err != nil {
					return nil, err
				}
				result[a.Nombre] = append(result[a.Nombre], InventarioFilaView{
					ID: s.IDs.New(), Fecha: fecha, AreaID: a.ID, ProductoID: p.ID,
					Inicio: inicio, ProductoNombre: p.Nombre, AreaNombre: a.Nombre, Comentario: "",
				})
			}
		}
		return result, nil
	}

	for _, iv := range existentes {
		a, ok := areaByID[iv.AreaID]
		if !ok {
			continue
		}
		nombreProd := "Producto no encontrado"
		if p, ok := prodByID[iv.ProductoID]; ok {
			nombreProd = p.Nombre
		}
		result[a.Nombre] = append(result[a.Nombre], viewFromDomain(iv, nombreProd, a.Nombre))
	}
	return result, nil
}

// CalcularConsumo reparte las ventas del día en consumo por producto y área.
// Devuelve un mapa con claves "producto_id|area_id" (formato de la versión Python).
func (s *IPVService) CalcularConsumo(ctx context.Context, fecha domain.Date) (map[string]float64, error) {
	ventas, err := s.Repos.Ventas().ListarPorFecha(ctx, fecha)
	if err != nil {
		return nil, err
	}
	recetas, err := s.Repos.Recetas().Listar(ctx, ports.ListarRecetasOpts{})
	if err != nil {
		return nil, err
	}
	porNombre := make(map[string]domain.Receta, len(recetas))
	for _, r := range recetas {
		porNombre[r.Nombre] = r
	}
	consumo := domain.CalcularConsumo(ventas, porNombre)
	out := make(map[string]float64, len(consumo))
	for k, v := range consumo {
		out[k.ProductoID+"|"+k.AreaID] = v
	}
	return out, nil
}

// Recalcular recalcula final_teorico y diferencia de cada fila, sin persistir.
// Es el caso de uso del endpoint /ipv/calcular (nuevo en la versión Go).
func (s *IPVService) Recalcular(filas []InventarioFilaView) []InventarioFilaView {
	for i := range filas {
		d := filas[i].toDomain()
		d.CalcularDiferencias()
		filas[i].FinalTeorico = d.FinalTeorico
		filas[i].Diferencia = d.Diferencia
	}
	return filas
}

// Guardar hace upsert del IPV de un día. Recalcula final_teorico y diferencia en
// el servidor; los valores enviados en esos campos se ignoran.
func (s *IPVService) Guardar(ctx context.Context, filas []InventarioFilaView) ([]InventarioFilaView, error) {
	dominio := make([]domain.InventarioDiario, len(filas))
	for i := range filas {
		d := filas[i].toDomain()
		if d.ID == "" {
			d.ID = s.IDs.New()
		}
		d.CalcularDiferencias()
		dominio[i] = d
		filas[i].ID = d.ID
		filas[i].FinalTeorico = d.FinalTeorico
		filas[i].Diferencia = d.Diferencia
	}
	err := s.UOW.Do(ctx, func(r ports.Repos) error {
		return r.InventarioDiario().GuardarTodos(ctx, dominio)
	})
	if err != nil {
		return nil, err
	}
	return filas, nil
}

// ObtenerModelos devuelve los modelos de IPV indexados por area_id.
func (s *IPVService) ObtenerModelos(ctx context.Context) (map[string][]ModeloItem, error) {
	rows, err := s.Repos.ModelosIPV().Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]ModeloItem)
	for _, m := range rows {
		out[m.AreaID] = append(out[m.AreaID], ModeloItem{ProductoID: m.ProductoID, Orden: m.Orden})
	}
	return out, nil
}

// GuardarModelo reemplaza por completo el modelo de IPV de un área.
func (s *IPVService) GuardarModelo(ctx context.Context, areaID string, items []ModeloItem) error {
	filas := make([]domain.ModeloIPV, len(items))
	for i, it := range items {
		filas[i] = domain.ModeloIPV{
			ID: s.IDs.New(), AreaID: areaID, ProductoID: it.ProductoID, Orden: it.Orden,
		}
	}
	return s.UOW.Do(ctx, func(r ports.Repos) error {
		return r.ModelosIPV().GuardarModelo(ctx, areaID, filas)
	})
}

// FechasConRegistros devuelve las fechas con IPV guardado, de más reciente a más
// antigua.
func (s *IPVService) FechasConRegistros(ctx context.Context) ([]domain.Date, error) {
	return s.Repos.InventarioDiario().FechasConRegistros(ctx)
}

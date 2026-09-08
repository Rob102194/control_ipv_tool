package httpapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// Este fichero define los DTO de entrada/salida y su mapeo desde/hacia el
// dominio y los servicios. La serialización vive AQUÍ, no en el dominio.
// Las claves JSON replican el to_dict de la versión Python.

// --- salida ------------------------------------------------------------

type productoDTO struct {
	ID           string `json:"id"`
	Nombre       string `json:"nombre"`
	UnidadMedida string `json:"unidad_medida"`
}

func toProductoDTO(p domain.Producto) productoDTO {
	return productoDTO{ID: p.ID, Nombre: p.Nombre, UnidadMedida: p.UnidadMedida}
}

type areaDTO struct {
	ID     string  `json:"id"`
	Nombre string  `json:"nombre"`
	Codigo *string `json:"codigo"`
}

func toAreaDTO(a domain.Area) areaDTO {
	var cod *string
	if a.Codigo != "" {
		c := a.Codigo
		cod = &c
	}
	return areaDTO{ID: a.ID, Nombre: a.Nombre, Codigo: cod}
}

type ingredienteDTO struct {
	ID         string  `json:"id"`
	ProductoID string  `json:"producto_id"`
	AreaID     string  `json:"area_id"`
	Cantidad   float64 `json:"cantidad"`
}

type recetaDTO struct {
	ID           string           `json:"id"`
	Nombre       string           `json:"nombre"`
	Activa       bool             `json:"activa"`
	Ingredientes []ingredienteDTO `json:"ingredientes"`
}

func toRecetaDTO(r domain.Receta) recetaDTO {
	ings := make([]ingredienteDTO, 0, len(r.Ingredientes))
	for _, i := range r.Ingredientes {
		ings = append(ings, ingredienteDTO{ID: i.ID, ProductoID: i.ProductoID, AreaID: i.AreaID, Cantidad: i.Cantidad})
	}
	return recetaDTO{ID: r.ID, Nombre: r.Nombre, Activa: r.Activa, Ingredientes: ings}
}

type ventaDTO struct {
	ID           string `json:"id"`
	RecetaNombre string `json:"receta_nombre"`
	Cantidad     int    `json:"cantidad"`
	Fecha        string `json:"fecha"`
}

func toVentaDTO(v domain.Venta) ventaDTO {
	return ventaDTO{ID: v.ID, RecetaNombre: v.RecetaNombre, Cantidad: v.Cantidad, Fecha: v.Fecha.String()}
}

type inventarioDTO struct {
	ID             string  `json:"id"`
	Fecha          string  `json:"fecha"`
	AreaID         string  `json:"area_id"`
	ProductoID     string  `json:"producto_id"`
	Inicio         float64 `json:"inicio"`
	Entradas       float64 `json:"entradas"`
	Consumo        float64 `json:"consumo"`
	Merma          float64 `json:"merma"`
	OtrasSalidas   float64 `json:"otras_salidas"`
	FinalFisico    float64 `json:"final_fisico"`
	FinalTeorico   float64 `json:"final_teorico"`
	Diferencia     float64 `json:"diferencia"`
	ProductoNombre string  `json:"producto_nombre"`
	AreaNombre     string  `json:"area_nombre"`
	Comentario     string  `json:"comentario"`
}

func toInventarioDTO(v usecases.InventarioFilaView) inventarioDTO {
	return inventarioDTO{
		ID: v.ID, Fecha: v.Fecha.String(), AreaID: v.AreaID, ProductoID: v.ProductoID,
		Inicio: v.Inicio, Entradas: v.Entradas, Consumo: v.Consumo, Merma: v.Merma,
		OtrasSalidas: v.OtrasSalidas, FinalFisico: v.FinalFisico,
		FinalTeorico: v.FinalTeorico, Diferencia: v.Diferencia,
		ProductoNombre: v.ProductoNombre, AreaNombre: v.AreaNombre, Comentario: v.Comentario,
	}
}

func toInventarioDTOs(vs []usecases.InventarioFilaView) []inventarioDTO {
	out := make([]inventarioDTO, len(vs))
	for i, v := range vs {
		out[i] = toInventarioDTO(v)
	}
	return out
}

type historialDTO struct {
	ID              string  `json:"id"`
	EntidadTipo     string  `json:"entidad_tipo"`
	EntidadID       string  `json:"entidad_id"`
	CampoModificado string  `json:"campo_modificado"`
	ValorAnterior   *string `json:"valor_anterior"`
	ValorNuevo      *string `json:"valor_nuevo"`
	FechaCambio     *string `json:"fecha_cambio"`
}

func toHistorialDTO(h domain.HistorialCambios) historialDTO {
	d := historialDTO{
		ID: h.ID, EntidadTipo: string(h.EntidadTipo), EntidadID: h.EntidadID,
		CampoModificado: h.CampoModificado,
		ValorAnterior:   nilIfEmpty(h.ValorAnterior),
		ValorNuevo:      nilIfEmpty(h.ValorNuevo),
	}
	if !h.FechaCambio.IsZero() {
		s := h.FechaCambio.UTC().Format("2006-01-02T15:04:05")
		d.FechaCambio = &s
	}
	return d
}

// --- reporte ---------------------------------------------------------

type reporteDTO struct {
	Fecha   string                       `json:"fecha"`
	Areas   map[string][]reporteFilaDTO  `json:"areas"`
	Resumen map[string]reporteResumenDTO `json:"resumen"`
	Notas   []string                     `json:"notas"`
}

type reporteFilaDTO struct {
	Producto     string  `json:"producto"`
	UM           string  `json:"um"`
	Inicio       float64 `json:"inicio"`
	Entradas     float64 `json:"entradas"`
	Consumo      float64 `json:"consumo"`
	Merma        float64 `json:"merma"`
	OtrasSalidas float64 `json:"otras_salidas"`
	FinalTeorico float64 `json:"final_teorico"`
	FinalFisico  float64 `json:"final_fisico"`
	Diferencia   float64 `json:"diferencia"`
}

type reporteResumenDTO struct {
	Faltantes []string `json:"faltantes"`
	Sobrantes []string `json:"sobrantes"`
	Mermas    []string `json:"mermas"`
}

func toReporteDTO(r usecases.ReporteIPV) reporteDTO {
	out := reporteDTO{
		Fecha:   r.Fecha.String(),
		Areas:   map[string][]reporteFilaDTO{},
		Resumen: map[string]reporteResumenDTO{},
		Notas:   []string{},
	}
	for area, filas := range r.Areas {
		dtos := make([]reporteFilaDTO, len(filas))
		for i, f := range filas {
			dtos[i] = reporteFilaDTO{
				Producto: f.Producto, UM: f.UM,
				Inicio: f.Inicio, Entradas: f.Entradas, Consumo: f.Consumo,
				Merma: f.Merma, OtrasSalidas: f.OtrasSalidas,
				FinalTeorico: f.FinalTeorico, FinalFisico: f.FinalFisico, Diferencia: f.Diferencia,
			}
		}
		out.Areas[area] = dtos
	}
	for area, res := range r.Resumen {
		out.Resumen[area] = reporteResumenDTO{
			Faltantes: deltasAStrings(res.Faltantes),
			Sobrantes: deltasAStrings(res.Sobrantes),
			Mermas:    deltasAStrings(res.Mermas),
		}
	}
	for _, n := range r.Notas {
		if n.Campo != "" {
			out.Notas = append(out.Notas, fmt.Sprintf("%s (%s): %s", n.Producto, n.Campo, n.Texto))
		} else {
			out.Notas = append(out.Notas, fmt.Sprintf("%s: %s", n.Producto, n.Texto))
		}
	}
	return out
}

func deltasAStrings(ds []usecases.ReporteDelta) []string {
	out := make([]string, 0, len(ds))
	for _, d := range ds {
		out = append(out, fmt.Sprintf("%s: %s %s", d.Producto, pyFloat(d.Cantidad), d.UM))
	}
	return out
}

// pyFloat formatea un float como lo haría repr()/str() de Python: dígitos
// mínimos que preservan el valor, pero los enteros llevan ".0" ("2.0", no "2").
func pyFloat(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eEnN") {
		s += ".0"
	}
	return s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- entrada ---------------------------------------------------------

type productoInputDTO struct {
	Nombre       string `json:"nombre" validate:"required"`
	UnidadMedida string `json:"unidad_medida" validate:"required"`
}

func (d productoInputDTO) toUC() usecases.ProductoInput {
	return usecases.ProductoInput{Nombre: d.Nombre, UnidadMedida: d.UnidadMedida}
}

type areaInputDTO struct {
	Nombre string `json:"nombre" validate:"required"`
	Codigo string `json:"codigo"`
}

func (d areaInputDTO) toUC() usecases.AreaInput {
	return usecases.AreaInput{Nombre: d.Nombre, Codigo: d.Codigo}
}

type recetaInputDTO struct {
	Nombre       string `json:"nombre" validate:"required"`
	Activa       *bool  `json:"activa"` // por defecto true (como Python)
	Ingredientes []struct {
		ProductoID string  `json:"producto_id" validate:"required"`
		AreaID     string  `json:"area_id" validate:"required"`
		Cantidad   float64 `json:"cantidad"`
	} `json:"ingredientes"`
}

func (d recetaInputDTO) toUC() usecases.RecetaInput {
	activa := true
	if d.Activa != nil {
		activa = *d.Activa
	}
	in := usecases.RecetaInput{Nombre: d.Nombre, Activa: activa}
	for _, i := range d.Ingredientes {
		in.Ingredientes = append(in.Ingredientes, usecases.IngredienteInput{
			ProductoID: i.ProductoID, AreaID: i.AreaID, Cantidad: i.Cantidad,
		})
	}
	return in
}

type ventaInputDTO struct {
	RecetaNombre string `json:"receta_nombre" validate:"required"`
	Cantidad     int    `json:"cantidad" validate:"required"`
	Fecha        string `json:"fecha"`
}

func (d ventaInputDTO) toUC() (usecases.VentaInput, error) {
	in := usecases.VentaInput{RecetaNombre: d.RecetaNombre, Cantidad: d.Cantidad}
	if d.Fecha != "" {
		f, err := domain.ParseDate(d.Fecha)
		if err != nil {
			return usecases.VentaInput{}, domain.Invalid("fecha", err.Error())
		}
		in.Fecha = f
	}
	return in, nil
}

// inventarioInputDTO es a la vez entrada de /ipv/guardar y /ipv/calcular, y la
// forma en que se serializan las filas del estado.
type inventarioInputDTO struct {
	ID             string  `json:"id"`
	Fecha          string  `json:"fecha"`
	AreaID         string  `json:"area_id"`
	ProductoID     string  `json:"producto_id"`
	Inicio         float64 `json:"inicio"`
	Entradas       float64 `json:"entradas"`
	Consumo        float64 `json:"consumo"`
	Merma          float64 `json:"merma"`
	OtrasSalidas   float64 `json:"otras_salidas"`
	FinalFisico    float64 `json:"final_fisico"`
	FinalTeorico   float64 `json:"final_teorico"`
	Diferencia     float64 `json:"diferencia"`
	ProductoNombre string  `json:"producto_nombre"`
	AreaNombre     string  `json:"area_nombre"`
	Comentario     string  `json:"comentario"`
}

func (d inventarioInputDTO) toView() (usecases.InventarioFilaView, error) {
	var fecha domain.Date
	if d.Fecha != "" {
		f, err := domain.ParseDate(d.Fecha)
		if err != nil {
			return usecases.InventarioFilaView{}, domain.Invalid("fecha", err.Error())
		}
		fecha = f
	}
	return usecases.InventarioFilaView{
		ID: d.ID, Fecha: fecha, AreaID: d.AreaID, ProductoID: d.ProductoID,
		Inicio: d.Inicio, Entradas: d.Entradas, Consumo: d.Consumo, Merma: d.Merma,
		OtrasSalidas: d.OtrasSalidas, FinalFisico: d.FinalFisico,
		FinalTeorico: d.FinalTeorico, Diferencia: d.Diferencia,
		ProductoNombre: d.ProductoNombre, AreaNombre: d.AreaNombre, Comentario: d.Comentario,
	}, nil
}

func viewsFromInput(ds []inventarioInputDTO) ([]usecases.InventarioFilaView, error) {
	out := make([]usecases.InventarioFilaView, len(ds))
	for i, d := range ds {
		v, err := d.toView()
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// modelo IPV
type modeloInputDTO struct {
	AreaID    string `json:"area_id" validate:"required"`
	Productos []struct {
		ID    string `json:"id"`
		Orden int    `json:"orden"`
	} `json:"productos"`
}

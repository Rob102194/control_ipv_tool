package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"strings"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// GenerarReporte construye los datos estructurados del reporte de IPV de una
// fecha. El formateo de textos ("ACEITE: 0.2 L", "SAL (merma): …") es tarea de
// la capa HTTP.
func (s *IPVService) GenerarReporte(ctx context.Context, fecha domain.Date) (ReporteIPV, error) {
	registros, err := s.Repos.InventarioDiario().ListarPorFecha(ctx, fecha)
	if err != nil {
		return ReporteIPV{}, err
	}
	if len(registros) == 0 {
		return ReporteIPV{}, domain.NotFoundf("No se encontraron registros para la fecha especificada.")
	}

	areas, err := s.Repos.Areas().Listar(ctx)
	if err != nil {
		return ReporteIPV{}, err
	}
	productos, err := s.Repos.Productos().Listar(ctx, 0)
	if err != nil {
		return ReporteIPV{}, err
	}
	areaByID := make(map[string]domain.Area, len(areas))
	for _, a := range areas {
		areaByID[a.ID] = a
	}
	prodByID := make(map[string]domain.Producto, len(productos))
	for _, p := range productos {
		prodByID[p.ID] = p
	}

	rep := ReporteIPV{
		Fecha:   fecha,
		Areas:   map[string][]ReporteFila{},
		Resumen: map[string]ReporteResumenArea{},
	}
	// El resumen se inicializa para TODAS las áreas (como en la versión Python).
	for _, a := range areas {
		rep.Resumen[a.Nombre] = ReporteResumenArea{}
	}

	for _, reg := range registros {
		area, okA := areaByID[reg.AreaID]
		producto, okP := prodByID[reg.ProductoID]
		if !okA || !okP {
			continue
		}
		nombreArea := area.Nombre

		if _, seen := rep.Areas[nombreArea]; !seen {
			rep.OrdenAreas = append(rep.OrdenAreas, nombreArea)
		}
		rep.Areas[nombreArea] = append(rep.Areas[nombreArea], ReporteFila{
			Producto: producto.Nombre, UM: producto.UnidadMedida,
			Inicio: reg.Inicio, Entradas: reg.Entradas, Consumo: reg.Consumo,
			Merma: reg.Merma, OtrasSalidas: reg.OtrasSalidas,
			FinalTeorico: reg.FinalTeorico, FinalFisico: reg.FinalFisico,
			Diferencia: reg.Diferencia,
		})

		res := rep.Resumen[nombreArea]
		switch {
		case reg.Diferencia < 0:
			res.Faltantes = append(res.Faltantes, ReporteDelta{producto.Nombre, math.Abs(reg.Diferencia), producto.UnidadMedida})
		case reg.Diferencia > 0:
			res.Sobrantes = append(res.Sobrantes, ReporteDelta{producto.Nombre, reg.Diferencia, producto.UnidadMedida})
		}
		if reg.Merma > 0 {
			res.Mermas = append(res.Mermas, ReporteDelta{producto.Nombre, reg.Merma, producto.UnidadMedida})
		}
		rep.Resumen[nombreArea] = res

		rep.Notas = append(rep.Notas, notasDeComentario(producto.Nombre, reg.Comentario)...)
	}

	return rep, nil
}

// notasDeComentario interpreta el campo comentario: si es un JSON {campo: texto}
// emite una nota por cada texto no vacío, EN EL ORDEN DEL JSON (como Python, que
// itera comentarios.items() en orden de inserción); si no es JSON válido, una
// nota con el texto plano.
func notasDeComentario(producto, comentario string) []ReporteNota {
	if strings.TrimSpace(comentario) == "" {
		return nil
	}
	pares, err := jsonObjetoOrdenado(comentario)
	if err != nil {
		return []ReporteNota{{Producto: producto, Texto: comentario}}
	}
	var out []ReporteNota
	for _, kv := range pares {
		if strings.TrimSpace(kv.valor) != "" {
			out = append(out, ReporteNota{Producto: producto, Campo: kv.clave, Texto: kv.valor})
		}
	}
	return out
}

type kvString struct{ clave, valor string }

// jsonObjetoOrdenado decodifica un objeto JSON plano {string: string}
// conservando el orden de las claves.
func jsonObjetoOrdenado(s string) ([]kvString, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(s)))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, io.ErrUnexpectedEOF
	}
	var out []kvString
	for dec.More() {
		claveT, err := dec.Token()
		if err != nil {
			return nil, err
		}
		clave, ok := claveT.(string)
		if !ok {
			return nil, io.ErrUnexpectedEOF
		}
		var valor string
		if err := dec.Decode(&valor); err != nil {
			return nil, err
		}
		out = append(out, kvString{clave, valor})
	}
	return out, nil
}

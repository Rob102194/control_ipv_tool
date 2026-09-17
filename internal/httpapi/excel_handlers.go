package httpapi

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/Rob102194/control_ipv_tool/internal/adapters/excel"
	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

const xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// decodeArchivo decodifica el campo archivo_base64 a los bytes crudos del
// Excel.
//
// El archivo viaja en Base64 dentro del cuerpo JSON, no como
// multipart/form-data: el webview de escritorio (Wails, WKWebView en macOS)
// pierde el body de las peticiones POST con archivos binarios al pasar por el
// esquema wails:// (bug conocido de WKWebView con WKURLSchemeHandler, ver
// https://github.com/wailsapp/wails/issues/3037). Un JSON con el archivo en
// Base64 sí llega bien, igual que el resto de peticiones de la app.
func (a *api) decodeArchivo(w http.ResponseWriter, r *http.Request, archivoBase64 string) ([]byte, bool) {
	data, err := base64.StdEncoding.DecodeString(archivoBase64)
	if err != nil {
		a.fail(w, r, domain.Invalid("archivo_base64", "el archivo no está codificado en base64 válido"))
		return nil, false
	}
	return data, true
}

// --- productos --------------------------------------------------------

func (a *api) productosExport(w http.ResponseWriter, r *http.Request) {
	ps, err := a.svc.Productos.Listar(r.Context(), ports.OrdenProductoNombre)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	rows := make([]excel.ProductoRow, len(ps))
	for i, p := range ps {
		rows[i] = excel.ProductoRow{Nombre: p.Nombre, UnidadMedida: p.UnidadMedida}
	}
	var buf bytes.Buffer
	if err := excel.WriteProductos(&buf, rows); err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", xlsxMIME)
	w.Header().Set("Content-Disposition", `attachment; filename="productos.xlsx"`)
	w.Write(buf.Bytes())
}

func (a *api) productosImport(w http.ResponseWriter, r *http.Request) {
	var in importArchivoDTO
	if !a.decode(w, r, &in) {
		return
	}
	data, ok := a.decodeArchivo(w, r, in.ArchivoBase64)
	if !ok {
		return
	}
	rows, err := excel.ParseProductos(bytes.NewReader(data))
	if err != nil {
		a.fail(w, r, domain.Invalid("archivo_base64", err.Error()))
		return
	}
	items := make([]usecases.ImportProductoItem, len(rows))
	for i, x := range rows {
		items[i] = usecases.ImportProductoItem{Nombre: x.Nombre, UnidadMedida: x.UnidadMedida}
	}
	if _, err := a.svc.Productos.Importar(r.Context(), items); err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "Productos importados correctamente"})
}

// --- recetas ---------------------------------------------------------

func (a *api) recetasExport(w http.ResponseWriter, r *http.Request) {
	recs, err := a.svc.Recetas.Listar(r.Context(), ports.ListarRecetasOpts{})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	prods, err := a.svc.Productos.Listar(r.Context(), ports.OrdenProductoNombre)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	areas, err := a.svc.Areas.Listar(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	prodByID := map[string]domain.Producto{}
	for _, p := range prods {
		prodByID[p.ID] = p
	}
	areaByID := map[string]domain.Area{}
	for _, x := range areas {
		areaByID[x.ID] = x
	}

	var rows []excel.RecetaRow
	for _, rec := range recs {
		if len(rec.Ingredientes) == 0 {
			rows = append(rows, excel.RecetaRow{RecetaNombre: rec.Nombre})
			continue
		}
		for _, ing := range rec.Ingredientes {
			pn, um := "Producto no encontrado", ""
			if p, ok := prodByID[ing.ProductoID]; ok {
				pn, um = p.Nombre, p.UnidadMedida
			}
			an := "Área no encontrada"
			if x, ok := areaByID[ing.AreaID]; ok {
				an = x.Nombre
			}
			rows = append(rows, excel.RecetaRow{
				RecetaNombre: rec.Nombre, ProductoNombre: pn, UnidadMedida: um,
				AreaNombre: an, Cantidad: ing.Cantidad,
			})
		}
	}
	var buf bytes.Buffer
	if err := excel.WriteRecetas(&buf, rows); err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", xlsxMIME)
	w.Header().Set("Content-Disposition", `attachment; filename="recetas.xlsx"`)
	w.Write(buf.Bytes())
}

func (a *api) recetasImport(w http.ResponseWriter, r *http.Request) {
	var in importArchivoDTO
	if !a.decode(w, r, &in) {
		return
	}
	data, ok := a.decodeArchivo(w, r, in.ArchivoBase64)
	if !ok {
		return
	}
	rows, err := excel.ParseRecetas(bytes.NewReader(data))
	if err != nil {
		a.fail(w, r, domain.Invalid("archivo_base64", err.Error()))
		return
	}
	items := make([]usecases.ImportRecetaItem, len(rows))
	for i, x := range rows {
		items[i] = usecases.ImportRecetaItem{
			RecetaNombre: x.RecetaNombre, ProductoNombre: x.ProductoNombre,
			UnidadMedida: x.UnidadMedida, AreaNombre: x.AreaNombre, Cantidad: x.Cantidad,
		}
	}
	if _, err := a.svc.Recetas.Importar(r.Context(), items); err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "Recetas importadas correctamente"})
}

// --- ventas --------------------------------------------------------

func (a *api) ventasImportar(w http.ResponseWriter, r *http.Request) {
	var in importVentasDTO
	if !a.decode(w, r, &in) {
		return
	}
	data, ok := a.decodeArchivo(w, r, in.ArchivoBase64)
	if !ok {
		return
	}
	rows, err := excel.ParseVentas(bytes.NewReader(data))
	if err != nil {
		a.fail(w, r, domain.Invalid("archivo_base64", err.Error()))
		return
	}
	items := make([]usecases.ImportVentaItem, len(rows))
	for i, x := range rows {
		items[i] = usecases.ImportVentaItem{
			Fila: x.Fila, Nombre: x.Nombre, Cantidad: x.Cantidad,
			CantidadValida: x.CantidadValida, RawCantidad: x.RawCantidad,
		}
	}

	var fecha *domain.Date
	if in.Fecha != "" {
		d, err := domain.ParseDate(in.Fecha)
		if err != nil {
			a.fail(w, r, domain.Invalid("fecha", err.Error()))
			return
		}
		fecha = &d
	}

	res, err := a.svc.Ventas.Importar(r.Context(), items, fecha)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	ventas := make([]ventaDTO, len(res.Ventas))
	for i, v := range res.Ventas {
		ventas[i] = toVentaDTO(v)
	}
	nuevas := make([]recetaDTO, len(res.NuevasRecetas))
	for i, rc := range res.NuevasRecetas {
		nuevas[i] = toRecetaDTO(rc)
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"message":        pluralVentas(len(res.Ventas)),
		"ventas":         ventas,
		"nuevas_recetas": nuevas,
	})
}

func pluralVentas(n int) string {
	return "Se importaron " + strconv.Itoa(n) + " ventas correctamente"
}

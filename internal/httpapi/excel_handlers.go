package httpapi

import (
	"net/http"
	"strconv"

	"github.com/Rob102194/control_ipv_tool/internal/adapters/excel"
	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

const xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// openUpload devuelve el fichero subido en el campo "file".
func (a *api) openUpload(w http.ResponseWriter, r *http.Request) (multipartFile, bool) {
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		a.fail(w, r, domain.Invalid("file", "no se pudo leer el formulario"))
		return nil, false
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		a.fail(w, r, domain.Invalid("file", "No se encontró el archivo"))
		return nil, false
	}
	if hdr.Filename == "" {
		f.Close()
		a.fail(w, r, domain.Invalid("file", "No se seleccionó ningún archivo"))
		return nil, false
	}
	return f, true
}

type multipartFile interface {
	Read([]byte) (int, error)
	Close() error
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
	w.Header().Set("Content-Type", xlsxMIME)
	w.Header().Set("Content-Disposition", `attachment; filename="productos.xlsx"`)
	if err := excel.WriteProductos(w, rows); err != nil {
		a.logger.Error("exportando productos", "err", err)
	}
}

func (a *api) productosImport(w http.ResponseWriter, r *http.Request) {
	f, ok := a.openUpload(w, r)
	if !ok {
		return
	}
	defer f.Close()
	rows, err := excel.ParseProductos(f)
	if err != nil {
		a.fail(w, r, domain.Invalid("file", err.Error()))
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
	w.Header().Set("Content-Type", xlsxMIME)
	w.Header().Set("Content-Disposition", `attachment; filename="recetas.xlsx"`)
	if err := excel.WriteRecetas(w, rows); err != nil {
		a.logger.Error("exportando recetas", "err", err)
	}
}

func (a *api) recetasImport(w http.ResponseWriter, r *http.Request) {
	f, ok := a.openUpload(w, r)
	if !ok {
		return
	}
	defer f.Close()
	rows, err := excel.ParseRecetas(f)
	if err != nil {
		a.fail(w, r, domain.Invalid("file", err.Error()))
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
	f, ok := a.openUpload(w, r)
	if !ok {
		return
	}
	defer f.Close()

	rows, err := excel.ParseVentas(f)
	if err != nil {
		a.fail(w, r, domain.Invalid("file", err.Error()))
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
	if s := r.FormValue("fecha"); s != "" {
		d, err := domain.ParseDate(s)
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

package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// api agrupa los handlers HTTP sobre los servicios de casos de uso.
type api struct {
	svc    *usecases.Services
	logger *slog.Logger
}

func (a *api) fail(w http.ResponseWriter, r *http.Request, err error) {
	writeError(w, r, a.logger, err)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, v)
}

// decode lee el cuerpo JSON en dst y valida sus etiquetas `validate`.
func (a *api) decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		a.fail(w, r, domain.Invalid("body", "no se pudo leer el cuerpo de la solicitud"))
		return false
	}
	if len(body) == 0 {
		a.fail(w, r, domain.Invalid("body", "el cuerpo de la solicitud está vacío"))
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		a.fail(w, r, domain.Invalid("body", "JSON inválido: "+err.Error()))
		return false
	}
	// validator.Struct solo aplica a structs; los cuerpos que son arrays
	// (p. ej. /ipv/guardar) se validan por campo en el mapeo a dominio.
	if v := reflect.Indirect(reflect.ValueOf(dst)); v.Kind() == reflect.Struct {
		if err := validate.Struct(dst); err != nil {
			a.fail(w, r, domain.Invalid("body", err.Error()))
			return false
		}
	}
	return true
}

// ============================ Productos ============================

func (a *api) productosList(w http.ResponseWriter, r *http.Request) {
	orden := ports.OrdenProductoNombre
	switch r.URL.Query().Get("sort_by") {
	case "modificado":
		orden = ports.OrdenProductoModificado
	case "", "nombre":
		orden = ports.OrdenProductoNombre
	default:
		orden = ports.OrdenProductoReciente
	}
	ps, err := a.svc.Productos.Listar(r.Context(), orden)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]productoDTO, len(ps))
	for i, p := range ps {
		out[i] = toProductoDTO(p)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) productoGet(w http.ResponseWriter, r *http.Request) {
	p, err := a.svc.Productos.Obtener(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toProductoDTO(p))
}

func (a *api) productoCreate(w http.ResponseWriter, r *http.Request) {
	var in productoInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	p, err := a.svc.Productos.Crear(r.Context(), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toProductoDTO(p))
}

func (a *api) productoUpdate(w http.ResponseWriter, r *http.Request) {
	var in productoInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	p, err := a.svc.Productos.Actualizar(r.Context(), chi.URLParam(r, "id"), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toProductoDTO(p))
}

func (a *api) productoDelete(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Productos.Eliminar(r.Context(), chi.URLParam(r, "id")); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================ Áreas ===============================

func (a *api) areasList(w http.ResponseWriter, r *http.Request) {
	as, err := a.svc.Areas.Listar(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]areaDTO, len(as))
	for i, x := range as {
		out[i] = toAreaDTO(x)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) areaGet(w http.ResponseWriter, r *http.Request) {
	x, err := a.svc.Areas.Obtener(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toAreaDTO(x))
}

func (a *api) areaCreate(w http.ResponseWriter, r *http.Request) {
	var in areaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	x, err := a.svc.Areas.Crear(r.Context(), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toAreaDTO(x))
}

func (a *api) areaUpdate(w http.ResponseWriter, r *http.Request) {
	var in areaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	x, err := a.svc.Areas.Actualizar(r.Context(), chi.URLParam(r, "id"), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toAreaDTO(x))
}

func (a *api) areaDelete(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Areas.Eliminar(r.Context(), chi.URLParam(r, "id")); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================ Recetas =============================

func (a *api) recetasList(w http.ResponseWriter, r *http.Request) {
	opts := ports.ListarRecetasOpts{}
	switch r.URL.Query().Get("sort_by") {
	case "modificado":
		opts.Orden = ports.OrdenRecetaModificado
	case "", "nombre":
		opts.Orden = ports.OrdenRecetaNombre
	default:
		opts.Orden = ports.OrdenRecetaReciente
	}
	if r.URL.Query().Get("filter_by") == "sin_ingredientes" {
		opts.SoloSinIngredientes = true
	}
	rs, err := a.svc.Recetas.Listar(r.Context(), opts)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]recetaDTO, len(rs))
	for i, x := range rs {
		out[i] = toRecetaDTO(x)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) recetaGet(w http.ResponseWriter, r *http.Request) {
	x, err := a.svc.Recetas.Obtener(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toRecetaDTO(x))
}

func (a *api) recetaCreate(w http.ResponseWriter, r *http.Request) {
	var in recetaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	x, err := a.svc.Recetas.Crear(r.Context(), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toRecetaDTO(x))
}

func (a *api) recetaUpdate(w http.ResponseWriter, r *http.Request) {
	var in recetaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	x, err := a.svc.Recetas.Actualizar(r.Context(), chi.URLParam(r, "id"), in.toUC())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toRecetaDTO(x))
}

func (a *api) recetaDelete(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Recetas.Eliminar(r.Context(), chi.URLParam(r, "id")); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================ Ventas ==============================

func (a *api) ventasList(w http.ResponseWriter, r *http.Request) {
	vs, err := a.svc.Ventas.Listar(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]ventaDTO, len(vs))
	for i, v := range vs {
		out[i] = toVentaDTO(v)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) ventaGet(w http.ResponseWriter, r *http.Request) {
	v, err := a.svc.Ventas.Obtener(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toVentaDTO(v))
}

func (a *api) ventaCreate(w http.ResponseWriter, r *http.Request) {
	var in ventaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	uc, err := in.toUC()
	if err != nil {
		a.fail(w, r, err)
		return
	}
	v, err := a.svc.Ventas.Crear(r.Context(), uc)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toVentaDTO(v))
}

func (a *api) ventaUpdate(w http.ResponseWriter, r *http.Request) {
	var in ventaInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	uc, err := in.toUC()
	if err != nil {
		a.fail(w, r, err)
		return
	}
	v, err := a.svc.Ventas.Actualizar(r.Context(), chi.URLParam(r, "id"), uc)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toVentaDTO(v))
}

func (a *api) ventaDelete(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Ventas.Eliminar(r.Context(), chi.URLParam(r, "id")); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) ventasDeleteMultiple(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids" validate:"required,min=1"`
	}
	if !a.decode(w, r, &in) {
		return
	}
	if err := a.svc.Ventas.EliminarMultiples(r.Context(), in.IDs); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================ IPV ================================

func (a *api) ipvEstado(w http.ResponseWriter, r *http.Request) {
	fecha, ok := a.fechaQuery(w, r)
	if !ok {
		return
	}
	estado, err := a.svc.IPV.ObtenerEstado(r.Context(), fecha)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make(map[string][]inventarioDTO, len(estado))
	for area, filas := range estado {
		out[area] = toInventarioDTOs(filas)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) ipvCalcularConsumo(w http.ResponseWriter, r *http.Request) {
	fecha, ok := a.fechaQuery(w, r)
	if !ok {
		return
	}
	consumo, err := a.svc.IPV.CalcularConsumo(r.Context(), fecha)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, consumo)
}

func (a *api) ipvCalcular(w http.ResponseWriter, r *http.Request) {
	var in []inventarioInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	views, err := viewsFromInput(in)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toInventarioDTOs(a.svc.IPV.Recalcular(views)))
}

func (a *api) ipvGuardar(w http.ResponseWriter, r *http.Request) {
	var in []inventarioInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	views, err := viewsFromInput(in)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	guardadas, err := a.svc.IPV.Guardar(r.Context(), views)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toInventarioDTOs(guardadas))
}

func (a *api) ipvModelosGet(w http.ResponseWriter, r *http.Request) {
	modelos, err := a.svc.IPV.ObtenerModelos(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	type item struct {
		ProductoID string `json:"producto_id"`
		Orden      int    `json:"orden"`
	}
	out := make(map[string][]item, len(modelos))
	for areaID, items := range modelos {
		for _, it := range items {
			out[areaID] = append(out[areaID], item{it.ProductoID, it.Orden})
		}
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) ipvModelosPost(w http.ResponseWriter, r *http.Request) {
	var in modeloInputDTO
	if !a.decode(w, r, &in) {
		return
	}
	items := make([]usecases.ModeloItem, len(in.Productos))
	for i, p := range in.Productos {
		items[i] = usecases.ModeloItem{ProductoID: p.ID, Orden: p.Orden}
	}
	if err := a.svc.IPV.GuardarModelo(r.Context(), in.AreaID, items); err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"area_id": in.AreaID, "productos": in.Productos})
}

func (a *api) ipvRegistros(w http.ResponseWriter, r *http.Request) {
	fechas, err := a.svc.IPV.FechasConRegistros(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]map[string]string, len(fechas))
	for i, f := range fechas {
		out[i] = map[string]string{"fecha": f.String()}
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) ipvReporte(w http.ResponseWriter, r *http.Request) {
	fecha, ok := a.fechaQuery(w, r)
	if !ok {
		return
	}
	rep, err := a.svc.IPV.GenerarReporte(r.Context(), fecha)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, toReporteDTO(rep))
}

// ============================ Historial ===========================

func (a *api) historialGet(w http.ResponseWriter, r *http.Request) {
	hs, err := a.svc.Historial.Listar(r.Context(), domain.TipoEntidad(chi.URLParam(r, "entidad_tipo")))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]historialDTO, len(hs))
	for i, h := range hs {
		out[i] = toHistorialDTO(h)
	}
	respondJSON(w, http.StatusOK, out)
}

// ============================ helpers =============================

func (a *api) fechaQuery(w http.ResponseWriter, r *http.Request) (domain.Date, bool) {
	s := r.URL.Query().Get("fecha")
	if s == "" {
		a.fail(w, r, domain.Invalid("fecha", "El parámetro 'fecha' es requerido"))
		return domain.Date{}, false
	}
	f, err := domain.ParseDate(s)
	if err != nil {
		a.fail(w, r, domain.Invalid("fecha", "Formato de fecha inválido. Use YYYY-MM-DD"))
		return domain.Date{}, false
	}
	return f, true
}

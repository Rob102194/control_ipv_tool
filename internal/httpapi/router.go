package httpapi

import (
	"database/sql"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
)

// Deps agrupa las dependencias que necesita el router.
type Deps struct {
	Logger      *slog.Logger
	DB          *sql.DB
	Services    *usecases.Services
	CORSOrigins []string // vacío en escritorio (mismo origen)
	// SchemaVersion, si se indica, expone la versión de migración en /healthz.
	SchemaVersion func() (int64, error)
	// SPA, si se indica, sirve el frontend embebido para toda ruta que no sea
	// /api ni /healthz (con fallback a index.html).
	SPA http.Handler
}

// NewRouter arma el http.Handler de la aplicación: middleware transversal, el
// endpoint de salud y el grupo /api.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(recoverer(d.Logger))
	r.Use(requestLogger(d.Logger))
	if len(d.CORSOrigins) > 0 {
		r.Use(cors(d.CORSOrigins))
	}

	r.Get("/healthz", healthHandler(d))

	if d.Services != nil {
		a := &api{svc: d.Services, logger: d.Logger}
		r.Route("/api", func(r chi.Router) { mountAPI(r, a) })
	}

	if d.SPA != nil {
		r.Handle("/*", d.SPA)
	}

	return r
}

// mountAPI registra todas las rutas de negocio. Las rutas llevan barra final
// porque así las montaban los blueprints de Flask; se añaden también sin barra
// para tolerancia.
func mountAPI(r chi.Router, a *api) {
	// reg registra la ruta con y sin barra final (los blueprints de Flask
	// montaban con barra final; el frontend a veces la omite).
	reg := func(method, pat string, h http.HandlerFunc) {
		r.Method(method, pat, h)
		r.Method(method, pat+"/", h)
	}

	// Productos
	reg("GET", "/productos", a.productosList)
	reg("POST", "/productos", a.productoCreate)
	reg("GET", "/productos/export", a.productosExport)
	reg("POST", "/productos/import", a.productosImport)
	reg("GET", "/productos/{id}", a.productoGet)
	reg("PUT", "/productos/{id}", a.productoUpdate)
	reg("DELETE", "/productos/{id}", a.productoDelete)

	// Áreas
	reg("GET", "/areas", a.areasList)
	reg("POST", "/areas", a.areaCreate)
	reg("GET", "/areas/{id}", a.areaGet)
	reg("PUT", "/areas/{id}", a.areaUpdate)
	reg("DELETE", "/areas/{id}", a.areaDelete)

	// Recetas
	reg("GET", "/recetas", a.recetasList)
	reg("POST", "/recetas", a.recetaCreate)
	reg("GET", "/recetas/export", a.recetasExport)
	reg("POST", "/recetas/import", a.recetasImport)
	reg("GET", "/recetas/{id}", a.recetaGet)
	reg("PUT", "/recetas/{id}", a.recetaUpdate)
	reg("DELETE", "/recetas/{id}", a.recetaDelete)

	// Ventas
	reg("GET", "/ventas", a.ventasList)
	reg("POST", "/ventas", a.ventaCreate)
	reg("POST", "/ventas/delete-multiple", a.ventasDeleteMultiple)
	reg("POST", "/ventas/importar", a.ventasImportar)
	reg("GET", "/ventas/{id}", a.ventaGet)
	reg("PUT", "/ventas/{id}", a.ventaUpdate)
	reg("DELETE", "/ventas/{id}", a.ventaDelete)

	// IPV
	r.Get("/ipv/estado", a.ipvEstado)
	r.Get("/ipv/calcular-consumo", a.ipvCalcularConsumo)
	r.Post("/ipv/calcular", a.ipvCalcular)
	r.Post("/ipv/guardar", a.ipvGuardar)
	r.Get("/ipv/modelos", a.ipvModelosGet)
	r.Post("/ipv/modelos", a.ipvModelosPost)
	r.Get("/ipv/registros", a.ipvRegistros)
	r.Get("/ipv/reporte", a.ipvReporte)

	// Historial
	reg("GET", "/historial/{entidad_tipo}", a.historialGet)
}

// healthHandler reporta el estado del proceso y verifica la BD.
func healthHandler(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		}
		if d.DB != nil {
			if err := d.DB.PingContext(r.Context()); err != nil {
				body["status"] = "degraded"
				body["db"] = "unreachable"
				writeJSON(w, http.StatusServiceUnavailable, body)
				return
			}
		}
		if d.SchemaVersion != nil {
			if v, err := d.SchemaVersion(); err == nil {
				body["schema_version"] = v
			}
		}
		writeJSON(w, http.StatusOK, body)
	}
}

// cors implementa un CORS mínimo con lista blanca de orígenes. Solo se activa en
// despliegue web; en escritorio el frontend y la API comparten origen.
func cors(allowed []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(allowed, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

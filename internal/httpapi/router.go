package httpapi

import (
	"database/sql"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
)

// Deps agrupa las dependencias que necesita el router. En fases posteriores se
// añadirán aquí los casos de uso.
type Deps struct {
	Logger      *slog.Logger
	DB          *sql.DB
	CORSOrigins []string // vacío en escritorio (mismo origen)
	// SchemaVersion, si se indica, expone la versión de migración en /healthz.
	SchemaVersion func() (int64, error)
}

// NewRouter arma el http.Handler de la aplicación: middleware transversal, el
// endpoint de salud y el grupo /api (que se irá llenando por fase).
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(recoverer(d.Logger))
	r.Use(requestLogger(d.Logger))
	if len(d.CORSOrigins) > 0 {
		r.Use(cors(d.CORSOrigins))
	}

	r.Get("/healthz", healthHandler(d))

	r.Route("/api", func(r chi.Router) {
		// Fase 5: productos, areas, recetas, ventas, ipv, historial.
	})

	return r
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

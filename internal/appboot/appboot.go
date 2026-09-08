// Package appboot arma la aplicación completa (config -> BD -> migraciones ->
// repositorios -> casos de uso -> router HTTP) a partir de la configuración.
//
// Lo usan tanto cmd/server (servidor HTTP suelto) como cmd/desktop (Wails):
// solo cambia el envoltorio, no el cableado.
package appboot

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/Rob102194/control_ipv_tool/internal/adapters/sqlite"
	"github.com/Rob102194/control_ipv_tool/internal/app/usecases"
	"github.com/Rob102194/control_ipv_tool/internal/httpapi"
	"github.com/Rob102194/control_ipv_tool/internal/platform"
)

// App reúne lo necesario para servir la aplicación.
type App struct {
	Handler http.Handler
	DB      *sql.DB
	DBPath  string
	Logger  *slog.Logger
}

// Close libera los recursos (la conexión a la BD).
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

// Options son ajustes opcionales para New (extensiones del despliegue web).
type Options struct {
	// SPA es el handler del frontend embebido (nil: rutas no-API dan 404).
	SPA http.Handler
	// AuthMiddleware envuelve /api (nil en escritorio). Ver docs/web-roadmap.md.
	AuthMiddleware func(http.Handler) http.Handler
}

// New construye la App: resuelve el directorio de datos, importa la BD legada si
// procede, abre y migra la BD, cablea repositorios y casos de uso, y devuelve el
// http.Handler listo.
func New(cfg platform.Config, logger *slog.Logger, spa http.Handler) (*App, error) {
	return NewWithOptions(cfg, logger, Options{SPA: spa})
}

// NewWithOptions es como New pero acepta Options.
func NewWithOptions(cfg platform.Config, logger *slog.Logger, opts Options) (*App, error) {
	dataDir, err := platform.EnsureDataDir(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	dbPath := platform.ResolveDBPath(dataDir, cfg.DBPath)

	// Primer arranque de la versión Go: importa la BD de la versión Python si
	// existe y aún no hay BD en el destino.
	if err := platform.ImportLegacyIfNeeded(dbPath, logger); err != nil {
		return nil, err
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := sqlite.Migrate(db, logger); err != nil {
		_ = db.Close()
		return nil, err
	}

	store := sqlite.NewStore(db)
	services := usecases.New(usecases.Deps{
		Repos: store,
		UOW:   store,
		Clock: platform.SystemClock{},
		IDs:   platform.UUIDGen{},
	})

	router := httpapi.NewRouter(httpapi.Deps{
		Logger:         logger,
		DB:             db,
		Services:       services,
		CORSOrigins:    cfg.CORSOrigins,
		SchemaVersion:  func() (int64, error) { return sqlite.SchemaVersion(db) },
		SPA:            opts.SPA,
		AuthMiddleware: opts.AuthMiddleware,
	})

	return &App{Handler: router, DB: db, DBPath: dbPath, Logger: logger}, nil
}

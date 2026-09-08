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

// New construye la App: resuelve el directorio de datos, abre y migra la BD,
// cablea repositorios y casos de uso, y devuelve el http.Handler listo.
//
// spa es el handler del frontend embebido (puede ser nil: entonces las rutas
// que no son /api ni /healthz dan 404).
func New(cfg platform.Config, logger *slog.Logger, spa http.Handler) (*App, error) {
	dataDir, err := platform.EnsureDataDir(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	dbPath := platform.ResolveDBPath(dataDir, cfg.DBPath)

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
		Logger:        logger,
		DB:            db,
		Services:      services,
		CORSOrigins:   cfg.CORSOrigins,
		SchemaVersion: func() (int64, error) { return sqlite.SchemaVersion(db) },
		SPA:           spa,
	})

	return &App{Handler: router, DB: db, DBPath: dbPath, Logger: logger}, nil
}

// Package platform contiene utilidades transversales del proceso: configuración,
// resolución del directorio de datos, logging y ejecución de migraciones.
// No depende de ninguna capa de negocio.
package platform

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config reúne toda la configuración del proceso. Se lee de variables de entorno
// con el prefijo CONTROL_IPV_. Todos los valores tienen un default razonable para
// desarrollo, de modo que `go run ./cmd/server` funcione sin configuración previa.
type Config struct {
	// Env: "dev" o "prod". Controla el formato de logs (texto vs JSON).
	Env string `env:"CONTROL_IPV_ENV" envDefault:"dev"`
	// Addr: dirección de escucha del servidor HTTP.
	Addr string `env:"CONTROL_IPV_ADDR" envDefault:"127.0.0.1:5175"`
	// LogLevel: "debug", "info", "warn" o "error".
	LogLevel string `env:"CONTROL_IPV_LOG_LEVEL" envDefault:"info"`
	// DataDir: carpeta base para datos del usuario. Si queda vacía se resuelve
	// con os.UserConfigDir()/ControlIPV (ver EnsureDataDir).
	DataDir string `env:"CONTROL_IPV_DATA_DIR"`
	// DBPath: ruta del fichero SQLite. Si queda vacía se usa <DataDir>/inventario.db.
	DBPath string `env:"CONTROL_IPV_DB_PATH"`
	// CORSOrigins: orígenes permitidos para /api (solo relevante en despliegue web).
	// Vacío => sin cabeceras CORS (caso escritorio, mismo origen).
	CORSOrigins []string `env:"CONTROL_IPV_CORS_ORIGINS" envSeparator:","`
}

// LoadConfig lee la configuración del entorno.
func LoadConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("leyendo configuración del entorno: %w", err)
	}
	return cfg, nil
}

// IsProd indica si el proceso corre en modo producción.
func (c Config) IsProd() bool { return c.Env == "prod" }

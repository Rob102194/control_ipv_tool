package usecases

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// ConfiguracionService reúne los casos de uso de configuración del negocio.
type ConfiguracionService struct{ Deps }

func (s *ConfiguracionService) Obtener(ctx context.Context) (domain.Configuracion, error) {
	return s.Repos.Configuracion().Obtener(ctx)
}

func (s *ConfiguracionService) Actualizar(ctx context.Context, nombreNegocio string) (domain.Configuracion, error) {
	c := domain.Configuracion{NombreNegocio: nombreNegocio}
	c.Normalizar()
	if err := c.Validar(); err != nil {
		return domain.Configuracion{}, err
	}
	if err := s.Repos.Configuracion().Guardar(ctx, c); err != nil {
		return domain.Configuracion{}, err
	}
	return c, nil
}

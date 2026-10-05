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

// Actualizar aplica una actualización PARCIAL: nombreNegocio/backupDir en nil
// significa "no tocar este campo". Lee el valor actual antes de escribir
// para que, p. ej., cambiar solo el nombre desde la barra de navegación no
// borre la carpeta de backups configurada en la página de ajustes (y
// viceversa) — son dos acciones de UI independientes sobre la misma fila.
func (s *ConfiguracionService) Actualizar(ctx context.Context, nombreNegocio, backupDir *string) (domain.Configuracion, error) {
	actual, err := s.Repos.Configuracion().Obtener(ctx)
	if err != nil {
		return domain.Configuracion{}, err
	}
	if nombreNegocio != nil {
		actual.NombreNegocio = *nombreNegocio
	}
	if backupDir != nil {
		actual.BackupDir = *backupDir
	}
	actual.Normalizar()
	if err := actual.Validar(); err != nil {
		return domain.Configuracion{}, err
	}
	if err := s.Repos.Configuracion().Guardar(ctx, actual); err != nil {
		return domain.Configuracion{}, err
	}
	return actual, nil
}

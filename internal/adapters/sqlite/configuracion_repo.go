package sqlite

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

type configuracionRepo struct{ q querier }

const (
	sqlConfiguracionGet    = `SELECT nombre_negocio FROM configuracion WHERE id = 1`
	sqlConfiguracionUpdate = `UPDATE configuracion SET nombre_negocio = ? WHERE id = 1`
)

// Obtener siempre encuentra una fila: la migración 00003 pre-inserta id=1.
func (r *configuracionRepo) Obtener(ctx context.Context) (domain.Configuracion, error) {
	var c domain.Configuracion
	if err := r.q.QueryRowContext(ctx, sqlConfiguracionGet).Scan(&c.NombreNegocio); err != nil {
		return domain.Configuracion{}, mapErr(err)
	}
	return c, nil
}

func (r *configuracionRepo) Guardar(ctx context.Context, c domain.Configuracion) error {
	_, err := r.q.ExecContext(ctx, sqlConfiguracionUpdate, c.NombreNegocio)
	return mapErr(err)
}

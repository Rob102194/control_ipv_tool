package sqlite

import (
	"context"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type modeloRepo struct{ q querier }

const (
	sqlModeloList    = `SELECT id, area_id, producto_id, orden FROM modelo_ipv ORDER BY orden`
	sqlModeloDelArea = `DELETE FROM modelo_ipv WHERE area_id = ?`
	sqlModeloInsert  = `INSERT INTO modelo_ipv (id, area_id, producto_id, orden) VALUES (?, ?, ?, ?)`
)

func (r *modeloRepo) Listar(ctx context.Context) ([]domain.ModeloIPV, error) {
	rows, err := r.q.QueryContext(ctx, sqlModeloList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ModeloIPV
	for rows.Next() {
		var m domain.ModeloIPV
		if err := rows.Scan(&m.ID, &m.AreaID, &m.ProductoID, &m.Orden); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GuardarModelo reemplaza por completo el modelo de un área: borra sus filas y
// vuelve a insertarlas (mismo enfoque que save_modelo en Python).
func (r *modeloRepo) GuardarModelo(ctx context.Context, areaID string, filas []domain.ModeloIPV) error {
	return withTx(ctx, r.q, func(q querier) error {
		if _, err := q.ExecContext(ctx, sqlModeloDelArea, areaID); err != nil {
			return mapErr(err)
		}
		for _, m := range filas {
			if m.ID == "" {
				return fmt.Errorf("sqlite: fila de modelo sin id (lo asigna el caso de uso)")
			}
			if _, err := q.ExecContext(ctx, sqlModeloInsert, m.ID, areaID, m.ProductoID, m.Orden); err != nil {
				return mapErr(err)
			}
		}
		return nil
	})
}

var _ ports.ModeloIPVRepository = (*modeloRepo)(nil)

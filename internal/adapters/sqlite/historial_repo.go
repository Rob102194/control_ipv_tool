package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type historialRepo struct{ q querier }

const (
	sqlHistorialInsert = `INSERT INTO historial_cambios
		(id, entidad_tipo, entidad_id, campo_modificado, valor_anterior, valor_nuevo, fecha_cambio)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	sqlHistorialByTipo = `SELECT id, entidad_tipo, entidad_id, campo_modificado, valor_anterior, valor_nuevo, fecha_cambio
		FROM historial_cambios WHERE entidad_tipo = ? ORDER BY fecha_cambio DESC`
)

func (r *historialRepo) Registrar(ctx context.Context, h domain.HistorialCambios) error {
	if h.ID == "" {
		return fmt.Errorf("sqlite: cambio de historial sin id (lo asigna el caso de uso)")
	}
	if h.FechaCambio.IsZero() {
		return fmt.Errorf("sqlite: cambio de historial sin fecha (la asigna el caso de uso)")
	}
	_, err := r.q.ExecContext(ctx, sqlHistorialInsert,
		h.ID, string(h.EntidadTipo), h.EntidadID, h.CampoModificado,
		nullString(h.ValorAnterior), nullString(h.ValorNuevo), dbDateTime(h.FechaCambio),
	)
	return mapErr(err)
}

func (r *historialRepo) PorEntidad(ctx context.Context, tipo domain.TipoEntidad) ([]domain.HistorialCambios, error) {
	rows, err := r.q.QueryContext(ctx, sqlHistorialByTipo, string(tipo))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.HistorialCambios
	for rows.Next() {
		var (
			h           domain.HistorialCambios
			tipoStr     string
			anterior    sql.NullString
			nuevo       sql.NullString
			fechaCambio sql.NullString
		)
		if err := rows.Scan(&h.ID, &tipoStr, &h.EntidadID, &h.CampoModificado, &anterior, &nuevo, &fechaCambio); err != nil {
			return nil, err
		}
		h.EntidadTipo = domain.TipoEntidad(tipoStr)
		h.ValorAnterior = strFromNull(anterior)
		h.ValorNuevo = strFromNull(nuevo)
		if fechaCambio.Valid {
			t, err := parseDateTime(fechaCambio.String)
			if err != nil {
				return nil, fmt.Errorf("historial %s: %w", h.ID, err)
			}
			h.FechaCambio = t
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

var _ ports.HistorialRepository = (*historialRepo)(nil)

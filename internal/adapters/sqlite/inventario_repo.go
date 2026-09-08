package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type inventarioRepo struct{ q querier }

const sqlInvCols = `id, fecha, area_id, producto_id, inicio, entradas, consumo, merma,
	otras_salidas, final_fisico, final_teorico, diferencia, comentario`

const sqlInvUpsert = `
INSERT INTO inventario_diario
	(id, fecha, area_id, producto_id, inicio, entradas, consumo, merma,
	 otras_salidas, final_fisico, final_teorico, diferencia, comentario)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(fecha, area_id, producto_id) DO UPDATE SET
	inicio        = excluded.inicio,
	entradas      = excluded.entradas,
	consumo       = excluded.consumo,
	merma         = excluded.merma,
	otras_salidas = excluded.otras_salidas,
	final_fisico  = excluded.final_fisico,
	final_teorico = excluded.final_teorico,
	diferencia    = excluded.diferencia,
	comentario    = excluded.comentario`

func scanInventario(s interface{ Scan(...any) error }) (domain.InventarioDiario, error) {
	var (
		iv         domain.InventarioDiario
		fecha      string
		comentario sql.NullString
		f          [8]sql.NullFloat64 // inicio..diferencia
	)
	err := s.Scan(
		&iv.ID, &fecha, &iv.AreaID, &iv.ProductoID,
		&f[0], &f[1], &f[2], &f[3], &f[4], &f[5], &f[6], &f[7],
		&comentario,
	)
	if err != nil {
		return domain.InventarioDiario{}, err
	}
	d, err := parseDate(fecha)
	if err != nil {
		return domain.InventarioDiario{}, fmt.Errorf("inventario %s: %w", iv.ID, err)
	}
	iv.Fecha = d
	iv.Inicio = f[0].Float64
	iv.Entradas = f[1].Float64
	iv.Consumo = f[2].Float64
	iv.Merma = f[3].Float64
	iv.OtrasSalidas = f[4].Float64
	iv.FinalFisico = f[5].Float64
	iv.FinalTeorico = f[6].Float64
	iv.Diferencia = f[7].Float64
	iv.Comentario = strFromNull(comentario)
	return iv, nil
}

func (r *inventarioRepo) ListarPorFecha(ctx context.Context, fecha domain.Date) ([]domain.InventarioDiario, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT `+sqlInvCols+` FROM inventario_diario WHERE fecha = ? ORDER BY rowid`, dbDate(fecha))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.InventarioDiario
	for rows.Next() {
		iv, err := scanInventario(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, iv)
	}
	return out, rows.Err()
}

func (r *inventarioRepo) BuscarPorFechaAreaProducto(ctx context.Context, fecha domain.Date, areaID, productoID string) (domain.InventarioDiario, error) {
	iv, err := scanInventario(r.q.QueryRowContext(ctx,
		`SELECT `+sqlInvCols+` FROM inventario_diario WHERE fecha = ? AND area_id = ? AND producto_id = ?`,
		dbDate(fecha), areaID, productoID))
	return iv, mapErr(err)
}

func (r *inventarioRepo) FechasConRegistros(ctx context.Context) ([]domain.Date, error) {
	rows, err := r.q.QueryContext(ctx,
		`SELECT DISTINCT fecha FROM inventario_diario ORDER BY fecha DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Date
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		d, err := parseDate(s)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// FinalFisicoDiaAnterior devuelve el final_fisico del día previo, o 0 si no hay
// registro (mismo comportamiento que get_inicio_from_previous_day en Python).
func (r *inventarioRepo) FinalFisicoDiaAnterior(ctx context.Context, fecha domain.Date, areaID, productoID string) (float64, error) {
	var v sql.NullFloat64
	err := r.q.QueryRowContext(ctx,
		`SELECT final_fisico FROM inventario_diario WHERE fecha = ? AND area_id = ? AND producto_id = ?`,
		dbDate(fecha.AddDays(-1)), areaID, productoID).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return v.Float64, nil
}

func (r *inventarioRepo) GuardarTodos(ctx context.Context, filas []domain.InventarioDiario) error {
	return withTx(ctx, r.q, func(q querier) error {
		for _, iv := range filas {
			if iv.ID == "" {
				return fmt.Errorf("sqlite: fila de inventario sin id (lo asigna el caso de uso)")
			}
			_, err := q.ExecContext(ctx, sqlInvUpsert,
				iv.ID, dbDate(iv.Fecha), iv.AreaID, iv.ProductoID,
				iv.Inicio, iv.Entradas, iv.Consumo, iv.Merma, iv.OtrasSalidas,
				iv.FinalFisico, iv.FinalTeorico, iv.Diferencia, nullString(iv.Comentario),
			)
			if err != nil {
				return mapErr(err)
			}
		}
		return nil
	})
}

var _ ports.InventarioDiarioRepository = (*inventarioRepo)(nil)

package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type ventaRepo struct{ q querier }

const (
	sqlVentaCols    = `id, receta_nombre, cantidad, fecha`
	sqlVentaInsert  = `INSERT INTO ventas (id, receta_nombre, cantidad, fecha) VALUES (?, ?, ?, ?)`
	sqlVentaByID    = `SELECT ` + sqlVentaCols + ` FROM ventas WHERE id = ?`
	sqlVentaList    = `SELECT ` + sqlVentaCols + ` FROM ventas ORDER BY rowid`
	sqlVentaByFecha = `SELECT ` + sqlVentaCols + ` FROM ventas WHERE fecha = ? ORDER BY rowid`
	sqlVentaUpdate  = `UPDATE ventas SET receta_nombre = ?, cantidad = ?, fecha = ? WHERE id = ?`
	sqlVentaDelete  = `DELETE FROM ventas WHERE id = ?`
)

func scanVenta(s interface{ Scan(...any) error }) (domain.Venta, error) {
	var (
		v     domain.Venta
		fecha string
	)
	if err := s.Scan(&v.ID, &v.RecetaNombre, &v.Cantidad, &fecha); err != nil {
		return domain.Venta{}, err
	}
	d, err := parseDate(fecha)
	if err != nil {
		return domain.Venta{}, fmt.Errorf("venta %s: %w", v.ID, err)
	}
	v.Fecha = d
	return v, nil
}

func (r *ventaRepo) Crear(ctx context.Context, v domain.Venta) (domain.Venta, error) {
	if v.ID == "" {
		return domain.Venta{}, fmt.Errorf("sqlite: venta sin id (lo asigna el caso de uso)")
	}
	if _, err := r.q.ExecContext(ctx, sqlVentaInsert, v.ID, v.RecetaNombre, v.Cantidad, dbDate(v.Fecha)); err != nil {
		return domain.Venta{}, mapErr(err)
	}
	return v, nil
}

func (r *ventaRepo) CrearMultiples(ctx context.Context, vs []domain.Venta) ([]domain.Venta, error) {
	err := withTx(ctx, r.q, func(q querier) error {
		for _, v := range vs {
			if v.ID == "" {
				return fmt.Errorf("sqlite: venta sin id en lote")
			}
			if _, err := q.ExecContext(ctx, sqlVentaInsert, v.ID, v.RecetaNombre, v.Cantidad, dbDate(v.Fecha)); err != nil {
				return mapErr(err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vs, nil
}

func (r *ventaRepo) ObtenerPorID(ctx context.Context, id string) (domain.Venta, error) {
	v, err := scanVenta(r.q.QueryRowContext(ctx, sqlVentaByID, id))
	return v, mapErr(err)
}

func (r *ventaRepo) Listar(ctx context.Context) ([]domain.Venta, error) {
	return r.query(ctx, sqlVentaList)
}

func (r *ventaRepo) ListarPorFecha(ctx context.Context, fecha domain.Date) ([]domain.Venta, error) {
	return r.query(ctx, sqlVentaByFecha, dbDate(fecha))
}

func (r *ventaRepo) query(ctx context.Context, q string, args ...any) ([]domain.Venta, error) {
	rows, err := r.q.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Venta
	for rows.Next() {
		v, err := scanVenta(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *ventaRepo) Actualizar(ctx context.Context, v domain.Venta) (domain.Venta, error) {
	res, err := r.q.ExecContext(ctx, sqlVentaUpdate, v.RecetaNombre, v.Cantidad, dbDate(v.Fecha), v.ID)
	if err != nil {
		return domain.Venta{}, mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Venta{}, ports.ErrNoEncontrado
	}
	return v, nil
}

func (r *ventaRepo) Eliminar(ctx context.Context, id string) error {
	res, err := r.q.ExecContext(ctx, sqlVentaDelete, id)
	if err != nil {
		return mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ports.ErrNoEncontrado
	}
	return nil
}

// EliminarMultiples borra en bloque. No falla si algún id no existe (igual que la
// versión Python, que hacía un DELETE ... WHERE id IN (...)).
func (r *ventaRepo) EliminarMultiples(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := r.q.ExecContext(ctx, `DELETE FROM ventas WHERE id IN (`+placeholders+`)`, args...)
	return mapErr(err)
}

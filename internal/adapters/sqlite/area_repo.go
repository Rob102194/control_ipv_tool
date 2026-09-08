package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type areaRepo struct{ q querier }

const (
	sqlAreaCols   = `id, nombre, codigo`
	sqlAreaInsert = `INSERT INTO areas (id, nombre, codigo) VALUES (?, ?, ?)`
	sqlAreaByID   = `SELECT ` + sqlAreaCols + ` FROM areas WHERE id = ?`
	sqlAreaByName = `SELECT ` + sqlAreaCols + ` FROM areas WHERE nombre = ?`
	sqlAreaList   = `SELECT ` + sqlAreaCols + ` FROM areas ORDER BY rowid`
	sqlAreaUpdate = `UPDATE areas SET nombre = ?, codigo = ? WHERE id = ?`
	sqlAreaDelete = `DELETE FROM areas WHERE id = ?`
)

func scanArea(s interface{ Scan(...any) error }) (domain.Area, error) {
	var a domain.Area
	var codigo sql.NullString
	if err := s.Scan(&a.ID, &a.Nombre, &codigo); err != nil {
		return domain.Area{}, err
	}
	a.Codigo = strFromNull(codigo)
	return a, nil
}

func (r *areaRepo) Crear(ctx context.Context, a domain.Area) (domain.Area, error) {
	if a.ID == "" {
		return domain.Area{}, fmt.Errorf("sqlite: área sin id (lo asigna el caso de uso)")
	}
	if _, err := r.q.ExecContext(ctx, sqlAreaInsert, a.ID, a.Nombre, nullString(a.Codigo)); err != nil {
		return domain.Area{}, mapErr(err)
	}
	return a, nil
}

func (r *areaRepo) ObtenerPorID(ctx context.Context, id string) (domain.Area, error) {
	a, err := scanArea(r.q.QueryRowContext(ctx, sqlAreaByID, id))
	return a, mapErr(err)
}

func (r *areaRepo) BuscarPorNombre(ctx context.Context, nombre string) (domain.Area, error) {
	a, err := scanArea(r.q.QueryRowContext(ctx, sqlAreaByName, nombre))
	return a, mapErr(err)
}

func (r *areaRepo) Listar(ctx context.Context) ([]domain.Area, error) {
	rows, err := r.q.QueryContext(ctx, sqlAreaList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Area
	for rows.Next() {
		a, err := scanArea(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *areaRepo) Actualizar(ctx context.Context, a domain.Area) (domain.Area, error) {
	res, err := r.q.ExecContext(ctx, sqlAreaUpdate, a.Nombre, nullString(a.Codigo), a.ID)
	if err != nil {
		return domain.Area{}, mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Area{}, ports.ErrNoEncontrado
	}
	return a, nil
}

func (r *areaRepo) Eliminar(ctx context.Context, id string) error {
	res, err := r.q.ExecContext(ctx, sqlAreaDelete, id)
	if err != nil {
		return mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ports.ErrNoEncontrado
	}
	return nil
}

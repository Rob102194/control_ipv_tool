package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type recetaRepo struct{ q querier }

const (
	sqlRecetaInsert   = `INSERT INTO recetas (id, nombre, activa) VALUES (?, ?, ?)`
	sqlRecetaByID     = `SELECT id, nombre, activa FROM recetas WHERE id = ?`
	sqlRecetaByName   = `SELECT id, nombre, activa FROM recetas WHERE nombre = ?`
	sqlRecetaUpdate   = `UPDATE recetas SET nombre = ?, activa = ? WHERE id = ?`
	sqlRecetaDelete   = `DELETE FROM recetas WHERE id = ?`
	sqlIngredienteIns = `INSERT INTO ingredientes (id, receta_id, producto_id, area_id, cantidad) VALUES (?, ?, ?, ?, ?)`
	sqlIngredienteDel = `DELETE FROM ingredientes WHERE receta_id = ?`
	sqlIngredienteByR = `SELECT id, receta_id, producto_id, area_id, cantidad FROM ingredientes WHERE receta_id = ? ORDER BY rowid`
)

func scanReceta(s interface{ Scan(...any) error }) (domain.Receta, error) {
	var r domain.Receta
	if err := s.Scan(&r.ID, &r.Nombre, &r.Activa); err != nil {
		return domain.Receta{}, err
	}
	return r, nil
}

func scanIngrediente(s interface{ Scan(...any) error }) (domain.Ingrediente, error) {
	var i domain.Ingrediente
	if err := s.Scan(&i.ID, &i.RecetaID, &i.ProductoID, &i.AreaID, &i.Cantidad); err != nil {
		return domain.Ingrediente{}, err
	}
	return i, nil
}

func (r *recetaRepo) Crear(ctx context.Context, rec domain.Receta) (domain.Receta, error) {
	if rec.ID == "" {
		return domain.Receta{}, fmt.Errorf("sqlite: receta sin id (lo asigna el caso de uso)")
	}
	err := withTx(ctx, r.q, func(q querier) error {
		if _, err := q.ExecContext(ctx, sqlRecetaInsert, rec.ID, rec.Nombre, rec.Activa); err != nil {
			return mapErr(err)
		}
		return insertarIngredientes(ctx, q, rec.ID, rec.Ingredientes)
	})
	if err != nil {
		return domain.Receta{}, err
	}
	return rec, nil
}

// CrearMultiples inserta solo las cabeceras de receta, SIN ingredientes, igual
// que crear_multiples de la versión Python (que se usa al importar ventas con
// recetas nuevas).
func (r *recetaRepo) CrearMultiples(ctx context.Context, recs []domain.Receta) ([]domain.Receta, error) {
	err := withTx(ctx, r.q, func(q querier) error {
		for _, rec := range recs {
			if rec.ID == "" {
				return fmt.Errorf("sqlite: receta sin id en lote")
			}
			if _, err := q.ExecContext(ctx, sqlRecetaInsert, rec.ID, rec.Nombre, rec.Activa); err != nil {
				return mapErr(err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func insertarIngredientes(ctx context.Context, q querier, recetaID string, ings []domain.Ingrediente) error {
	for _, ing := range ings {
		if ing.ID == "" {
			return fmt.Errorf("sqlite: ingrediente sin id (lo asigna el caso de uso)")
		}
		if _, err := q.ExecContext(ctx, sqlIngredienteIns, ing.ID, recetaID, ing.ProductoID, ing.AreaID, ing.Cantidad); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func (r *recetaRepo) ObtenerPorID(ctx context.Context, id string) (domain.Receta, error) {
	return r.cargarUna(ctx, sqlRecetaByID, id)
}

func (r *recetaRepo) BuscarPorNombre(ctx context.Context, nombre string) (domain.Receta, error) {
	return r.cargarUna(ctx, sqlRecetaByName, nombre)
}

func (r *recetaRepo) cargarUna(ctx context.Context, query, arg string) (domain.Receta, error) {
	rec, err := scanReceta(r.q.QueryRowContext(ctx, query, arg))
	if err != nil {
		return domain.Receta{}, mapErr(err)
	}
	ings, err := r.ingredientesDe(ctx, rec.ID)
	if err != nil {
		return domain.Receta{}, err
	}
	rec.Ingredientes = ings
	return rec, nil
}

func (r *recetaRepo) ingredientesDe(ctx context.Context, recetaID string) ([]domain.Ingrediente, error) {
	rows, err := r.q.QueryContext(ctx, sqlIngredienteByR, recetaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ingrediente
	for rows.Next() {
		ing, err := scanIngrediente(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ing)
	}
	return out, rows.Err()
}

func (r *recetaRepo) Listar(ctx context.Context, opts ports.ListarRecetasOpts) ([]domain.Receta, error) {
	var b strings.Builder
	b.WriteString(`SELECT r.id, r.nombre, r.activa FROM recetas r`)

	switch opts.Orden {
	case ports.OrdenRecetaModificado:
		b.WriteString(` LEFT JOIN (SELECT entidad_id, MAX(fecha_cambio) AS mx FROM historial_cambios WHERE entidad_tipo = 'Receta' GROUP BY entidad_id) h ON h.entidad_id = r.id`)
	}
	if opts.SoloSinIngredientes {
		b.WriteString(` WHERE NOT EXISTS (SELECT 1 FROM ingredientes i WHERE i.receta_id = r.id)`)
	}
	switch opts.Orden {
	case ports.OrdenRecetaModificado:
		b.WriteString(` ORDER BY h.mx IS NULL, h.mx DESC, r.nombre`)
	case ports.OrdenRecetaReciente:
		b.WriteString(` ORDER BY r.id DESC`)
	default:
		b.WriteString(` ORDER BY r.nombre`)
	}

	rows, err := r.q.QueryContext(ctx, b.String())
	if err != nil {
		return nil, err
	}
	recetas, err := collectRecetas(rows)
	if err != nil {
		return nil, err
	}
	if len(recetas) == 0 {
		return recetas, nil
	}

	// Carga de ingredientes en un solo golpe y agrupación por receta.
	ids := make([]any, len(recetas))
	idx := make(map[string]int, len(recetas))
	for i := range recetas {
		ids[i] = recetas[i].ID
		idx[recetas[i].ID] = i
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	irows, err := r.q.QueryContext(ctx,
		`SELECT id, receta_id, producto_id, area_id, cantidad FROM ingredientes WHERE receta_id IN (`+ph+`) ORDER BY rowid`, ids...)
	if err != nil {
		return nil, err
	}
	defer irows.Close()
	for irows.Next() {
		ing, err := scanIngrediente(irows)
		if err != nil {
			return nil, err
		}
		if i, ok := idx[ing.RecetaID]; ok {
			recetas[i].Ingredientes = append(recetas[i].Ingredientes, ing)
		}
	}
	return recetas, irows.Err()
}

func collectRecetas(rows *sql.Rows) ([]domain.Receta, error) {
	defer rows.Close()
	var out []domain.Receta
	for rows.Next() {
		rec, err := scanReceta(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *recetaRepo) Actualizar(ctx context.Context, rec domain.Receta) (domain.Receta, error) {
	err := withTx(ctx, r.q, func(q querier) error {
		res, err := q.ExecContext(ctx, sqlRecetaUpdate, rec.Nombre, rec.Activa, rec.ID)
		if err != nil {
			return mapErr(err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ports.ErrNoEncontrado
		}
		if _, err := q.ExecContext(ctx, sqlIngredienteDel, rec.ID); err != nil {
			return mapErr(err)
		}
		return insertarIngredientes(ctx, q, rec.ID, rec.Ingredientes)
	})
	if err != nil {
		return domain.Receta{}, err
	}
	return rec, nil
}

func (r *recetaRepo) Eliminar(ctx context.Context, id string) error {
	return withTx(ctx, r.q, func(q querier) error {
		if _, err := q.ExecContext(ctx, sqlIngredienteDel, id); err != nil {
			return mapErr(err)
		}
		res, err := q.ExecContext(ctx, sqlRecetaDelete, id)
		if err != nil {
			return mapErr(err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ports.ErrNoEncontrado
		}
		return nil
	})
}

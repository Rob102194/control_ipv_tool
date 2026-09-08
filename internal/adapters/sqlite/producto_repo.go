package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

type productoRepo struct{ q querier }

const (
	sqlProductoCols   = `id, nombre, unidad_medida`
	sqlProductoInsert = `INSERT INTO productos (id, nombre, unidad_medida) VALUES (?, ?, ?)`
	sqlProductoByID   = `SELECT ` + sqlProductoCols + ` FROM productos WHERE id = ?`
	sqlProductoByName = `SELECT ` + sqlProductoCols + ` FROM productos WHERE nombre = ?`
	sqlProductoUpdate = `UPDATE productos SET nombre = ?, unidad_medida = ? WHERE id = ?`
	sqlProductoDelete = `DELETE FROM productos WHERE id = ?`
	sqlProductoEnUso  = `SELECT
		EXISTS(SELECT 1 FROM ingredientes      WHERE producto_id = ?) OR
		EXISTS(SELECT 1 FROM movimientos       WHERE producto_id = ?) OR
		EXISTS(SELECT 1 FROM inventario_diario WHERE producto_id = ?) OR
		EXISTS(SELECT 1 FROM modelo_ipv        WHERE producto_id = ?)`
)

func scanProducto(s interface{ Scan(...any) error }) (domain.Producto, error) {
	var p domain.Producto
	if err := s.Scan(&p.ID, &p.Nombre, &p.UnidadMedida); err != nil {
		return domain.Producto{}, err
	}
	return p, nil
}

func (r *productoRepo) Crear(ctx context.Context, p domain.Producto) (domain.Producto, error) {
	if p.ID == "" {
		return domain.Producto{}, fmt.Errorf("sqlite: producto sin id (lo asigna el caso de uso)")
	}
	if _, err := r.q.ExecContext(ctx, sqlProductoInsert, p.ID, p.Nombre, p.UnidadMedida); err != nil {
		return domain.Producto{}, mapErr(err)
	}
	return p, nil
}

func (r *productoRepo) ObtenerPorID(ctx context.Context, id string) (domain.Producto, error) {
	p, err := scanProducto(r.q.QueryRowContext(ctx, sqlProductoByID, id))
	return p, mapErr(err)
}

func (r *productoRepo) BuscarPorNombre(ctx context.Context, nombre string) (domain.Producto, error) {
	p, err := scanProducto(r.q.QueryRowContext(ctx, sqlProductoByName, nombre))
	return p, mapErr(err)
}

func (r *productoRepo) Listar(ctx context.Context, orden ports.OrdenProducto) ([]domain.Producto, error) {
	var order string
	switch orden {
	case ports.OrdenProductoModificado:
		// LEFT JOIN con el último cambio registrado por producto.
		return r.listarPorModificado(ctx)
	case ports.OrdenProductoReciente:
		order = `ORDER BY id DESC`
	default:
		order = `ORDER BY nombre`
	}
	rows, err := r.q.QueryContext(ctx, `SELECT `+sqlProductoCols+` FROM productos `+order)
	if err != nil {
		return nil, err
	}
	return collectProductos(rows)
}

func (r *productoRepo) listarPorModificado(ctx context.Context) ([]domain.Producto, error) {
	const q = `
		SELECT p.id, p.nombre, p.unidad_medida
		FROM productos p
		LEFT JOIN (
			SELECT entidad_id, MAX(fecha_cambio) AS mx
			FROM historial_cambios
			WHERE entidad_tipo = 'Producto'
			GROUP BY entidad_id
		) h ON h.entidad_id = p.id
		ORDER BY h.mx IS NULL, h.mx DESC, p.nombre`
	rows, err := r.q.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	return collectProductos(rows)
}

func collectProductos(rows *sql.Rows) ([]domain.Producto, error) {
	defer rows.Close()
	var out []domain.Producto
	for rows.Next() {
		p, err := scanProducto(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *productoRepo) Actualizar(ctx context.Context, p domain.Producto) (domain.Producto, error) {
	res, err := r.q.ExecContext(ctx, sqlProductoUpdate, p.Nombre, p.UnidadMedida, p.ID)
	if err != nil {
		return domain.Producto{}, mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Producto{}, ports.ErrNoEncontrado
	}
	return p, nil
}

func (r *productoRepo) Eliminar(ctx context.Context, id string) error {
	res, err := r.q.ExecContext(ctx, sqlProductoDelete, id)
	if err != nil {
		return mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ports.ErrNoEncontrado
	}
	return nil
}

func (r *productoRepo) EnUso(ctx context.Context, id string) (bool, error) {
	var enUso bool
	if err := r.q.QueryRowContext(ctx, sqlProductoEnUso, id, id, id, id).Scan(&enUso); err != nil {
		return false, err
	}
	return enUso, nil
}

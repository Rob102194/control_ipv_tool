package sqlite

import (
	"context"
	"database/sql"

	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// querier es lo común entre *sql.DB y *sql.Tx: los repositorios trabajan contra
// esta interfaz, así el mismo código sirve dentro y fuera de una transacción.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// repoProvider construye los repositorios sobre un querier concreto.
// Implementa ports.Repos.
type repoProvider struct{ q querier }

func (r repoProvider) Productos() ports.ProductoRepository { return &productoRepo{q: r.q} }
func (r repoProvider) Areas() ports.AreaRepository         { return &areaRepo{q: r.q} }
func (r repoProvider) Recetas() ports.RecetaRepository     { return &recetaRepo{q: r.q} }
func (r repoProvider) Ventas() ports.VentaRepository       { return &ventaRepo{q: r.q} }
func (r repoProvider) InventarioDiario() ports.InventarioDiarioRepository {
	return &inventarioRepo{q: r.q}
}
func (r repoProvider) ModelosIPV() ports.ModeloIPVRepository { return &modeloRepo{q: r.q} }
func (r repoProvider) Historial() ports.HistorialRepository  { return &historialRepo{q: r.q} }

// Store es el adaptador de persistencia. Como ports.Repos opera sin transacción
// (lecturas y escrituras sueltas); como ports.UnitOfWork agrupa varias
// operaciones en una transacción con Do.
type Store struct {
	db *sql.DB
	repoProvider
}

// NewStore envuelve un *sql.DB ya abierto y migrado.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db, repoProvider: repoProvider{q: db}}
}

// Do ejecuta fn dentro de una transacción: commit si fn devuelve nil, rollback
// si devuelve error o entra en pánico.
func (s *Store) Do(ctx context.Context, fn func(ports.Repos) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(repoProvider{q: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// withTx garantiza que fn corra dentro de una transacción. Si q ya es una
// transacción (estamos dentro de UnitOfWork.Do) la reutiliza; si es *sql.DB
// abre una transacción propia. Así cada método de repositorio es atómico
// aunque se llame suelto.
func withTx(ctx context.Context, q querier, fn func(querier) error) error {
	if _, ok := q.(*sql.Tx); ok {
		return fn(q)
	}
	db := q.(*sql.DB)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

var (
	_ ports.Repos      = (*Store)(nil)
	_ ports.UnitOfWork = (*Store)(nil)
)

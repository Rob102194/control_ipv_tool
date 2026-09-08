package ports

import (
	"context"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
)

// --- Opciones de listado --------------------------------------------------

// OrdenProducto controla el orden de ProductoRepository.Listar.
type OrdenProducto int

const (
	OrdenProductoNombre     OrdenProducto = iota // alfabético (por defecto)
	OrdenProductoModificado                      // por fecha del último cambio (historial), desc
	OrdenProductoReciente                        // por id, desc (equivale al "else" de la versión Python)
)

// OrdenReceta controla el orden de RecetaRepository.Listar.
type OrdenReceta int

const (
	OrdenRecetaNombre     OrdenReceta = iota
	OrdenRecetaModificado             // por fecha del último cambio, desc
)

// ListarRecetasOpts agrupa los parámetros de RecetaRepository.Listar.
type ListarRecetasOpts struct {
	Orden  OrdenReceta
	Filtro string // subcadena por nombre; "" = sin filtro
}

// --- Repositorios -------------------------------------------------------

// ProductoRepository persiste productos.
type ProductoRepository interface {
	Crear(ctx context.Context, p domain.Producto) (domain.Producto, error)
	ObtenerPorID(ctx context.Context, id string) (domain.Producto, error)
	BuscarPorNombre(ctx context.Context, nombre string) (domain.Producto, error)
	Listar(ctx context.Context, orden OrdenProducto) ([]domain.Producto, error)
	Actualizar(ctx context.Context, p domain.Producto) (domain.Producto, error)
	Eliminar(ctx context.Context, id string) error
	// EnUso indica si el producto está referenciado por ingredientes,
	// movimientos, inventario_diario o modelo_ipv.
	EnUso(ctx context.Context, id string) (bool, error)
}

// AreaRepository persiste áreas.
type AreaRepository interface {
	Crear(ctx context.Context, a domain.Area) (domain.Area, error)
	ObtenerPorID(ctx context.Context, id string) (domain.Area, error)
	BuscarPorNombre(ctx context.Context, nombre string) (domain.Area, error)
	Listar(ctx context.Context) ([]domain.Area, error)
	Actualizar(ctx context.Context, a domain.Area) (domain.Area, error)
	Eliminar(ctx context.Context, id string) error
}

// RecetaRepository persiste recetas con sus ingredientes.
type RecetaRepository interface {
	Crear(ctx context.Context, r domain.Receta) (domain.Receta, error)
	CrearMultiples(ctx context.Context, rs []domain.Receta) ([]domain.Receta, error)
	ObtenerPorID(ctx context.Context, id string) (domain.Receta, error)
	BuscarPorNombre(ctx context.Context, nombre string) (domain.Receta, error)
	Listar(ctx context.Context, opts ListarRecetasOpts) ([]domain.Receta, error)
	Actualizar(ctx context.Context, r domain.Receta) (domain.Receta, error)
	Eliminar(ctx context.Context, id string) error
}

// VentaRepository persiste ventas.
type VentaRepository interface {
	Crear(ctx context.Context, v domain.Venta) (domain.Venta, error)
	CrearMultiples(ctx context.Context, vs []domain.Venta) ([]domain.Venta, error)
	ObtenerPorID(ctx context.Context, id string) (domain.Venta, error)
	Listar(ctx context.Context) ([]domain.Venta, error)
	ListarPorFecha(ctx context.Context, fecha domain.Date) ([]domain.Venta, error)
	Actualizar(ctx context.Context, v domain.Venta) (domain.Venta, error)
	Eliminar(ctx context.Context, id string) error
	EliminarMultiples(ctx context.Context, ids []string) error
}

// InventarioDiarioRepository persiste la hoja de IPV.
//
// Es el puerto que la versión Python NO tenía (los casos de uso dependían
// directamente de la clase SQLite).
type InventarioDiarioRepository interface {
	ListarPorFecha(ctx context.Context, fecha domain.Date) ([]domain.InventarioDiario, error)
	BuscarPorFechaAreaProducto(ctx context.Context, fecha domain.Date, areaID, productoID string) (domain.InventarioDiario, error)
	// FechasConRegistros devuelve las fechas distintas con datos, de más
	// reciente a más antigua.
	FechasConRegistros(ctx context.Context) ([]domain.Date, error)
	// FinalFisicoDiaAnterior devuelve el final_fisico del día previo para ese
	// producto y área, o 0 si no hay registro.
	FinalFisicoDiaAnterior(ctx context.Context, fecha domain.Date, areaID, productoID string) (float64, error)
	// GuardarTodos hace upsert por (fecha, area_id, producto_id).
	GuardarTodos(ctx context.Context, filas []domain.InventarioDiario) error
}

// ModeloIPVRepository persiste los modelos de IPV por área.
type ModeloIPVRepository interface {
	// Listar devuelve todas las filas de todos los modelos (sin agrupar).
	Listar(ctx context.Context) ([]domain.ModeloIPV, error)
	// GuardarModelo reemplaza por completo el modelo de un área.
	GuardarModelo(ctx context.Context, areaID string, filas []domain.ModeloIPV) error
}

// HistorialRepository persiste la auditoría de cambios.
type HistorialRepository interface {
	Registrar(ctx context.Context, h domain.HistorialCambios) error
	PorEntidad(ctx context.Context, tipo domain.TipoEntidad) ([]domain.HistorialCambios, error)
}

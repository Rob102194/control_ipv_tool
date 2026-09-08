package ports

import "context"

// Repos es el conjunto de repositorios disponibles dentro de una unidad de
// trabajo. Todos comparten la misma transacción.
type Repos interface {
	Productos() ProductoRepository
	Areas() AreaRepository
	Recetas() RecetaRepository
	Ventas() VentaRepository
	InventarioDiario() InventarioDiarioRepository
	ModelosIPV() ModeloIPVRepository
	Historial() HistorialRepository
}

// UnitOfWork ejecuta una función dentro de una transacción: confirma si fn
// devuelve nil y revierte si devuelve error (o entra en pánico).
//
// Los casos de uso con escrituras de varios pasos (crear receta + ingredientes,
// importar ventas + recetas, guardar la hoja de IPV) se envuelven en Do. Las
// lecturas simples pueden recibir el repositorio concreto directamente.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(r Repos) error) error
}

// Package ports declara las interfaces que la capa de aplicación necesita del
// exterior: repositorios de persistencia, unidad de trabajo (transacciones),
// reloj y generador de identificadores.
//
// Son "puertos" en el sentido de la arquitectura hexagonal. Los "adaptadores"
// que los implementan viven en internal/adapters/*. Este paquete solo depende de
// internal/core/domain.
//
// Contrato común de los repositorios:
//   - Todos los métodos reciben context.Context como primer parámetro.
//   - Las búsquedas por identidad devuelven ErrNoEncontrado si no hay fila.
//   - Los repositorios reciben y devuelven entidades de dominio, nunca filas ni
//     tipos de base de datos.
package ports

import "errors"

// ErrNoEncontrado lo devuelve cualquier búsqueda por identidad cuando no existe
// la fila. Los llamadores lo detectan con errors.Is.
var ErrNoEncontrado = errors.New("no encontrado")

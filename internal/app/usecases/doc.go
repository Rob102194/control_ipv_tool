// Package usecases contiene los casos de uso de la aplicación, agrupados por
// agregado en "servicios" (ProductoService, RecetaService, IPVService…).
//
// Solo depende de internal/core/domain y de internal/core/ports. No importa
// ningún adaptador concreto (SQLite, Excel, HTTP): recibe sus dependencias como
// interfaces de ports.
//
// Correcciones respecto a la versión Python:
//   - Ninguna dependencia de infraestructura concreta (Python importaba las
//     clases SQLite en los casos de uso del IPV).
//   - Las escrituras de varios pasos (receta + ingredientes, importaciones,
//     guardado del IPV, cambios + historial) se ejecutan dentro de
//     ports.UnitOfWork.Do, es decir, en una única transacción.
//   - ObtenerEstadoInventario y GenerarReporte trabajan con entidades de
//     dominio; el formateo de textos de presentación se deja para la capa HTTP.
//   - Los identificadores de las entidades nuevas los genera este paquete con
//     ports.IDGen; los repositorios nunca generan ids.
package usecases

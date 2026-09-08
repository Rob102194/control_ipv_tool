// Package domain contiene las entidades y reglas de negocio de Control IPV.
//
// Es el centro de la arquitectura hexagonal: no importa nada de infraestructura
// (ni base de datos, ni HTTP, ni frameworks) y no sabe cómo se serializa ni se
// persiste. Las interfaces que necesita del exterior viven en
// github.com/Rob102194/control_ipv_tool/internal/core/ports.
//
// Vocabulario (se mantiene en español, como en la versión Python):
//   - IPV: Inventario Por Venta. Hoja diaria por área y producto.
//   - final_teorico = (inicio + entradas) - consumo - merma - otras_salidas
//   - diferencia    = final_fisico - final_teorico
//   - consumo: se deriva de las ventas del día multiplicadas por los
//     ingredientes de cada receta.
package domain

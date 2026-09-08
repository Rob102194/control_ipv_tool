package domain

// ConsumoKey identifica el consumo acumulado de un producto en un área concreta.
type ConsumoKey struct {
	ProductoID string
	AreaID     string
}

// CalcularConsumo reparte las ventas de un día en consumo por producto y área,
// según los ingredientes de cada receta:
//
//	consumo[producto, área] = Σ  venta.cantidad × ingrediente.cantidad
//
// recetasPorNombre debe estar indexado por Receta.Nombre, tal cual aparece en
// Venta.RecetaNombre. Las ventas cuya receta no esté en el mapa se ignoran, igual
// que en la versión Python (que hacía find_by_name y saltaba si no existía).
//
// Es una función pura: no conoce repositorios. El caso de uso arma el mapa de
// recetas y llama aquí.
func CalcularConsumo(ventas []Venta, recetasPorNombre map[string]Receta) map[ConsumoKey]float64 {
	consumo := make(map[ConsumoKey]float64)
	for _, v := range ventas {
		receta, ok := recetasPorNombre[v.RecetaNombre]
		if !ok {
			continue
		}
		for _, ing := range receta.Ingredientes {
			k := ConsumoKey{ProductoID: ing.ProductoID, AreaID: ing.AreaID}
			consumo[k] += float64(v.Cantidad) * ing.Cantidad
		}
	}
	return consumo
}

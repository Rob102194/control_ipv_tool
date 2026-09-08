package domain

import "sort"

// ModeloIPV define que un producto forma parte de la hoja de IPV de un área, y en
// qué posición. El "modelo" de un área es el conjunto ordenado de estas filas.
type ModeloIPV struct {
	ID         string
	AreaID     string
	ProductoID string
	Orden      int
}

// OrdenarModelo ordena in situ una lista de filas de modelo por su campo Orden.
func OrdenarModelo(filas []ModeloIPV) {
	sort.SliceStable(filas, func(a, b int) bool { return filas[a].Orden < filas[b].Orden })
}

// ReindexarModelo reasigna Orden = 0,1,2… respetando el orden actual de la lista.
// Se usa al guardar un modelo para normalizar la numeración.
func ReindexarModelo(filas []ModeloIPV) {
	for i := range filas {
		filas[i].Orden = i
	}
}

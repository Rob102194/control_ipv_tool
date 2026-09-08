package domain

// InventarioDiario es una fila del IPV: el movimiento de un producto en un área
// durante una fecha.
//
// A diferencia de la versión Python, la entidad NO lleva producto_nombre ni
// area_nombre: esos campos son decoración de lectura y viven en el DTO de salida.
type InventarioDiario struct {
	ID           string
	Fecha        Date
	AreaID       string
	ProductoID   string
	Inicio       float64
	Entradas     float64
	Consumo      float64
	Merma        float64
	OtrasSalidas float64
	FinalFisico  float64
	FinalTeorico float64
	Diferencia   float64
	// Comentario es una cadena JSON (mapa campo -> texto) o "". La entidad la
	// trata como texto opaco; interpretarla es tarea del reporte.
	Comentario string
}

// CalcularDiferencias recalcula final_teorico y diferencia a partir del resto de
// campos. El orden de las operaciones se conserva idéntico al de la versión
// Python para obtener exactamente el mismo resultado en coma flotante:
//
//	final_teorico = (inicio + entradas) - consumo - merma - otras_salidas
//	diferencia    = final_fisico - final_teorico
func (i *InventarioDiario) CalcularDiferencias() {
	i.FinalTeorico = (i.Inicio + i.Entradas) - i.Consumo - i.Merma - i.OtrasSalidas
	i.Diferencia = i.FinalFisico - i.FinalTeorico
}

// EsFaltante indica que el conteo físico quedó por debajo del teórico.
func (i InventarioDiario) EsFaltante() bool { return i.Diferencia < 0 }

// EsSobrante indica que el conteo físico superó al teórico.
func (i InventarioDiario) EsSobrante() bool { return i.Diferencia > 0 }

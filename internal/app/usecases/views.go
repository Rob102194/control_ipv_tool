package usecases

import "github.com/Rob102194/control_ipv_tool/internal/core/domain"

// InventarioFilaView es una fila de IPV enriquecida con los nombres de producto
// y área. Es el modelo de lectura de la aplicación: el dominio no carga con esos
// nombres (eran decoración en la versión Python). La capa HTTP la serializa tal
// cual (equivale al to_dict de InventarioDiario).
type InventarioFilaView struct {
	ID             string
	Fecha          domain.Date
	AreaID         string
	ProductoID     string
	Inicio         float64
	Entradas       float64
	Consumo        float64
	Merma          float64
	OtrasSalidas   float64
	FinalFisico    float64
	FinalTeorico   float64
	Diferencia     float64
	ProductoNombre string
	AreaNombre     string
	Comentario     string
}

func (v InventarioFilaView) toDomain() domain.InventarioDiario {
	return domain.InventarioDiario{
		ID: v.ID, Fecha: v.Fecha, AreaID: v.AreaID, ProductoID: v.ProductoID,
		Inicio: v.Inicio, Entradas: v.Entradas, Consumo: v.Consumo, Merma: v.Merma,
		OtrasSalidas: v.OtrasSalidas, FinalFisico: v.FinalFisico,
		FinalTeorico: v.FinalTeorico, Diferencia: v.Diferencia,
		Comentario: v.Comentario,
	}
}

func viewFromDomain(d domain.InventarioDiario, productoNombre, areaNombre string) InventarioFilaView {
	return InventarioFilaView{
		ID: d.ID, Fecha: d.Fecha, AreaID: d.AreaID, ProductoID: d.ProductoID,
		Inicio: d.Inicio, Entradas: d.Entradas, Consumo: d.Consumo, Merma: d.Merma,
		OtrasSalidas: d.OtrasSalidas, FinalFisico: d.FinalFisico,
		FinalTeorico: d.FinalTeorico, Diferencia: d.Diferencia,
		ProductoNombre: productoNombre, AreaNombre: areaNombre,
		Comentario: d.Comentario,
	}
}

// ModeloItem es una entrada del modelo de IPV de un área.
type ModeloItem struct {
	ProductoID string
	Orden      int
}

// --- Reporte -----------------------------------------------------------

// ReporteIPV son los datos estructurados de un reporte de IPV. El formateo de
// los textos (p. ej. "ACEITE: 0.2 L") es tarea de la capa HTTP.
type ReporteIPV struct {
	Fecha   domain.Date
	Areas   map[string][]ReporteFila
	Resumen map[string]ReporteResumenArea
	Notas   []ReporteNota
	// OrdenAreas conserva el orden en que aparecieron las áreas.
	OrdenAreas []string
}

// ReporteFila es una fila del reporte por área.
type ReporteFila struct {
	Producto     string
	UM           string
	Inicio       float64
	Entradas     float64
	Consumo      float64
	Merma        float64
	OtrasSalidas float64
	FinalTeorico float64
	FinalFisico  float64
	Diferencia   float64
}

// ReporteResumenArea agrupa las diferencias y mermas de un área.
type ReporteResumenArea struct {
	Faltantes []ReporteDelta
	Sobrantes []ReporteDelta
	Mermas    []ReporteDelta
}

// ReporteDelta es un producto con una cantidad y su unidad de medida.
type ReporteDelta struct {
	Producto string
	Cantidad float64
	UM       string
}

// ReporteNota es un comentario de una fila del IPV.
type ReporteNota struct {
	Producto string
	Campo    string // p. ej. "merma", "diferencia"; "" si el comentario era texto plano
	Texto    string
}

// ImportResult resume una importación (creados vs. omitidos).
type ImportResult struct {
	Creados  int
	Omitidos int
}

// ImportVentasResult es lo que devuelve la importación de ventas.
type ImportVentasResult struct {
	Ventas        []domain.Venta
	NuevasRecetas []domain.Receta
}

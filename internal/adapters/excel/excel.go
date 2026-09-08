// Package excel es el adaptador de importación y exportación en formato .xlsx.
//
// Solo se ocupa de la mecánica del fichero: convierte entre bytes y filas
// tipadas. Resolver productos/áreas/recetas y persistir es tarea de los casos de
// uso. Reproduce los contratos de columnas de la versión Python (pandas +
// openpyxl); ver las notas de paridad en MIGRATION.md.
package excel

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// --- Filas tipadas ------------------------------------------------------

// ProductoRow es una fila de la hoja de productos (columnas: nombre, unidad_medida).
type ProductoRow struct {
	Nombre       string
	UnidadMedida string
}

// RecetaRow es una fila de la hoja de recetas. Columnas:
// receta_nombre, producto_nombre, unidad_medida, cantidad, area_nombre.
// Una receta sin ingredientes aparece como una única fila con
// ProductoNombre == "" (y el resto de campos de ingrediente vacíos).
type RecetaRow struct {
	RecetaNombre   string
	ProductoNombre string
	UnidadMedida   string
	AreaNombre     string
	Cantidad       float64
}

// VentaRow es una fila de la hoja de ventas (columnas: Nombre, Cantidad).
type VentaRow struct {
	// Fila es el número de fila en la hoja (cabecera = 1, primera fila de
	// datos = 2), para reproducir los mensajes de error de la versión Python.
	Fila           int
	Nombre         string
	Cantidad       float64
	CantidadValida bool
	// RawCantidad es el texto original de la celda Cantidad, para reproducir el
	// mensaje "Cantidad inválida (<raw>)" de la versión Python.
	RawCantidad string
}

// --- Import -----------------------------------------------------------

// ParseProductos lee la hoja de productos. Error si faltan columnas obligatorias.
func ParseProductos(r io.Reader) ([]ProductoRow, error) {
	table, err := readSheet(r)
	if err != nil {
		return nil, err
	}
	idx, err := table.requireColumns("nombre", "unidad_medida")
	if err != nil {
		return nil, err
	}
	var out []ProductoRow
	for _, row := range table.data {
		nombre := cell(row, idx["nombre"])
		if strings.TrimSpace(nombre) == "" {
			continue
		}
		out = append(out, ProductoRow{
			Nombre:       nombre,
			UnidadMedida: cell(row, idx["unidad_medida"]),
		})
	}
	return out, nil
}

// ParseRecetas lee la hoja de recetas. Error si faltan columnas obligatorias.
func ParseRecetas(r io.Reader) ([]RecetaRow, error) {
	table, err := readSheet(r)
	if err != nil {
		return nil, err
	}
	idx, err := table.requireColumns("receta_nombre", "producto_nombre", "unidad_medida", "cantidad", "area_nombre")
	if err != nil {
		return nil, err
	}
	var out []RecetaRow
	for _, row := range table.data {
		rn := strings.TrimSpace(cell(row, idx["receta_nombre"]))
		if rn == "" {
			continue
		}
		cant, _ := parseNumber(cell(row, idx["cantidad"]))
		out = append(out, RecetaRow{
			RecetaNombre:   cell(row, idx["receta_nombre"]),
			ProductoNombre: cell(row, idx["producto_nombre"]),
			UnidadMedida:   cell(row, idx["unidad_medida"]),
			AreaNombre:     cell(row, idx["area_nombre"]),
			Cantidad:       cant,
		})
	}
	return out, nil
}

// ParseVentas lee la hoja de ventas (columnas 'Nombre' y 'Cantidad', con esa
// capitalización, igual que la versión Python).
func ParseVentas(r io.Reader) ([]VentaRow, error) {
	table, err := readSheet(r)
	if err != nil {
		return nil, err
	}
	if !table.hasColumn("Nombre") || !table.hasColumn("Cantidad") {
		return nil, fmt.Errorf("el archivo debe contener las columnas 'Nombre' y 'Cantidad'")
	}
	idx, _ := table.requireColumns("Nombre", "Cantidad")

	var out []VentaRow
	for i, row := range table.data {
		nombre := strings.TrimSpace(cell(row, idx["Nombre"]))
		if nombre == "" {
			continue
		}
		raw := cell(row, idx["Cantidad"])
		cant, ok := parseNumber(raw)
		out = append(out, VentaRow{
			Fila:           i + 2, // +1 por índice 0, +1 por la cabecera
			Nombre:         nombre,
			Cantidad:       cant,
			CantidadValida: ok && cant > 0,
			RawCantidad:    raw,
		})
	}
	return out, nil
}

// --- Export ---------------------------------------------------------

// WriteProductos escribe la hoja 'Productos' con columnas nombre, unidad_medida.
func WriteProductos(w io.Writer, rows []ProductoRow) error {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Productos"
	f.SetSheetName(f.GetSheetName(0), sheet)
	if err := writeRow(f, sheet, 1, "nombre", "unidad_medida"); err != nil {
		return err
	}
	for i, r := range rows {
		if err := writeRow(f, sheet, i+2, r.Nombre, r.UnidadMedida); err != nil {
			return err
		}
	}
	return f.Write(w)
}

// WriteRecetas escribe la hoja 'Recetas'. El caso de uso arma las filas
// (una por ingrediente; receta sin ingredientes -> una fila con campos vacíos).
func WriteRecetas(w io.Writer, rows []RecetaRow) error {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Recetas"
	f.SetSheetName(f.GetSheetName(0), sheet)
	if err := writeRow(f, sheet, 1, "receta_nombre", "producto_nombre", "unidad_medida", "cantidad", "area_nombre"); err != nil {
		return err
	}
	for i, r := range rows {
		cant := ""
		if r.ProductoNombre != "" {
			cant = strconv.FormatFloat(r.Cantidad, 'g', -1, 64)
		}
		if err := writeRow(f, sheet, i+2, r.RecetaNombre, r.ProductoNombre, r.UnidadMedida, cant, r.AreaNombre); err != nil {
			return err
		}
	}
	return f.Write(w)
}

// --- Interno ------------------------------------------------------

type sheetTable struct {
	header []string
	data   [][]string
}

func readSheet(r io.Reader) (sheetTable, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return sheetTable{}, fmt.Errorf("no se pudo leer el archivo Excel: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return sheetTable{}, fmt.Errorf("el archivo Excel no tiene hojas")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return sheetTable{}, fmt.Errorf("no se pudieron leer las filas: %w", err)
	}
	if len(rows) == 0 {
		return sheetTable{}, fmt.Errorf("la hoja está vacía")
	}
	header := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		header[i] = strings.TrimSpace(h)
	}
	return sheetTable{header: header, data: rows[1:]}, nil
}

func (t sheetTable) hasColumn(name string) bool {
	for _, h := range t.header {
		if h == name {
			return true
		}
	}
	return false
}

func (t sheetTable) requireColumns(names ...string) (map[string]int, error) {
	idx := make(map[string]int, len(names))
	var faltan []string
	for _, n := range names {
		pos := -1
		for i, h := range t.header {
			if h == n {
				pos = i
				break
			}
		}
		if pos < 0 {
			faltan = append(faltan, n)
			continue
		}
		idx[n] = pos
	}
	if len(faltan) > 0 {
		return nil, fmt.Errorf("El archivo Excel debe contener las columnas: %s", strings.Join(names, ", "))
	}
	return idx, nil
}

func cell(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func writeRow(f *excelize.File, sheet string, row int, vals ...string) error {
	cellRef, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return err
	}
	anyVals := make([]any, len(vals))
	for i, v := range vals {
		anyVals[i] = v
	}
	return f.SetSheetRow(sheet, cellRef, &anyVals)
}

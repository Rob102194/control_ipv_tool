package domain

import (
	"fmt"
	"time"
)

// Date es una fecha de calendario sin hora ni zona horaria. Se corresponde con
// las columnas DATE de la versión Python (ventas.fecha, inventario_diario.fecha),
// que en SQLite se guardan como texto "YYYY-MM-DD".
//
// Usar un tipo propio (en lugar de time.Time) evita de raíz los desplazamientos
// por zona horaria: aquí no hay hora que desplazar.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

const dateLayout = "2006-01-02"

// ParseDate interpreta una cadena "YYYY-MM-DD".
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("fecha inválida %q (se esperaba YYYY-MM-DD): %w", s, err)
	}
	return DateFromTime(t), nil
}

// MustParseDate es como ParseDate pero entra en pánico ante un error. Solo para
// literales de test.
func MustParseDate(s string) Date {
	d, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

// DateFromTime toma la parte de calendario de un time.Time (en su propia zona).
func DateFromTime(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Year: y, Month: m, Day: d}
}

// Time devuelve el instante de medianoche UTC de esta fecha. Útil para aritmética.
func (d Date) Time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

// String devuelve la representación "YYYY-MM-DD".
func (d Date) String() string {
	return d.Time().Format(dateLayout)
}

// IsZero indica si la fecha no se ha fijado.
func (d Date) IsZero() bool { return d == Date{} }

// AddDays devuelve la fecha desplazada n días (n puede ser negativo).
func (d Date) AddDays(n int) Date {
	return DateFromTime(d.Time().AddDate(0, 0, n))
}

// Before, After y Equal comparan por orden de calendario.
func (d Date) Before(o Date) bool { return d.Time().Before(o.Time()) }
func (d Date) After(o Date) bool  { return d.Time().After(o.Time()) }
func (d Date) Equal(o Date) bool  { return d == o }

// MarshalText / UnmarshalText hacen que Date se serialice como "YYYY-MM-DD"
// tanto en JSON como en cualquier codificador basado en texto. Es la única
// representación válida del tipo, así que vive con el tipo (no es lógica de
// presentación, igual que time.Time trae la suya).
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(b []byte) error {
	parsed, err := ParseDate(string(b))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

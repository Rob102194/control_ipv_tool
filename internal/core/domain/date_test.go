package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2026-09-05")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if d.Year != 2026 || d.Month != time.September || d.Day != 5 {
		t.Fatalf("ParseDate = %+v", d)
	}
	if d.String() != "2026-09-05" {
		t.Fatalf("String() = %q", d.String())
	}
	for _, bad := range []string{"", "2026-13-01", "05/09/2026", "2026-09-5", "hoy"} {
		if _, err := ParseDate(bad); err == nil {
			t.Errorf("ParseDate(%q) no dio error", bad)
		}
	}
}

func TestDateAddDays(t *testing.T) {
	casos := []struct {
		from string
		n    int
		want string
	}{
		{"2026-09-05", -1, "2026-09-04"},
		{"2026-09-01", -1, "2026-08-31"}, // cruce de mes (arrastre de `inicio`)
		{"2026-01-01", -1, "2025-12-31"}, // cruce de año
		{"2024-02-28", 1, "2024-02-29"},  // año bisiesto
		{"2026-09-05", 0, "2026-09-05"},
	}
	for _, c := range casos {
		got := MustParseDate(c.from).AddDays(c.n).String()
		if got != c.want {
			t.Errorf("%s.AddDays(%d) = %s, se esperaba %s", c.from, c.n, got, c.want)
		}
	}
}

func TestDateCompare(t *testing.T) {
	a := MustParseDate("2026-09-04")
	b := MustParseDate("2026-09-05")
	if !a.Before(b) || !b.After(a) || a.Equal(b) || !a.Equal(a) {
		t.Fatal("comparadores de Date incoherentes")
	}
}

func TestDateJSON(t *testing.T) {
	type wrap struct {
		Fecha Date `json:"fecha"`
	}
	in := wrap{Fecha: MustParseDate("2026-09-05")}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != `{"fecha":"2026-09-05"}` {
		t.Fatalf("JSON = %s", b)
	}

	var out wrap
	if err := json.Unmarshal([]byte(`{"fecha":"2026-08-31"}`), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Fecha.String() != "2026-08-31" {
		t.Fatalf("Unmarshal -> %s", out.Fecha)
	}
	if err := json.Unmarshal([]byte(`{"fecha":"nope"}`), &out); err == nil {
		t.Error("Unmarshal de fecha inválida no dio error")
	}
}

func TestDateZero(t *testing.T) {
	var d Date
	if !d.IsZero() {
		t.Fatal("Date cero debería ser IsZero()")
	}
	if MustParseDate("2026-09-05").IsZero() {
		t.Fatal("fecha con valor no debería ser IsZero()")
	}
}

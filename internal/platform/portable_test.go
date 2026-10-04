package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPortableDataDirFor_Writable(t *testing.T) {
	base := t.TempDir()
	exePath := filepath.Join(base, "restaurante1", "Control IPV.exe")
	if err := os.MkdirAll(filepath.Dir(exePath), 0o755); err != nil {
		t.Fatal(err)
	}

	dir, ok := portableDataDirFor(exePath)
	if !ok {
		t.Fatal("se esperaba modo portable activado (carpeta escribible)")
	}
	want := filepath.Join(base, "restaurante1", "data")
	if dir != want {
		t.Fatalf("dir = %q, se esperaba %q", dir, want)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("el directorio de datos no se creó: %v", err)
	}
}

// TestPortableDataDirFor_NoEscribible simula una ubicación de solo lectura
// (el caso real: Program Files) sin depender de bits de permiso POSIX, que en
// Windows no reflejan ACLs fielmente: en vez de eso, un segmento de la ruta
// es un ARCHIVO, no un directorio, así que crear <esa ruta>/data falla
// igual en cualquier sistema operativo.
func TestPortableDataDirFor_NoEscribible(t *testing.T) {
	base := t.TempDir()
	bloqueador := filepath.Join(base, "bloqueador")
	if err := os.WriteFile(bloqueador, []byte("soy un archivo, no un directorio"), 0o644); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(bloqueador, "sub", "Control IPV.exe")

	dir, ok := portableDataDirFor(exePath)
	if ok {
		t.Fatalf("se esperaba modo portable desactivado, pero dio dir=%q", dir)
	}
}

func TestIsInsideMacAppBundle(t *testing.T) {
	sep := string(filepath.Separator)
	casos := []struct {
		ruta    string
		adentro bool
	}{
		{"/Applications/Control IPV.app" + sep + "Contents" + sep + "MacOS" + sep + "Control IPV", true},
		{"/Users/roma/Descargas/Control IPV.exe", false},
		{"C:\\Users\\roma\\Desktop\\Restaurante1\\Control IPV.exe", false},
	}
	for _, c := range casos {
		if got := isInsideMacAppBundle(c.ruta); got != c.adentro {
			t.Errorf("isInsideMacAppBundle(%q) = %v, se esperaba %v", c.ruta, got, c.adentro)
		}
	}
}

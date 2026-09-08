package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testDeps() Deps {
	return Deps{
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		SchemaVersion: func() (int64, error) { return 1, nil },
	}
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(NewRouter(testDeps()))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba 200", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decodificando cuerpo: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %v, se esperaba \"ok\"", body["status"])
	}
	if body["schema_version"].(float64) != 1 {
		t.Fatalf("schema_version = %v, se esperaba 1", body["schema_version"])
	}
}

func TestUnknownRouteIs404(t *testing.T) {
	srv := httptest.NewServer(NewRouter(testDeps()))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/no-existe")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, se esperaba 404", resp.StatusCode)
	}
}

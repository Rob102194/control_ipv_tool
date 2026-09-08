// Package web embebe el SPA compilado (frontend/) y lo sirve con fallback a
// index.html para las rutas del router del lado cliente.
//
// El contenido de dist/ lo genera `npm run build` en frontend/ (vite.config.js
// apunta outDir a ../web/dist). Antes del primer build hay un index.html
// provisional para que `go build` funcione igual.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler devuelve un http.Handler que sirve los ficheros de dist/ y, para
// cualquier ruta que no sea un fichero existente, devuelve index.html (SPA).
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			// No es un fichero: lo resuelve el enrutador de React.
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			serveIndex(w, r, sub)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, sub fs.FS) {
	b, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		http.Error(w, "SPA no compilado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

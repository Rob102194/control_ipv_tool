package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Rob102194/control_ipv_tool/internal/platform"
)

// BackupManager es el contrato que necesita la capa HTTP para /api/backups.
// Lo implementa appboot.backupManager; se define aquí como interfaz (en vez
// de importar el tipo concreto) para no crear un ciclo de imports: appboot
// importa httpapi para construir el router.
type BackupManager interface {
	Crear(ctx context.Context) (platform.BackupInfo, error)
	Listar(ctx context.Context) ([]platform.BackupInfo, error)
	RutaArchivo(ctx context.Context, nombre string) (string, error)
	Restaurar(ctx context.Context, nombre string) error
}

type backupDTO struct {
	Nombre      string `json:"nombre"`
	CreadoEn    string `json:"creado_en"`
	TamanoBytes int64  `json:"tamano_bytes"`
}

func toBackupDTO(b platform.BackupInfo) backupDTO {
	return backupDTO{Nombre: b.Nombre, CreadoEn: b.CreadoEn.UTC().Format(time.RFC3339), TamanoBytes: b.TamanoBytes}
}

func (a *api) backupsList(w http.ResponseWriter, r *http.Request) {
	backups, err := a.backups.Listar(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	out := make([]backupDTO, len(backups))
	for i, b := range backups {
		out[i] = toBackupDTO(b)
	}
	respondJSON(w, http.StatusOK, out)
}

func (a *api) backupsCreate(w http.ResponseWriter, r *http.Request) {
	info, err := a.backups.Crear(r.Context())
	if err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, toBackupDTO(info))
}

// backupsDownload sirve el fichero crudo (no pasa por writeJSON): la ruta ya
// viene validada/resuelta por BackupManager.RutaArchivo (rechaza path
// traversal y nombres que no existan), así que es segura pasarla a
// http.ServeFile.
func (a *api) backupsDownload(w http.ResponseWriter, r *http.Request) {
	nombre := chi.URLParam(r, "nombre")
	ruta, err := a.backups.RutaArchivo(r.Context(), nombre)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, nombre))
	http.ServeFile(w, r, ruta)
}

func (a *api) backupsRestore(w http.ResponseWriter, r *http.Request) {
	nombre := chi.URLParam(r, "nombre")
	if err := a.backups.Restaurar(r.Context(), nombre); err != nil {
		a.fail(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Base de datos restaurada. Cierra y vuelve a abrir la aplicación para continuar.",
	})
}

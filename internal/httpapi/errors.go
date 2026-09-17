// Package httpapi expone la capa de entrega HTTP: router, middleware, DTOs y el
// mapeo de errores de dominio a códigos de estado. Es una de varias entregas
// posibles sobre los mismos casos de uso (escritorio vía Wails, servidor web).
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Rob102194/control_ipv_tool/internal/core/domain"
	"github.com/Rob102194/control_ipv_tool/internal/core/ports"
)

// statusFor traduce un error a (código HTTP, mensaje para el cliente).
//
// Reconoce los errores del dominio (*domain.ValidationError -> 422,
// *domain.ConflictError -> 409, *domain.NotFoundError -> 404) y el centinela
// ports.ErrNoEncontrado -> 404.
func statusFor(err error) (int, string) {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		return http.StatusUnprocessableEntity, ve.Msg
	}
	var ce *domain.ConflictError
	if errors.As(err, &ce) {
		return http.StatusConflict, ce.Msg
	}
	var nf *domain.NotFoundError
	if errors.As(err, &nf) {
		return http.StatusNotFound, nf.Msg
	}
	if errors.Is(err, ports.ErrNoEncontrado) {
		return http.StatusNotFound, "Recurso no encontrado"
	}

	return http.StatusInternalServerError, "Error interno del servidor"
}

// writeJSON serializa v como JSON con el código indicado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError responde con {"error": "..."} y registra la causa si es interna.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	status, msg := statusFor(err)
	if status >= 500 {
		logger.ErrorContext(r.Context(), "error atendiendo petición",
			"method", r.Method, "path", r.URL.Path, "err", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

// Package httpapi expone la capa de entrega HTTP: router, middleware, DTOs y el
// mapeo de errores de dominio a códigos de estado. Es una de varias entregas
// posibles sobre los mismos casos de uso (escritorio vía Wails, servidor web).
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Kind clasifica un error de negocio para traducirlo a un código HTTP.
type Kind int

const (
	KindInternal   Kind = iota // 500 — no se filtra el detalle al cliente
	KindValidation             // 422 — entrada inválida
	KindNotFound               // 404 — recurso inexistente
	KindConflict               // 409 — viola una invariante (p. ej. nombre duplicado)
)

// Error es el error tipado que la capa de aplicación devuelve hacia HTTP.
// En la Fase 5 los casos de uso devolverán estos valores; por ahora define el
// contrato y el mapeo.
type Error struct {
	Kind    Kind
	Message string // mensaje apto para mostrar al usuario
	Err     error  // causa subyacente, solo para logs
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// Constructores.
func Validation(msg string) *Error { return &Error{Kind: KindValidation, Message: msg} }
func NotFound(msg string) *Error   { return &Error{Kind: KindNotFound, Message: msg} }
func Conflict(msg string) *Error   { return &Error{Kind: KindConflict, Message: msg} }
func Internal(err error) *Error {
	return &Error{Kind: KindInternal, Message: "Error interno del servidor", Err: err}
}

// statusFor traduce un error a (código HTTP, mensaje para el cliente).
func statusFor(err error) (int, string) {
	var e *Error
	if errors.As(err, &e) {
		switch e.Kind {
		case KindValidation:
			return http.StatusUnprocessableEntity, e.Message
		case KindNotFound:
			return http.StatusNotFound, e.Message
		case KindConflict:
			return http.StatusConflict, e.Message
		}
		return http.StatusInternalServerError, "Error interno del servidor"
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

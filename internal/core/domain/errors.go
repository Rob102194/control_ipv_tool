package domain

import "fmt"

// ValidationError señala una entrada que viola una invariante de negocio.
// La capa HTTP lo traducirá a 422.
type ValidationError struct {
	Campo string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e.Campo != "" {
		return fmt.Sprintf("%s: %s", e.Campo, e.Msg)
	}
	return e.Msg
}

// ConflictError señala una operación que choca con el estado actual (por ejemplo
// un nombre único ya usado). La capa HTTP lo traducirá a 409.
type ConflictError struct {
	Msg string
}

func (e *ConflictError) Error() string { return e.Msg }

// NotFoundError señala un recurso inexistente con un mensaje propio (cuando el
// centinela ports.ErrNoEncontrado se queda corto). La capa HTTP lo traducirá a 404.
type NotFoundError struct {
	Msg string
}

func (e *NotFoundError) Error() string { return e.Msg }

// Constructores de conveniencia.
func Invalid(campo, msg string) error { return &ValidationError{Campo: campo, Msg: msg} }
func Conflictf(format string, a ...any) error {
	return &ConflictError{Msg: fmt.Sprintf(format, a...)}
}
func NotFoundf(format string, a ...any) error {
	return &NotFoundError{Msg: fmt.Sprintf(format, a...)}
}

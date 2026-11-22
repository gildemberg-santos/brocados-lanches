package domain

import (
	"errors"
	"fmt"
)

// Erros sentinela devolvidos pelo domínio e traduzidos pela camada HTTP.
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

// Códigos estáveis de validação usados no envelope de erro da API.
const (
	CodeRequired   = "REQUIRED"
	CodeInvalid    = "INVALID"
	CodeOutOfRange = "OUT_OF_RANGE"
	CodeTooLong    = "TOO_LONG"
	CodeDuplicate  = "DUPLICATE"
)

// ValidationError descreve um único campo inválido.
type ValidationError struct {
	Field   string
	Code    string
	Message string
}

// Error implementa error.
func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap permite que chamadores reconheçam erros de validação com errors.Is.
func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

func invalid(field, code, message string) error {
	return &ValidationError{
		Field:   field,
		Code:    code,
		Message: message,
	}
}

// prefixValidationError adiciona o caminho do pai ao campo de um erro de validação.
func prefixValidationError(prefix string, err error) error {
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		return err
	}

	prefixed := *validationError
	if prefixed.Field == "" {
		prefixed.Field = prefix
	} else {
		prefixed.Field = prefix + "." + prefixed.Field
	}
	return &prefixed
}

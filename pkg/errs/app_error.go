// pkg/errs/app_error.go
package errs

import (
	"fmt"
	"net/http"
)

// AppError adalah custom error struct kita
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"` // error asli (untuk logging, tidak dikirim ke json)
}

// Implement interface error bawaan Go
func (e *AppError) Error() string {
	return e.Message
}

// --- Factory Functions (Mirip NestJS Exceptions) ---

func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewNotFound(message string) *AppError {
	defaultMessage := "DATA_NOT_FOUND"
	if message == "" {
		message = defaultMessage
	}
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NewForbidden(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

func NewInternalServerError(err error) *AppError {
	fmt.Println(err.Error())
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "INTERNAL_SERVER_ERROR",
		Err:     err, // Simpan error asli untuk keperluan logging nanti
	}
}

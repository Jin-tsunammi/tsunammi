package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

var ErrInvalidKeys = errors.New("invalid API, Secret key or Passphrase")

type AppError struct {
	BaseError       error  `swaggerignore:"true"`
	Message         string `json:"message"`
	HTTPStatus      int    `json:"http_status"`
	FileErrOccurred string `swaggerignore:"true"`
	LineErrOccurred int    `swaggerignore:"true"`
}

func newAppError(httpStatus int, message string, err ...error) *AppError {
	var baseError error
	if len(err) == 0 {
		baseError = nil
	} else {
		baseError = err[0]
	}

	var e *AppError
	ok := errors.As(baseError, &e)
	if ok {
		return &AppError{
			BaseError:       fmt.Errorf("%w: %w", baseError, e.BaseError),
			Message:         e.Message,
			HTTPStatus:      httpStatus,
			FileErrOccurred: e.FileErrOccurred,
			LineErrOccurred: e.LineErrOccurred,
		}
	}

	_, fileErrOccurred, lineErrOccurred, ok := runtime.Caller(2)
	if !ok {
		fileErrOccurred = ""
		lineErrOccurred = 0
	}

	appErr := &AppError{
		BaseError:       baseError,
		Message:         message,
		HTTPStatus:      httpStatus,
		FileErrOccurred: fileErrOccurred,
		LineErrOccurred: lineErrOccurred,
	}

	return appErr
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Path() string {
	return fmt.Sprintf("%s:%d", e.FileErrOccurred, e.LineErrOccurred)
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)

	return appErr, ok
}

func Unauthorized(message string, err ...error) error {
	return newAppError(http.StatusUnauthorized, message, err...)
}

func Internal(message string, err ...error) error {
	return newAppError(http.StatusInternalServerError, message, err...)
}

func BadRequest(message string, err ...error) error {
	return newAppError(http.StatusBadRequest, message, err...)
}

func ServiceUnavailable(message string, err ...error) error {
	return newAppError(http.StatusServiceUnavailable, message, err...)
}

const (
	StatusLoginTimeout = 440
)

func LoginTimeout(message string, err ...error) error {
	return newAppError(StatusLoginTimeout, message, err...)
}

func NotFound(message string, err ...error) error {
	return newAppError(http.StatusNotFound, message, err...)
}

func AlreadyExist(message string, err ...error) error {
	return newAppError(http.StatusConflict, message, err...)
}

// Teapot indicates that handler is a teapot and unable to brew a coffee
func Teapot(message string, err ...error) error {
	return newAppError(http.StatusTeapot, message, err...)
}

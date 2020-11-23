package apierror

import (
	"fmt"
)

// Error object for zale-api errors
type Error struct {
	HTTPStatus int    `json:"status,omitempty"`
	Message    string `json:"message"`
	APICode    Code   `json:"api_code"`
}

func (e Error) Error() string {
	return fmt.Sprintf("HTTPStatus: %v, Message: %v, Code: %v", e.HTTPStatus, e.Message, e.APICode)
}

// New creates a new Error
func New(httpStatusCode int, message string, code Code) *Error {
	return &Error{
		HTTPStatus: httpStatusCode,
		Message:    message,
		APICode:    code,
	}
}

// FromError creates a new Error from an error
func FromError(err error) *Error {
	return &Error{
		HTTPStatus: 0,
		Message:    err.Error(),
		APICode:    CodeUnknown,
	}
}

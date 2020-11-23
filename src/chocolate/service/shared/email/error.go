package email

import (
	"fmt"
)

// Error structue to handle Email related errors
type Error struct {
	Message string
	Inner   error
}

// NewError creates a new EmailError
func NewError(msg string, inner error) *Error {
	return &Error{
		Message: msg,
		Inner:   inner,
	}
}

// Error returns error string
func (e Error) Error() string {
	return fmt.Sprintf("EmailError { message: '%s', inner: '%+v' }", e.Message, e.Inner)
}

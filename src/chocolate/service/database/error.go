package database

import (
	"fmt"
)

type errorCode int

const (
	ErrorInternal      = errorCode(0)
	ErrorGeneric       = errorCode(1)
	ErrorCreation      = errorCode(2)
	ErrorTableCreation = errorCode(3)
	ErrorModelInvalid  = errorCode(4)
	ErrorAlreadyExists = errorCode(5)
	ErrorNoData        = errorCode(6)
	// ErrorExecute for any errors made by a bad query
	ErrorExecute = errorCode(6)
	ErrorNoRows  = errorCode(7)
)

// Error holds DB errors
type Error struct {
	Message string
	Query   string
	Table   string
	Inner   error
	Code    errorCode
}

func (e Error) Error() string {
	return fmt.Sprintf("Code: %v, Message: %s, Table: %s, Query: %s, Inner error: %v", e.Code, e.Message, e.Table, e.Query, e.Inner)
}

// NewError returns a  new db Errors
func NewError(code errorCode, message, query, table string, inner error) *Error {
	return &Error{message, query, table, inner, code}
}

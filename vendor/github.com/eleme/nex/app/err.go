package app

import (
	"fmt"
	"reflect"

	json "github.com/json-iterator/go"
)

// ErrorCode is type for error code.
type ErrorCode int64

const (
	// ErrorCodeUnknownError is 0
	ErrorCodeUnknownError ErrorCode = 0
)

// String implements the native format for ErrorCode.
func (c ErrorCode) String() string {
	switch c {
	case ErrorCodeUnknownError:
		return "UNKNOWN_ERROR"
	}
	return "<UNSET>"
}

// MarshalJSON returns error code as the JSON encoding data.
func (c ErrorCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnknownException represents a unknown exception.
type UnknownException struct {
	ErrorCode ErrorCode `json:"error_code"`
	ErrorName string    `json:"error_name"`
	Message   string    `json:"message"`
}

// NewUnknownException creates a unknown exception with name and message.
func NewUnknownException(name, msg string) *UnknownException {
	return &UnknownException{
		ErrorCode: ErrorCodeUnknownError,
		ErrorName: name,
		Message:   msg,
	}
}

// String returns the native firmat.
func (p *UnknownException) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("UnknownException(%+v)", *p)
}

// Error returns the error message.
func (p *UnknownException) Error() string {
	return p.String()
}

// ThriftUnknownExceptionFrom convert the fake UnknownException into real thrift UnknownException.
func ThriftUnknownExceptionFrom(e *UnknownException, typ reflect.Type) error {
	realErr := reflect.New(typ.Elem()).Elem()
	errCode := realErr.FieldByName("ErrorCode")
	if !errCode.IsValid() || !errCode.CanSet() {
		return nil
	}
	errCode.SetInt(int64(e.ErrorCode))
	errName := realErr.FieldByName("ErrorName")
	if !errName.IsValid() || !errName.CanSet() {
		return nil
	}
	errName.SetString(e.ErrorName)
	errMsg := realErr.FieldByName("Message")
	if !errMsg.IsValid() || !errMsg.CanSet() {
		return nil
	}
	errMsg.SetString(e.Message)
	i := realErr.Addr().Interface()
	if err, ok := i.(error); ok {
		return err
	}
	return nil
}

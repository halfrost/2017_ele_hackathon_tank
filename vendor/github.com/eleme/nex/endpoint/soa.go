package endpoint

import "reflect"

// ErrTypes is used to refer application defined error types.
type ErrTypes struct {
	UserErr  reflect.Type
	SysErr   reflect.Type
	UnkwnErr reflect.Type
}

// SOAMiddlewareArgs is the common used arguments for SOA related middlewares.
type SOAMiddlewareArgs struct {
	AppID             string // remote AppID for client
	ThriftServiceName string
	RemoteAddr        string
	ErrTypes          *ErrTypes
}

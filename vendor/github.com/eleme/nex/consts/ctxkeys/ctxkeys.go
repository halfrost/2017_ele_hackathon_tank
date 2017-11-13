// Package ctxkeys defines internal use key names in context.Value.
// Keys for RPCID and RequestID in github.com/eleme/thrift-tracker.
//
// Context.WithValue is not silver bullet, If you do not use it well,
// they may drive you crazy as well as global variables.
package ctxkeys

// CtxKey is a custom type for context keys.
type CtxKey string

const (
	// APIName is the key to access the func name of the thrift handler.
	APIName CtxKey = "__api_name"
	// CliAPIName is the key to access the func name of the thrift client.
	CliAPIName CtxKey = "__client_api_name"
	// OthAPIName is the key to access the others func name.
	OthAPIName CtxKey = "__other_api_name"
	// EtraceTransactioner is the key to access upstream's etrace.Transactioner.
	EtraceTransactioner CtxKey = "__etracet_transactioner"
	// AppName is the appid of current app
	AppName CtxKey = "__app_name"
	// RemoteAddr is the key to access remote address, for client.
	RemoteAddr CtxKey = "__remote_addr"
	// EtraceInfo is key to access etrace information.
	EtraceInfo CtxKey = "__etrace_info"
)

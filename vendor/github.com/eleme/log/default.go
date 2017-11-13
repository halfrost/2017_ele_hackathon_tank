package log

const (
	// TplDefault is the default log format.
	TplDefault = "{{level}} {{date}} {{time}} {{name}} {{}}"
	// TplSyslog is the log format for syslog.
	TplSyslog = "[{{app_id}} {{rpc_id}} {{request_id}}] ## {{}}"
)

var defaultLevel = INFO
var defaultLogger = New("")

var (
	// Level returns the level of default logger
	Level = defaultLogger.Level

	// SetLevel sets the level of default logger
	SetLevel = defaultLogger.SetLevel

	// Print calls Output to print to the default logger.
	// Arguments are handled in the manner of fmt.Print.
	Print = defaultLogger.Print

	// Printf calls Output to print to the default logger.
	// Arguments are handled in the manner of fmt.Printf.
	Printf = defaultLogger.Printf

	// Println calls Output to print to the default logger.
	// Arguments are handled in the manner of fmt.Println.
	Println = defaultLogger.Println

	// Fatal is equivalent to Print() followed by a call to os.Exit(1).
	Fatal = defaultLogger.Fatal

	// Fatalf is equivalent to Print() followed by a call to os.Exit(1).
	Fatalf = defaultLogger.Fatalf
)

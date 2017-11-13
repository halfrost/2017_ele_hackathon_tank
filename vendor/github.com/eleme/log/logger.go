package log

// Namer represents a named object
type Namer interface {
	Name() string
}

// Leveler represents a leveled object
type Leveler interface {
	Level() LevelType
	SetLevel(lv LevelType)
}

// NamedLeveler is the combination of Namer and Leveler
type NamedLeveler interface {
	Namer
	Leveler
}

// AsyncedLogger async log
type AsyncedLogger interface {
	SetAsync(async bool)
}

// MultiHandler represents an object with multiple logging handlers
type MultiHandler interface {
	AddHandler(h Handler)
	RemoveHandler(h Handler)
	Handlers() []Handler
}

// SimpleLogger represents a named logger which is capable of logging with
// multiple handlers and different levels.
//
// Normally this is the logger you should use.
type SimpleLogger interface {
	// Basic
	NamedLeveler
	AsyncedLogger

	// multiple handlers
	MultiHandler

	// level APIs
	Debugger
	Printer
	Infoer
	Warner
	Errorer
	Fataler

	RecordFactory() RecordFactory
	Output(record Record)
}

// RPCLogger contains a SimpleLogger with extra RPC APIs
type RPCLogger interface {
	SimpleLogger
	// RPC APIs
	WithRPCID(string) RPCLogger
	WithRequestID(string) RPCLogger
}

// Debugger represents a logger with Debug APIs
type Debugger interface {
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
}

// Printer represents a logger with Print APIs
type Printer interface {
	Print(a ...interface{})
	Println(a ...interface{})
	Printf(f string, a ...interface{})
}

// Infoer represents a logger with Info APIs
type Infoer interface {
	Info(a ...interface{})
	Infof(f string, a ...interface{})
}

// Warner represents a logger with Warn APIs
type Warner interface {
	Warn(a ...interface{})
	Warnf(f string, a ...interface{})
}

// Errorer represents a logger with Error APIs
type Errorer interface {
	Error(a ...interface{})
	Errorf(f string, a ...interface{})
}

// Fataler represents a logger with Fatal APIs
type Fataler interface {
	Fatal(a ...interface{})
	Fatalf(f string, a ...interface{})
}

package log

import (
	"context"

	"github.com/eleme/log"
	tracker "github.com/eleme/thrift-tracker"
)

// RPCContextLogger contains a RPCLogger with extra context APIs.
type RPCContextLogger interface {
	log.RPCLogger
	ContextDebug(ctx context.Context, a ...interface{})
	ContextDebugf(ctx context.Context, format string, a ...interface{})
	ContextPrint(ctx context.Context, a ...interface{})
	ContextPrintf(ctx context.Context, format string, a ...interface{})
	ContextPrintln(ctx context.Context, a ...interface{})
	ContextInfo(ctx context.Context, a ...interface{})
	ContextInfof(ctx context.Context, format string, a ...interface{})
	ContextWarn(ctx context.Context, a ...interface{})
	ContextWarnf(ctx context.Context, format string, a ...interface{})
	ContextError(ctx context.Context, a ...interface{})
	ContextErrorf(ctx context.Context, format string, a ...interface{})
	ContextFatal(ctx context.Context, a ...interface{})
	ContextFatalf(ctx context.Context, format string, a ...interface{})
}

// EContextLogger implements RPCContextLogger.
type EContextLogger struct {
	log.RPCLogger
}

// NewEContextLogger creates a new EContextLogger.
func NewEContextLogger(logger log.RPCLogger) *EContextLogger {
	return &EContextLogger{
		RPCLogger: logger,
	}
}

// ContextDebug outputs log with DEBUG level.
func (l *EContextLogger) ContextDebug(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Debug(a...)
}

// ContextDebugf outputs log with DEBUG level and given format.
func (l *EContextLogger) ContextDebugf(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Debugf(format, a...)
}

// ContextPrint outputs log with default level.
func (l *EContextLogger) ContextPrint(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Print(a...)
}

// ContextPrintf outputs log with default level and given format.
func (l *EContextLogger) ContextPrintf(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Printf(format, a...)
}

// ContextPrintln outputs log with DEBUG level.
func (l *EContextLogger) ContextPrintln(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Println(a...)
}

// ContextInfo outputs log with INFO level.
func (l *EContextLogger) ContextInfo(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Info(a...)
}

// ContextInfof outputs log with INFO level and given format.
func (l *EContextLogger) ContextInfof(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Infof(format, a...)
}

// ContextWarn outputs log with WARN level.
func (l *EContextLogger) ContextWarn(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Warn(a...)
}

// ContextWarnf outputs log with WARN level and given format.
func (l *EContextLogger) ContextWarnf(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Warnf(format, a...)
}

// ContextError outputs log with ERRO level.
func (l *EContextLogger) ContextError(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Error(a...)
}

// ContextErrorf outputs log with ERRO level and given format.
func (l *EContextLogger) ContextErrorf(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Errorf(format, a...)
}

// ContextFatal outputs log with FATA level, then exit program immediately.
func (l *EContextLogger) ContextFatal(ctx context.Context, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Fatal(a...)
}

// ContextFatalf outputs log with FATA level and given format, then exit program immediately.
func (l *EContextLogger) ContextFatalf(ctx context.Context, format string, a ...interface{}) {
	logger := LoggerWithMeta(ctx, l.RPCLogger)
	logger.Fatalf(format, a...)
}

// LoggerWithMeta attaches context info to logger, and return a new logger.
func LoggerWithMeta(ctx context.Context, logger log.RPCLogger) log.RPCLogger {
	if rpcID, ok := ctx.Value(tracker.CtxKeySequenceID).(string); ok {
		logger = logger.WithRPCID(rpcID)
	} else {
		logger = logger.WithRPCID("-")
	}
	if reqID, ok := ctx.Value(tracker.CtxKeyRequestID).(string); ok {
		logger = logger.WithRequestID(reqID)
	} else {
		logger = logger.WithRequestID("-")
	}
	return logger
}

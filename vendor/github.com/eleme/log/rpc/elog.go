package rpc

import (
	"fmt"
	"os"

	"github.com/eleme/log"
)

// ELogger implements the RPCLogger.
type ELogger struct {
	*log.Logger
	rpcID         string
	requestID     string
	recordFactory log.RecordFactory
}

// NewELogger creates a ELogger with given name as a RPCLogger.
func NewELogger(name string) log.RPCLogger {
	l := log.NewWithWriter(name, os.Stdout)
	elog := &ELogger{Logger: l}
	elog.recordFactory = NewELogRecordFactory(elog.rpcID, elog.requestID)
	return elog
}

// RPCID returns the RPCID of elog.
func (e *ELogger) RPCID() string {
	return e.rpcID
}

// RequestID returns the requestID of elog.
func (e *ELogger) RequestID() string {
	return e.requestID
}

// WithRPCID set rpcID on logger.
func (e *ELogger) WithRPCID(rpcID string) log.RPCLogger {
	elog := &ELogger{
		Logger:    e.Logger,
		rpcID:     rpcID,
		requestID: e.requestID,
	}
	elog.recordFactory = NewELogRecordFactory(elog.rpcID, elog.requestID)
	return elog
}

// WithRequestID set requestID on logger.
func (e *ELogger) WithRequestID(requestID string) log.RPCLogger {
	elog := &ELogger{
		Logger:    e.Logger,
		rpcID:     e.rpcID,
		requestID: requestID,
	}
	elog.recordFactory = NewELogRecordFactory(elog.rpcID, elog.requestID)
	return elog
}

// Debug APIs

// Debug calls Output to log with DEBUG level
func (e *ELogger) Debug(a ...interface{}) {
	if log.DEBUG < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.DEBUG, fmt.Sprint(a...)))
}

// Debugf calls Output to log with DEBUG level and given format
func (e *ELogger) Debugf(format string, a ...interface{}) {
	if log.DEBUG < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.DEBUG, fmt.Sprintf(format, a...)))
}

// Print APIs

// Print calls Output to log with default level
func (e *ELogger) Print(a ...interface{}) {
	e.Output(e.recordFactory(e.Name(), 2, e.Level(), fmt.Sprint(a...)))
}

// Println calls Output to log with default level
func (e *ELogger) Println(a ...interface{}) {
	e.Output(e.recordFactory(e.Name(), 2, e.Level(), fmt.Sprint(a...)))
}

// Printf calls Output to log with default level and given format
func (e *ELogger) Printf(f string, a ...interface{}) {
	e.Output(e.recordFactory(e.Name(), 2, e.Level(), fmt.Sprintf(f, a...)))
}

// Info APIs

// Info calls Output to log with INFO level
func (e *ELogger) Info(a ...interface{}) {
	if log.INFO < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.INFO, fmt.Sprint(a...)))
}

// Infof calls Output to log with INFO level and given format
func (e *ELogger) Infof(f string, a ...interface{}) {
	if log.INFO < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.INFO, fmt.Sprintf(f, a...)))
}

// Warn APIs

// Warn calls Output to log with WARN level
func (e *ELogger) Warn(a ...interface{}) {
	if log.WARN < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.WARN, fmt.Sprint(a...)))
}

// Warnf calls Output to log with WARN level and given format
func (e *ELogger) Warnf(f string, a ...interface{}) {
	if log.WARN < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.WARN, fmt.Sprintf(f, a...)))
}

// Error APIs

// Error calls Output to log with ERRO level
func (e *ELogger) Error(a ...interface{}) {
	if log.ERRO < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.ERRO, fmt.Sprint(a...)))
}

// Errorf calls Output to log with ERRO level and given format
func (e *ELogger) Errorf(f string, a ...interface{}) {
	if log.ERRO < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.ERRO, fmt.Sprintf(f, a...)))
}

// Fatal APIs

// Fatal calls Output to log with FATA level followed by a call to os.Exit(1)
func (e *ELogger) Fatal(a ...interface{}) {
	if log.FATA < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.FATA, fmt.Sprint(a...)))
	os.Exit(1)
}

// Fatalf calls Output to log with FATA level with given format, followed by a call to os.Exit(1)
func (e *ELogger) Fatalf(f string, a ...interface{}) {
	if log.FATA < e.Level() {
		return
	}
	e.Output(e.recordFactory(e.Name(), 2, log.FATA, fmt.Sprintf(f, a...)))
	os.Exit(1)
}

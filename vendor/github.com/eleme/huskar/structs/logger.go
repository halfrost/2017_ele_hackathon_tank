// Copyright 2016 Eleme Inc. All rights reserved.

package structs

import "log"

// Logger is an interface that can be implemented to provide custom log output.
type Logger interface {
	Printf(format string, v ...interface{})
}

// DefaultLogger uses the stdlib log package for logging.
type DefaultLogger struct{}

// Printf implementes Logger interface.
func (l DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

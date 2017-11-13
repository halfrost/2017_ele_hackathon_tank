package syslog

import (
	"io"
	"log/syslog"

	"github.com/eleme/log"
)

// Handler can send log to syslog
type Handler struct {
	log.Formatter
	w *syslog.Writer
}

// NewHandler creates a Handler with given syslog.Writer which
// could be created by syslog.New, the log format as follows:
//
//	"[{{app_id}} {{rpc_id}} {{request_id}}] ## {{}}"
func NewHandler(w *syslog.Writer) (*Handler, error) {
	f := log.NewBaseFormatter(false)
	if err := f.ParseFormat(log.TplSyslog); err != nil {
		return nil, err
	}
	return NewHandlerWithFormat(w, f), nil
}

// NewHandlerWithFormat is just like NewHandler but with customized
// format string
func NewHandlerWithFormat(w *syslog.Writer, f log.Formatter) *Handler {
	h := new(Handler)
	h.w = w
	h.Formatter = f
	return h
}

// Log prints the Record info syslog writer
func (sh *Handler) Log(r log.Record) {
	b := string(sh.Formatter.Format(r))
	switch r.Level() {
	case log.DEBUG:
		sh.w.Debug(b)
	case log.INFO:
		sh.w.Info(b)
	case log.WARN:
		sh.w.Warning(b)
	case log.ERRO:
		sh.w.Err(b)
	case log.FATA:
		sh.w.Crit(b)
	}
}

// Writer return the writer
func (sh *Handler) Writer() io.Writer {
	return sh.w
}

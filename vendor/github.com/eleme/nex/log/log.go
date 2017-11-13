package log

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"sync"

	"github.com/eleme/log"
	"github.com/eleme/log/rpc"
	esyslog "github.com/eleme/log/syslog"
)

const (
	// NexLogSimpleTpl is a simple version log format template.
	NexLogSimpleTpl = "{{datetime}} {{level}} {{name}} ## {{}}"
	// NexLogTpl defines standard log format template.
	NexLogTpl = "{{datetime}} {{level}} {{name}}[{{pid}}]: [{{app_id}} {{rpc_id}} {{request_id}}] ## {{}}"
	// NexSyslogTpl is the log format template for syslog.
	NexSyslogTpl = "[{{app_id}} {{rpc_id}} {{request_id}}] ## {{}}"
)

type eloggerHub struct {
	loggers    map[string]RPCContextLogger
	appName    string
	colored    bool
	syslogAddr string
	mu         sync.RWMutex
}

var loggerHub = &eloggerHub{
	loggers:    make(map[string]RPCContextLogger),
	appName:    "",
	colored:    true,
	syslogAddr: ":514",
	mu:         sync.RWMutex{},
}

// Setup set up app id and syslog address.
func Setup(appID string, colored bool, syslogAddr string) {
	loggerHub.mu.Lock()
	log.SetGlobalAppID(appID)
	loggerHub.appName = appID
	loggerHub.colored = colored
	loggerHub.syslogAddr = syslogAddr
	loggerHub.mu.Unlock()
}

// GetLogger returns a RPCLogger by name.
func GetLogger(name string) (log.RPCLogger, error) {
	return loggerHub.getLogger(name)
}

// GetContextLogger returns a RPCContextLogger by name.
func GetContextLogger(name string) (RPCContextLogger, error) {
	return loggerHub.getLogger(name)
}

func (h *eloggerHub) getLogger(name string) (RPCContextLogger, error) {
	h.mu.RLock()
	name = h.addNamePrefix(name)
	logger, exists := h.loggers[name]
	h.mu.RUnlock()
	if exists {
		return logger, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if createdLogger, exists := h.loggers[name]; exists { // double check
		return createdLogger, nil
	}

	newLogger, err := h.makeLogger(name)
	if err != nil {
		return nil, err
	}
	ctxLogger := NewEContextLogger(newLogger)
	h.loggers[name] = ctxLogger
	return ctxLogger, nil
}

func (h *eloggerHub) addNamePrefix(name string) string {
	if !strings.HasPrefix(name, h.appName) {
		return fmt.Sprintf("%s.%s", h.appName, name)
	}
	return name
}

func (h *eloggerHub) makeLogger(name string) (log.RPCLogger, error) {
	elogger := rpc.NewELogger(name)
	for _, hdr := range elogger.Handlers() {
		elogger.RemoveHandler(hdr)
	}

	streamHandler, err := h.newRPCStreamHandler()
	if err != nil {
		return nil, err
	}
	elogger.AddHandler(streamHandler)

	// add syslog handler
	if h.syslogAddr != "" {
		syslogHandler, err := h.newRPCSyslogHandler(name)
		if err != nil {
			return nil, err
		}
		elogger.AddHandler(syslogHandler)
	}
	return elogger, nil
}

func (h *eloggerHub) newRPCStreamHandler() (log.Handler, error) {
	f := rpc.NewELogFormatter(h.colored)
	if err := f.ParseFormat(NexLogTpl); err != nil {
		return nil, err
	}
	return log.NewStreamHandler(os.Stdout, f), nil
}

func (h *eloggerHub) newRPCSyslogHandler(tag string) (log.Handler, error) {
	sw, err := syslog.Dial("udp", h.syslogAddr, syslog.LOG_DEBUG|syslog.LOG_LOCAL6, tag)
	if err != nil {
		return nil, err
	}
	f := rpc.NewELogFormatter(false)
	if err = f.ParseFormat(NexSyslogTpl); err != nil {
		return nil, err
	}
	return esyslog.NewHandlerWithFormat(sw, f), nil
}

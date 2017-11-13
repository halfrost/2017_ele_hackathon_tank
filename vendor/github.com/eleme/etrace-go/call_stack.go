package etrace

import (
	"sync"
	"time"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// CallStack is the transactions container.
type CallStack struct {
	manager     *Trace
	root        *Transaction
	cfg         *config.Config
	clientAppID string
	id          int64
	requestID   string
	rpcID       string
	typ         string
	name        string
	lock        *sync.Mutex
	isCompleted bool
}

func newCallStack(t *Trace, cfg *config.Config, requestID, rpcID, typ, name string) *CallStack {
	if requestID == "" {
		requestID = CreateRequestID(cfg.AppID)
	}
	rpcID, clientAppID := CreateRPCID(cfg.AppID, rpcID)
	s := &CallStack{
		manager:     t,
		cfg:         cfg,
		clientAppID: clientAppID,
		requestID:   requestID,
		rpcID:       rpcID,
		typ:         typ,
		name:        name,
		isCompleted: false,
		lock:        &sync.Mutex{},
	}
	s.root = newTransaction(s, nil, &s.id, typ, name)
	return s
}

// GetClientAppID get current client app id .
func (s *CallStack) GetClientAppID() string {
	return s.clientAppID
}

// GetRequestID get current client request id .
func (s *CallStack) GetRequestID() string {
	return s.requestID
}

// GetUpstreamRPCID returns the upstream rpc id.
func (s *CallStack) GetUpstreamRPCID() string {
	return s.root.GetUpstreamRPCID()
}

// GetCurrentRPCID returns current rpc id.
func (s *CallStack) GetCurrentRPCID() string {
	return s.root.GetCurrentRPCID()
}

// GetCurrentRPCIDWithAppID returns current rpc id.
func (s *CallStack) GetCurrentRPCIDWithAppID() string {
	return s.root.GetCurrentRPCIDWithAppID()
}

// CreateEvent add an event in the root transaction.
func (s *CallStack) CreateEvent(typ, name string) *Event {
	return s.root.CreateEvent(typ, name)
}

// Fork create a child transaction in root transaction.
func (s *CallStack) Fork(typ, name string) Transactioner {
	return s.root.Fork(typ, name)
}

// AddTag add a tag in root transaction.
func (s *CallStack) AddTag(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isCompleted {
		return
	}
	s.root.AddTag(key, val)
}

// SetStatus set root transaction status,default status is "0".
func (s *CallStack) SetStatus(status string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isCompleted {
		return
	}
	s.root.SetStatus(status)
}

// LogError tries to catch exception.
func (s *CallStack) LogError(typ string, msg string) {
	s.root.LogError(typ, msg)
}

// Commit commit current callStack.
func (s *CallStack) Commit() {
	s.lock.Lock()
	if s.isCompleted {
		s.lock.Unlock()
		return
	}
	s.isCompleted = true
	s.root.Commit()
	s.lock.Unlock()
	if s.manager != nil {
		s.manager.commitCallStack(s)
	}
}

// CommitWithDuration commit current transaction with given duration.
func (s *CallStack) CommitWithDuration(duration time.Duration) {
	s.lock.Lock()
	if s.isCompleted {
		s.lock.Unlock()
		return
	}
	s.isCompleted = true
	s.root.CommitWithDuration(duration)
	s.lock.Unlock()
	if s.manager != nil {
		s.manager.commitCallStack(s)
	}
}

// MarshalJSON returns bytes based on etrace protocol.
func (s *CallStack) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	s.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON appends bytes to the writer.
func (s *CallStack) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawByte('[')
	w.String(s.cfg.AppID)
	w.RawByte(',')
	w.String(s.cfg.HostIP)
	w.RawByte(',')
	w.String(s.cfg.HostName)
	w.RawByte(',')
	w.String(s.requestID)
	w.RawByte(',')
	w.String(s.rpcID)
	w.RawByte(',')
	s.root.MarshalEasyJSON(w)
	w.RawByte(',')
	w.String(s.cfg.Cluster)
	w.RawByte(',')
	w.String(s.cfg.EZone)
	w.RawByte(',')
	w.String(s.cfg.IDC)
	w.RawByte(']')
}

//Package etrace is a golang etrace client.
package etrace

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/eleme/etrace-go/client"
	"github.com/eleme/etrace-go/config"
	"github.com/eleme/etrace-go/metric"
	"github.com/mailru/easyjson/jwriter"
)

// Transactioner trace rpc call stacks.
type Transactioner interface {
	Fork(typ, name string) Transactioner
	AddTag(key, val string)
	SetStatus(status string)
	Commit()
	CommitWithDuration(duration time.Duration)
	CreateEvent(typ, name string) *Event
	GetClientAppID() string
	GetRequestID() string
	GetUpstreamRPCID() string
	GetCurrentRPCID() string
	GetCurrentRPCIDWithAppID() string
	LogError(typ string, msg string)
}

// Trace watches configurations from etrace server and flush transaction messages periodically.
type Trace struct {
	sync.Mutex
	cfg          *config.Config
	thriftclient client.Client
	mc           *metric.Collector
	callStackCh  chan *CallStack
}

// New create Trace instance. If AppID  or EtraceConfigURL is empty, New will return error.
func New(cfg Config) (*Trace, error) {
	cfg2 := withDefaultConfig(cfg)
	thriftClient := client.New(&cfg2)
	mc := metric.NewCollector(&cfg2, thriftClient)
	t := &Trace{
		cfg:          &cfg2,
		mc:           mc,
		thriftclient: thriftClient,
		callStackCh:  make(chan *CallStack, 512),
	}
	go t.transactionWork()
	return t, nil
}

// NewTransaction starts a new transaction.
func (t *Trace) NewTransaction(requestID, rpcID, typ, name string) Transactioner {
	return newCallStack(t, t.cfg, requestID, rpcID, typ, name)
}

func (t *Trace) commitCallStack(s *CallStack) {
	select {
	case t.callStackCh <- s:
	}
}

func (t *Trace) transactionWork() {
	var stacks []*CallStack
	ticker := time.NewTicker(t.cfg.EtraceMaxCacheTime)
	for {
		select {
		case <-ticker.C:
			t.flushTransaction(stacks)
			stacks = nil
		case trans := <-t.callStackCh:
			remoteCfg := t.cfg.Remoter.RemoteConfig()
			if !remoteCfg.Enabled {
				t.cfg.Logger.Printf("remote config is not enabled")
				continue
			}
			stacks = append(stacks, trans)
			if len(stacks) >= remoteCfg.MessageCount {
				t.flushTransaction(stacks)
				stacks = nil
			}
		}
	}
}

func (t *Trace) flushTransaction(stacks []*CallStack) {
	if len(stacks) == 0 {
		return
	}
	header, message, err := t.buildMessage(stacks)
	if err != nil {
		t.cfg.Logger.Printf("%v", err)
	}
	t.thriftclient.Send(header, message)
}

func (t *Trace) buildMessage(stacks []*CallStack) (header, message []byte, err error) {
	now := time.Now()
	headerMap := map[string]interface{}{
		"appId":    t.cfg.AppID,
		"hostIp":   t.cfg.HostIP,
		"hostName": t.cfg.HostName,
		"ast":      now.UnixNano() / 1e6,
	}
	header, err = json.Marshal(headerMap)
	if err != nil {
		return
	}
	w := jwriter.Writer{}
	w.RawByte('[')
	childFirst := true
	for _, s := range stacks {
		if !childFirst {
			w.RawByte(',')
		}
		childFirst = false
		s.MarshalEasyJSON(&w)
	}
	w.RawByte(']')
	message, err = w.BuildBytes()
	return
}

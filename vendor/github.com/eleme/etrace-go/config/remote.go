package config

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// Remoter is remote configuration container.
type Remoter interface {
	RemoteConfig() *RemoteConfig
	Collectors() []Collector
}

// Remote is for remote configuration.
type Remote struct {
	interval   time.Duration
	remoteCfg  unsafe.Pointer
	collectors unsafe.Pointer
	client     *configAgent
}

// NewRemoter creates a new remote configuration client.
func NewRemoter(endpoint, appid, hostip string, interval, timeout time.Duration) Remoter {
	remoteCfg := defaultRemoteConfig()
	cols := []Collector{}
	c := &Remote{
		interval:   interval,
		client:     newConfigAgent(endpoint, appid, hostip, timeout),
		remoteCfg:  unsafe.Pointer(&remoteCfg),
		collectors: unsafe.Pointer(&cols),
	}
	c.fetch()
	go c.work()
	return c
}

// RemoteConfig returns remote configuration data.
func (cfg *Remote) RemoteConfig() *RemoteConfig {
	ptr := atomic.LoadPointer(&cfg.remoteCfg)
	return (*RemoteConfig)(ptr)
}

// Collectors returns remote server list.
func (cfg *Remote) Collectors() []Collector {
	ptr := atomic.LoadPointer(&cfg.collectors)
	return *(*[]Collector)(ptr)
}

func (cfg *Remote) work() {
	ticker := time.NewTicker(cfg.interval)
	for {
		select {
		case <-ticker.C:
			cfg.fetch()
		}
	}
}

func (cfg *Remote) fetch() {
	config, err := cfg.client.fetchConfig()
	if err == nil {
		atomic.StorePointer(&cfg.remoteCfg, unsafe.Pointer(&config))
	}
	collectors, err := cfg.client.fetchServer()
	if err == nil {
		atomic.StorePointer(&cfg.collectors, unsafe.Pointer(&collectors))
	}
}

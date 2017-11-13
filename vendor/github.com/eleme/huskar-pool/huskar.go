package huskarPool

import (
	"errors"
	"fmt"
	"sync"
	"time"

	huskarService "github.com/eleme/huskar/service"
	"github.com/eleme/huskar/structs"
)

var (
	// ErrNoHosts no host available to create connection.
	ErrNoHosts = errors.New("no hosts available")
)

// Option for huskar initialization.
type Option struct {
	Endpoint    string        // Endpoint is a Huskar API URL prefix.
	Token       string        // Token is the authorized token string.
	Service     string        // Service is the name of your service(e.g. zeus.eos).
	Cluster     string        // Cluster is the cluster name of your servers.
	WaitTimeout time.Duration // WaitTimeout is the timeout of first connect to Huskar API. Default: 5 seconds.
	DialTimeout time.Duration // DialTimeout is the timeout of dialing to Huskar API. Default: 1 second.
	RetryDelay  time.Duration // RetryDelay is retry delay to connect to Huskar API. Default: 1 second.
	SOAMode     string        // SOAMode is the 'soa_mode' in /etc/eleme/env.yaml.
}

// Huskar services.
type Huskar struct {
	sync.RWMutex
	huskar   huskarService.Registrator
	services map[string]*service
}

// New create a huskar instance.
func New(op Option) (*Huskar, error) {
	huskar, err := huskarService.New(structs.Config{
		Endpoint:    op.Endpoint,
		Token:       op.Token,
		Service:     op.Service,
		Cluster:     op.Cluster,
		WaitTimeout: op.WaitTimeout,
		DialTimeout: op.DialTimeout,
		RetryDelay:  op.RetryDelay,
		SOAMode:     op.SOAMode,
	})
	if err != nil {
		return nil, err
	}
	h := NewWithService(huskar)
	return h, nil
}

// NewWithService init huskar instance with existed registrator.
func NewWithService(reg huskarService.Registrator) *Huskar {
	h := &Huskar{
		huskar:   reg,
		services: make(map[string]*service),
	}
	return h
}

func (h *Huskar) serviceKey(appid, cluster string) string {
	return fmt.Sprintf(`{"appid":"%s","cluster":"%s"}`, appid, cluster)
}

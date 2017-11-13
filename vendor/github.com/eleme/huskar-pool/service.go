package huskarPool

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	huskarService "github.com/eleme/huskar/service"
	"github.com/eleme/huskar/structs"
)

// Service is a container for one pair with appid and cluster.
type service struct {
	once  sync.Once
	waitC chan struct{}
	sync.RWMutex
	huskar  huskarService.Registrator
	appid   string
	cluster string
	hosts   map[string]*HostInstance
}

func newService(huskar huskarService.Registrator, appid string, cluster string) (*service, error) {
	s := &service{
		huskar:  huskar,
		appid:   appid,
		cluster: cluster,
		hosts:   make(map[string]*HostInstance),
		waitC:   make(chan struct{}),
	}
	evtC, _, err := huskar.Watch(appid, cluster)
	if err != nil {
		return nil, err
	}
	go s.watch(evtC)
	return s, nil
}

func (s *service) watch(evtC <-chan *huskarService.Event) {
	for {
		evt := <-evtC
		if evt.Err == nil {
			s.processEvent(evt)
		} else {
			log.Printf("service error event:%s\n", evt.Err)
		}
		s.once.Do(func() { close(s.waitC) })
	}
}

func (s *service) processEvent(evt *huskarService.Event) {
	s.Lock()
	defer s.Unlock()
	if evt.Type == structs.EventAll {
		hosts := make(map[string]*HostInstance)
		for _, host := range evt.Instances {
			if host.Value.State == huskarService.StateDown {
				continue
			}
			if h, ok := s.hosts[host.Key]; ok {
				hosts[host.Key] = h
			} else {
				hosts[host.Key] = newHostInstance(host)
			}
		}
		s.hosts = hosts
	} else if evt.Type == structs.EventUpdate {
		for _, host := range evt.Instances {
			if host.Value.State == huskarService.StateDown {
				delete(s.hosts, host.Key)
				continue
			}
			s.hosts[host.Key] = newHostInstance(host)
		}
	} else if evt.Type == structs.EventDelete {
		for _, host := range evt.Instances {
			delete(s.hosts, host.Key)
		}
	}
}

func (s *service) wait(ctx context.Context) error {
	select {
	case <-s.waitC:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *service) getHostList() []*hostWrapper {
	var hosts []*hostWrapper
	s.RLock()
	for _, h := range s.hosts {
		hosts = append(hosts, h.wrapper())
	}
	s.RUnlock()
	sort.Sort(byConn(hosts))
	return hosts
}

// HostInstance track connection numbers on one host.
type HostInstance struct {
	sync.RWMutex
	meta Meta
	addr string
	conn int
}

func newHostInstance(ins *huskarService.Instance) *HostInstance {
	meta := Meta{
		IP:   ins.Value.IP,
		Port: ins.Value.Port["main"],
	}
	if port, in := ins.Value.Port["back"]; in {
		meta.BackPort = port
	}
	h := &HostInstance{
		addr: fmt.Sprintf("%s:%d", meta.IP, meta.Port),
		meta: meta,
	}
	h.meta.Ins = h
	return h
}

// GetConnNum get the connection number.
func (h *HostInstance) GetConnNum() int {
	h.RLock()
	defer h.RUnlock()
	return h.conn
}

// Add adds a delta to connection counter.
func (h *HostInstance) Add(delta int) {
	h.Lock()
	defer h.Unlock()
	h.conn = h.conn + delta
}

func (h *HostInstance) wrapper() *hostWrapper {
	h.RLock()
	defer h.RUnlock()
	return &hostWrapper{
		conn: h.conn,
		ins:  h,
	}
}

type hostWrapper struct {
	conn int
	ins  *HostInstance
}

type byConn []*hostWrapper

func (b byConn) Len() int           { return len(b) }
func (b byConn) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byConn) Less(i, j int) bool { return b[i].conn < b[j].conn }

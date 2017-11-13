package huskarPool

import (
	"context"
	"net"
	"time"

	"github.com/eleme/huskar-pool/pool"
)

const (
	// MaxTries is the default max tries for get resource from pool.
	MaxTries = 5
)

// Factory use addr to create a new resource.
type Factory func(meta Meta, conn net.Conn) (pool.Resource, error)

// Pinger call ping to test service if is available.
type Pinger interface {
	Ping() error
}

// PoolStatis is the pool callback to watch the pool internal state.The callback will be called every second.
type PoolStatis func(appid, cluster string, workingClients, idleClients, capacity, waitCount int64)

// Meta is the meta info for factory.
type Meta struct {
	IP       string
	Port     int
	BackPort int
	Ins      *HostInstance
}

// PoolOption custom pool
type PoolOption struct {
	Capacity    int
	MaxCap      int
	IdleTimeout time.Duration
	MaxTries    int
	DialTimeout time.Duration
	Statis      PoolStatis
}

// Pool resource .
type Pool struct {
	*pool.ResourcePool
	appid   string
	cluster string
	srv     *service
	factory Factory
	option  PoolOption
}

// NewResourcePool creates a pool for appid and cluster.
func (h *Huskar) NewResourcePool(appid, cluster string, factory Factory, option PoolOption) (_ *Pool, err error) {
	if option.MaxTries == 0 {
		option.MaxTries = MaxTries
	}
	key := h.serviceKey(appid, cluster)
	h.Lock()
	srv, ok := h.services[key]
	if !ok {
		srv, err = newService(h.huskar, appid, cluster)
		if err != nil {
			h.Unlock()
			return nil, err
		}
		h.services[key] = srv
	}
	h.Unlock()
	pl := pool.NewResourcePool(h.factoryFn(srv, factory, option), option.Capacity, option.MaxCap, option.IdleTimeout)
	return newPool(pl, appid, cluster, srv, option), nil
}

func (h *Huskar) factoryFn(srv *service, factory Factory, op PoolOption) pool.Factory {
	return func() (pool.Resource, error) {
		hosts := srv.getHostList()
		for _, h := range hosts {
			ins := h.ins
			conn, err := net.DialTimeout("tcp", ins.addr, op.DialTimeout)
			if err != nil {
				continue
			}
			conn = HookClose(conn, func() { ins.Add(-1) })
			ins.Add(1)
			if factory == nil {
				return conn, nil
			} else {
				re, err := factory(ins.meta, conn)
				if err != nil {
					conn.Close()
					continue
				}
				return re, nil
			}
		}
		return nil, ErrNoHosts
	}
}

func newPool(pl *pool.ResourcePool, appid, cluster string, srv *service, op PoolOption) *Pool {
	npl := &Pool{
		ResourcePool: pl,
		appid:        appid,
		cluster:      cluster,
		srv:          srv,
		option:       op,
	}
	go npl.watcher()
	return npl
}

// Wait for huskar first event.
func (pl *Pool) Wait(ctx context.Context) error {
	return pl.srv.wait(ctx)
}

// Get get a resource from pool.
func (pl *Pool) Get(ctx context.Context) (pool.Resource, error) {
	err := pl.Wait(ctx)
	if err != nil {
		return nil, err
	}
	var cc pool.Resource
	for i := 0; i < pl.option.MaxTries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		cc, err = pl.ResourcePool.Get(ctx)
		if err != nil {
			continue
		}
		pg, ok := cc.(Pinger)
		if !ok {
			return cc, nil
		}
		if err = pg.Ping(); err == nil {
			return cc, nil
		} else {
			cc.Close()
			pl.ResourcePool.Put(nil)
		}
	}
	return nil, err
}

func (pl *Pool) watcher() {
	if pl.option.Statis != nil {
		pl.watchStats()
	}
}

func (pl *Pool) watchStats() {
	var lastWaitCount int64
	for {
		<-time.After(time.Second)
		capacity, available, _, waitCount, _, _ := pl.Stats()
		pl.option.Statis(pl.appid, pl.cluster, capacity-available, available, capacity, waitCount-lastWaitCount)
		lastWaitCount = waitCount
	}
}

// Put puts the resource into pool,if the resource is broken,you should call close for the resource and put nil into pool.
func (pl *Pool) Put(re pool.Resource) {
	pl.ResourcePool.Put(re)
}

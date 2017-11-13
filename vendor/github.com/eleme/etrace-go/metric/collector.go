package metric

import (
	"time"

	"github.com/eleme/etrace-go/client"
	"github.com/eleme/etrace-go/config"
)

// Collector is metric collecting.
type Collector struct {
	cfg          *config.Config
	thriftClient client.Client
	metricCh     chan Metric
}

// NewCollector initialize a new collector.
func NewCollector(cfg *config.Config, thriftClient client.Client) *Collector {
	c := &Collector{
		cfg:          cfg,
		thriftClient: thriftClient,
		metricCh:     make(chan Metric, 1024),
	}
	go c.work()
	return c
}

// NewCounter returns a new counter.
func (c *Collector) NewCounter(name string) *Counter {
	return newCounter(c.cfg, c, name)
}

// NewTimer returns a new timer.
func (c *Collector) NewTimer(name string) *Timer {
	return newTimer(c.cfg, c, name)
}

// NewGauge returns a new gauge.
func (c *Collector) NewGauge(name string) *Gauge {
	return newGauge(c.cfg, c, name)
}

// NewPayload returns a new payload.
func (c *Collector) NewPayload(name string) *Payload {
	return newPayload(c.cfg, c, name)
}

// Commit add metric for flush.
func (c *Collector) Commit(m Metric) {
	select {
	case c.metricCh <- m:
	}
}

func (c *Collector) work() {
	mf := newBuffer(c.cfg)
	ticker := time.NewTicker(c.cfg.EtraceMaxCacheTime)
	for {
		select {
		case <-ticker.C:
			c.flush(mf)
			mf.reset()
		case m := <-c.metricCh:
			remoteCfg := c.cfg.Remoter.RemoteConfig()
			if !remoteCfg.Enabled {
				c.cfg.Logger.Printf("remote config is not enabled")
				continue
			}
			mf.addMetric(m)
		}
	}
}

func (c *Collector) flush(mf *buffer) {
	if mf.count() == 0 {
		return
	}
	for _, p := range mf.packages() {
		header, message, err := p.BuildMessage()
		if err != nil {
			c.cfg.Logger.Printf("%v", err)
		}
		c.thriftClient.Send(header, message)
	}
}

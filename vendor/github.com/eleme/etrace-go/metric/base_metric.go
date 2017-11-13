package metric

import (
	"sync/atomic"
	"time"

	"github.com/eleme/etrace-go/config"
)

type baseMetric struct {
	cfg       *config.Config
	cm        Commiter
	typ       Type
	completed int32
	name      string
	timestamp time.Time
	tag       *tag
}

func newBaseMetric(cfg *config.Config, cm Commiter, typ Type, name string) baseMetric {
	return baseMetric{
		cfg:       cfg,
		cm:        cm,
		typ:       typ,
		name:      name,
		timestamp: time.Now(),
		tag:       newTag(),
	}
}

// Hash generate metric hash for counter.
func (b *baseMetric) Hash() Hash {
	timeKey := GetTimespanKey(b.timestamp, b.cfg.EtraceMaxCacheTime)
	return Hash{
		Type:    b.typ,
		Time:    timeKey,
		TagHash: b.tag.HashCode(),
	}
}

// SetType change type of metric.
func (b *baseMetric) setType(typ Type) {
	b.typ = typ
}

// Type returns const TypeCounter.
func (b *baseMetric) Type() Type {
	return b.typ
}

// Name returns counter name.
func (b *baseMetric) Name() string {
	return b.name
}

// TimeStamp returns counter creation time.
func (b *baseMetric) Timestamp() time.Time {
	return b.timestamp
}

// AddTag adds a new tag for base metric.
func (b *baseMetric) AddTag(key, val string) {
	if atomic.LoadInt32(&b.completed) == 0 {
		b.tag.Add(key, val)
	}
}

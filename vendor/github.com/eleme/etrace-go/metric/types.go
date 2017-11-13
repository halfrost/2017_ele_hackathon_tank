package metric

import (
	"time"

	"github.com/mailru/easyjson/jwriter"
)

// Type for metric classification.
type Type byte

const (
	// TypeCounter is counter type,
	TypeCounter Type = iota
	// TypeGauge is gauge type.
	TypeGauge
	// TypePayload is payload type.
	TypePayload
	// TypeTimer is timer type.
	TypeTimer
	// TypeHistogram is histogram type.
	TypeHistogram
	// TypeTotalCount is type number guard.
	TypeTotalCount
)

// Hash for metric container hash metric.
type Hash struct {
	Type    Type
	Time    int64
	TagHash uint64
}

// Metric show metric basic functions.
type Metric interface {
	Hash() Hash
	Type() Type
	Name() string
	Timestamp() time.Time
	Merge(Metric)
	MarshalEasyJSON(w *jwriter.Writer)
	MarshalJSON() ([]byte, error)
}

// Commiter for metric container.
type Commiter interface {
	Commit(m Metric)
}

type dummyCommiter struct {
	vals []Metric
}

func (d *dummyCommiter) Commit(m Metric) {
	d.vals = append(d.vals, m)
}

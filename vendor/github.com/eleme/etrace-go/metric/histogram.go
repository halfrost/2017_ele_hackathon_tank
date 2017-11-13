package metric

import "github.com/mailru/easyjson/jwriter"

// Histogram for timer histogram.
type Histogram struct {
	baseMetric
	baseNumber       int64
	min              int64
	max              int64
	sum              int64
	count            int64
	distributionType int
	values           []int64
}

// FromTimer transform a timer into histogram.
func FromTimer(t *Timer) *Histogram {
	h := &Histogram{
		baseMetric:       t.baseMetric,
		count:            t.count,
		min:              t.min,
		max:              t.max,
		sum:              t.sum,
		values:           make([]int64, MaxSlot, MaxSlot),
		distributionType: 1,
	}
	h.baseMetric.setType(TypeHistogram)
	idx := GetPercentilIndex(h.sum)
	h.values[idx] = 1
	return h
}

// Merge merge two histogram with same name and same tags;
func (h *Histogram) Merge(other Metric) {
	if tt, ok := other.(*Timer); ok {
		h.sum = h.sum + tt.sum
		h.count = h.count + tt.count
		if h.max < tt.max {
			h.max = tt.max
		}
		if h.min > tt.min {
			h.min = tt.min
		}
		idx := GetPercentilIndex(h.sum)
		h.values[idx] = h.values[idx] + 1
	} else if hh, ok := other.(*Histogram); ok {
		h.sum = h.sum + hh.sum
		h.count = h.count + hh.count
		if h.max < hh.max {
			h.max = hh.max
		}
		if h.min > hh.min {
			h.min = hh.min
		}
		for i, v := range hh.values {
			h.values[i] += v
		}
	}
}

// MarshalJSON encode histogram into json format.
func (h *Histogram) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	h.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode histogram using easy json writer.
func (h *Histogram) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawString(`["histogram",`)
	w.String(h.name)
	w.RawByte(',')
	w.Int64(h.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	h.tag.MarshalEasyJSON(w)
	w.RawByte(',')
	w.Int64(h.baseNumber)
	w.RawByte(',')
	w.Int64(h.min)
	w.RawByte(',')
	w.Int64(h.max)
	w.RawByte(',')
	w.Int64(h.sum)
	w.RawByte(',')
	w.Int64(h.count)
	w.RawByte(',')
	w.Int(h.distributionType)
	w.RawString(",[")
	childFirst := true
	for _, val := range h.values {
		if !childFirst {
			w.RawByte(',')
		}
		childFirst = false
		w.Int64(val)
	}
	w.RawString("]]")
}

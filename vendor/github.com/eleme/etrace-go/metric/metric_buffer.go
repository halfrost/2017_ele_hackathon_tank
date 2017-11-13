package metric

import (
	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// buffer for metric container with same metric name.
type buffer struct {
	cfg     *config.Config
	metrics map[string]*pkg
}

//  creates a new metric buffer.
func newBuffer(cfg *config.Config) *buffer {
	return &buffer{
		cfg:     cfg,
		metrics: make(map[string]*pkg),
	}
}

func (b *buffer) addMetric(me Metric) {
	name := me.Name()
	p, ok := b.metrics[name]
	if !ok {
		p = newPkg(b.cfg, name)
		b.metrics[name] = p
	}
	p.addMetric(me)
}

func (b *buffer) packages() map[string]*pkg {
	return b.metrics
}

func (b *buffer) reset() {
	b.metrics = make(map[string]*pkg)
}

func (b *buffer) count() int {
	return len(b.metrics)
}

// pkg is metric container with same metrics.
type pkg struct {
	cfg     *config.Config
	h       *Header
	name    string
	metrics map[Hash]Metric
}

// newPkg creates a new metric package.
func newPkg(cfg *config.Config, name string) *pkg {
	p := &pkg{
		cfg:     cfg,
		h:       NewHeader(cfg, name),
		name:    name,
		metrics: make(map[Hash]Metric),
	}
	return p
}

// Count returns current number of pending metrics' .
func (p *pkg) Count() int {
	return len(p.metrics)
}

func (p *pkg) reset() {
	p.metrics = make(map[Hash]Metric)
}

func (p *pkg) addMetric(me Metric) {
	if tt, ok := me.(*Timer); ok && tt.UpperEnabled() {
		me = FromTimer(tt)
	}
	hash := me.Hash()
	lastM, ok := p.metrics[hash]
	if ok {
		lastM.Merge(me)
	} else {
		p.metrics[hash] = me
	}
}

// BuildMessage encode header and metrics into bytes
func (p *pkg) BuildMessage() (header, message []byte, err error) {
	header, err = p.h.MarshalJSON()
	if err != nil {
		return
	}
	message, err = p.MarshalJSON()
	return
}

// MarshalJSON encode metric buffer into json format.
func (p *pkg) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	p.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode metric buffer using easy json writer.
func (p *pkg) MarshalEasyJSON(w *jwriter.Writer) {
	cfg := p.cfg
	vals := []string{
		"test",
		cfg.AppID,
		cfg.HostIP,
		cfg.HostName,
		cfg.Cluster,
		cfg.EZone,
		cfg.IDC,
		cfg.MesosTaskID,
		cfg.EleapposLabel,
		cfg.EleapposSlaveFqdn,
	}
	w.RawString("[[")
	for _, val := range vals {
		w.String(val)
		w.RawByte(',')
	}
	w.RawByte('[')
	childFirst := true
	for _, m := range p.metrics {
		if !childFirst {
			w.RawByte(',')
		}
		childFirst = false
		m.MarshalEasyJSON(w)
	}
	w.RawString("]]]")
}

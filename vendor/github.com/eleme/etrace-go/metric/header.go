package metric

import (
	"time"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// Header is for etrace header.
type Header struct {
	cfg       *config.Config
	name      string
	timestamp time.Time
}

// NewHeader create a new header.
func NewHeader(cfg *config.Config, name string) *Header {
	return &Header{
		cfg:       cfg,
		name:      name,
		timestamp: time.Now(),
	}
}

// MarshalJSON encode header into json.
func (h *Header) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	h.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode header using easy json.
func (h *Header) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawString(`{"appId":`)
	w.String(h.cfg.AppID)
	w.RawString(`,"hostIp":`)
	w.String(h.cfg.HostIP)
	w.RawString(`,"hostName":`)
	w.String(h.cfg.HostName)
	w.RawString(`,"messageType":"Metric","ast":`)
	w.Int64(h.timestamp.UnixNano() / 1e6)
	w.RawString(`,"key":`)
	w.String(h.cfg.Topic + "##" + h.name)
	w.RawByte('}')
}

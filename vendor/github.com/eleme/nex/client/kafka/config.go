package kafka

import (
	"strings"
	"time"

	"github.com/damnever/cc"
)

// Config is the configurations about the application.
type Config struct {
	AppID                string
	StatsdURL            string
	Brokers              []string
	MetadataRetryBackoff time.Duration
	MetadataRetryMax     int
}

// ParseConfig parses nex configuration and huskar kafka settings.
func ParseConfig(nexCfg cc.Configer, jsonKafkaSettings string) (*Config, error) {
	kafkaCfg, err := cc.NewConfigFromJSON([]byte(jsonKafkaSettings))
	if err != nil {
		return nil, err
	}
	statsdURL := nexCfg.String("statsd_url")
	statsdURL = strings.TrimPrefix(statsdURL, "statsd://")
	brokers := strings.Split(kafkaCfg.String("brokers"), ",")
	cfg := Config{
		AppID:                nexCfg.String("app_name"),
		StatsdURL:            statsdURL,
		Brokers:              brokers,
		MetadataRetryBackoff: kafkaCfg.DurationOr("metadata_retry_backoff", 2000) * time.Millisecond,
		MetadataRetryMax:     kafkaCfg.IntOr("metadata_retry_max", 20000),
	}
	return &cfg, nil
}

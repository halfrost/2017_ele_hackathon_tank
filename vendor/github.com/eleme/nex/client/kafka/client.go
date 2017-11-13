package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/damnever/cc"
	statsd "github.com/eleme/go-metrics-statsd"
	"github.com/eleme/log"
)

// Client is the kafka client to produce message.
type Client interface {
	Producer
}

// New create a kafka client from nex configuration,it is a wrapper function for NewClient.
func New(nexCfg cc.Configer, jsonKafkaSettings string, logger log.SimpleLogger) (Client, error) {
	cfg, err := ParseConfig(nexCfg, jsonKafkaSettings)
	if err != nil {
		return nil, err
	}
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.RequiredAcks = sarama.WaitForAll
	kafkaCfg.Metadata.Retry.Backoff = cfg.MetadataRetryBackoff
	kafkaCfg.Metadata.Retry.Max = cfg.MetadataRetryMax
	return NewClient(cfg, kafkaCfg, logger)
}

// NewClient create a kafka from Config and sarama Config.
func NewClient(cfg *Config, kafkaCfg *sarama.Config, logger log.SimpleLogger) (Client, error) {
	p, err := NewProducer(kafkaCfg, cfg.Brokers, logger)
	if err != nil {
		return nil, err
	}
	go statsd.WithHostname(kafkaCfg.MetricRegistry, cfg.AppID, cfg.StatsdURL, true)
	return &client{
		cfg:      cfg,
		kafkaCfg: kafkaCfg,
		producer: p,
	}, nil
}

type client struct {
	cfg      *Config
	kafkaCfg *sarama.Config
	producer Producer
}

// Send send a message to the kafka servers.
func (c *client) Send(ctx context.Context, topic, key, value string) error {
	return c.producer.Send(ctx, topic, key, value)
}

// Close close the client.
func (c *client) Close() {
	c.producer.Close()
}

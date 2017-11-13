package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/eleme/log"
)

// Producer is the abstraction for kafka producer.
type Producer interface {
	Send(ctx context.Context, topic, key, value string) error
	Close()
}

type producer struct {
	logger   log.SimpleLogger
	producer sarama.AsyncProducer
	closeCh  chan struct{}
}

// NewProducer creates a new kafka producer.
func NewProducer(cfg *sarama.Config, brokers []string, logger log.SimpleLogger) (Producer, error) {
	asyncProducer, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	p := &producer{
		logger:   logger,
		producer: asyncProducer,
		closeCh:  make(chan struct{}),
	}
	go p.loop()
	return p, nil
}

func (p *producer) Send(ctx context.Context, topic, key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}
	select {
	case p.producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *producer) loop() {
	for {
		select {
		case err := <-p.producer.Errors():
			p.logger.Errorf("kafka producer error:%v", err)
		case <-p.closeCh:
			return
		}
	}
}

func (p *producer) Close() {
	close(p.closeCh)
	p.producer.AsyncClose()
}

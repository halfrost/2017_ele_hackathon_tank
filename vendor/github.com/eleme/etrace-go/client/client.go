package client

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/eleme/etrace-go/client/thrift/gen-go/etrace"
	"github.com/eleme/etrace-go/config"

	"git.apache.org/thrift.git/lib/go/thrift"
)

var errCollectorsEmtpy = errors.New("collector list empty")

type param struct {
	head    []byte
	message []byte
}

// Client is thrift client for send messages.
type Client interface {
	Send(head, message []byte)
}

type client struct {
	sync.Mutex
	cfg  *config.Config
	msgC chan *param
}

// New creates a new thrift client.
func New(cfg *config.Config) Client {
	c := &client{
		cfg:  cfg,
		msgC: make(chan *param, 8),
	}
	go c.work()
	return c
}

// Send accept header and message for sending.
func (c *client) Send(head, message []byte) {
	msg := param{
		head:    head,
		message: message,
	}
	select {
	case c.msgC <- &msg:
	default:
	}
}

func (c *client) work() {
	var err error
	var client *etrace.MessageServiceClient
	for {
		select {
		case msg := <-c.msgC:
			if client == nil {
				client, err = c.newThriftClient()
			}
			if err != nil {
				continue
			}
			err = client.Send(msg.head, msg.message)
			if err != nil {
				client.Transport.Close()
				client = nil
			}
		}
	}
}

func (c *client) newThriftClient() (*etrace.MessageServiceClient, error) {
	collectors := c.cfg.Remoter.Collectors()
	if len(collectors) == 0 {
		return nil, errCollectorsEmtpy
	}
	idx := rand.Intn(len(collectors))
	collector := collectors[idx]
	var transport thrift.TTransport
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	transport, err := thrift.NewTSocketTimeout(fmt.Sprintf("%s:%d", collector.IP, collector.Port), 5000*time.Millisecond)
	if err != nil {
		return nil, err
	}
	transport = transportFactory.GetTransport(transport)
	err = transport.Open()
	if err != nil {
		transport.Close()
		return nil, err
	}
	client := etrace.NewMessageServiceClientFactory(transport, protocolFactory)
	return client, nil
}

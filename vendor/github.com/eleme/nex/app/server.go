package app

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/eleme/log"
	"github.com/eleme/nex/consts"
)

// TNexServerConfig is the TNexServer required configurations.
type TNexServerConfig struct {
	Addr             string
	ClientTimeout    time.Duration
	GracefulTimeout  time.Duration
	Logger           log.SimpleLogger
	ProcessorFactory thrift.TProcessorFactory
}

// TNexServer represents a thrift server.
type TNexServer struct {
	sync.WaitGroup
	quit   chan struct{}
	config *TNexServerConfig

	serverTransport        thrift.TServerTransport
	inputTransportFactory  thrift.TTransportFactory
	outputTransportFactory thrift.TTransportFactory
	inputProtocolFactory   thrift.TProtocolFactory
	outputProtocolFactory  thrift.TProtocolFactory
}

// NewTNexServer creates a new TNexServer.
func NewTNexServer(config *TNexServerConfig) *TNexServer {
	transportFactory := thrift.NewTBufferedTransportFactory(consts.BufferSize)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	return &TNexServer{
		inputTransportFactory:  transportFactory,
		outputTransportFactory: transportFactory,
		inputProtocolFactory:   protocolFactory,
		outputProtocolFactory:  protocolFactory,
		config:                 config,
		quit:                   make(chan struct{}),
	}
}

// Serve starts the server.
func (s *TNexServer) Serve() error {
	if err := s.Listen(); err != nil {
		return err
	}
	return s.AcceptLoop()
}

// Listen opens a listening socket.
func (s *TNexServer) Listen() error {
	serverTransport, err := thrift.NewTServerSocketTimeout(s.config.Addr, s.config.ClientTimeout)
	if err != nil {
		return err
	}
	s.serverTransport = serverTransport
	if err := s.serverTransport.Listen(); err != nil {
		return err
	}
	return nil
}

// AcceptLoop loops and accepts connections.
func (s *TNexServer) AcceptLoop() error {
	l := s.config.Logger
	for {
		client, err := s.serverTransport.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
			}
			return err
		}
		if client != nil {
			s.Add(1)
			expCounters.Add("AcceptConnections", 1)
			go func() {
				defer s.Done()
				defer func() {
					if e := recover(); e != nil {
						l.Errorf("Panic in processor: %s: %s", e, string(debug.Stack()))
					}
				}()
				if err := s.processRequests(client); err != nil {
					l.Errorf("Error processing request: %v", err)
				}
			}()
		}
	}
}

// Stop stops the server.
func (s *TNexServer) Stop() {
	select {
	case <-s.quit: // already closed
		return
	default:
	}

	close(s.quit)
	s.serverTransport.Interrupt()

	timer := time.NewTimer(s.config.GracefulTimeout)
	waitCh := make(chan struct{})
	go func() {
		s.Wait()
		close(waitCh)
	}()
	select {
	case <-waitCh:
	case <-timer.C:
	}
}

func (s *TNexServer) processRequests(client thrift.TTransport) error {
	processor := s.config.ProcessorFactory.GetProcessor(client)
	inputTransport := s.inputTransportFactory.GetTransport(client)
	outputTransport := s.outputTransportFactory.GetTransport(client)
	inputProtocol := s.inputProtocolFactory.GetProtocol(inputTransport)
	outputProtocol := s.outputProtocolFactory.GetProtocol(outputTransport)

	if inputTransport != nil {
		defer inputTransport.Close()
	}
	if outputTransport != nil {
		defer outputTransport.Close()
	}

	for {
		select {
		case <-s.quit:
			return nil
		default:
		}

		ok, err := processor.Process(inputProtocol, outputProtocol)
		if err != nil {
			switch x := err.(type) {
			case thrift.TTransportException:
				if x.TypeId() == thrift.END_OF_FILE {
					return nil
				}
				return x
			case thrift.TApplicationException:
				if x.TypeId() == thrift.UNKNOWN_METHOD {
					continue
				}
			default: // user defined exceptions
			}
		}
		if !ok {
			break
		}
	}
	return nil
}

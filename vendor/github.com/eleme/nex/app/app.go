package app

import (
	"time"

	"github.com/damnever/cc"
	"github.com/eleme/huskar/service"
	"github.com/eleme/log"
	"github.com/eleme/nex/hook"

	"github.com/apache/thrift/lib/go/thrift"
)

var (
	defaultGracefulTimeout int64 = 3  // sec
	defaultClientTimeout   int64 = 20 // min
)

// ThriftApp is an implementation of Application interface,
// featuring Thrift with BufferedTransport and BinaryProtocol
type ThriftApp struct {
	Core
	Name string
	addr string
	log.SimpleLogger
	server    *TNexServer
	nexConfig cc.Configer
}

// NewThriftApplication creates an Thrift Application with given processor
func NewThriftApplication(nexConfig cc.Configer, logger log.SimpleLogger,
	registrator service.Registrator, processorFactory thrift.TProcessorFactory) *ThriftApp {
	addr := nexConfig.String("addr")
	serverConfig := &TNexServerConfig{
		Addr:             addr,
		ClientTimeout:    nexConfig.DurationOr("client_timeout", defaultClientTimeout) * time.Minute,
		GracefulTimeout:  nexConfig.DurationOr("graceful_timeout", defaultGracefulTimeout) * time.Second,
		Logger:           logger,
		ProcessorFactory: processorFactory,
	}
	app := &ThriftApp{
		Name:         nexConfig.String("app_name"),
		addr:         addr,
		SimpleLogger: logger,
		nexConfig:    nexConfig,
		server:       NewTNexServer(serverConfig),
	}
	app.Registrator = registrator
	return app
}

// Run starts the Thrift server, blocks forever.
func (ta *ThriftApp) Run() (err error) {
	hook.BeforeServerStarting.Notify(nil)

	if err = ta.server.Listen(); err != nil {
		ta.Errorf("Thrift application opening listening socket: %s failed: %s", ta.addr, err)
		return
	}
	ta.Infof("Starting %s server at %s", ta.Name, ta.addr)

	serviceInstance, err := ta.MakeHuskarServiceInstance(ta.addr, ta.nexConfig.String("version"))
	if err != nil {
		ta.Errorf("Make Huskar service instance failed: %v", err)
		return
	}
	ta.Infof("Register application: %s with key: %s", ta.Name, serviceInstance.Key)
	if err := ta.Register(serviceInstance); err != nil {
		ta.Errorf("Register application failed: %v", err)
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		if err := ta.server.AcceptLoop(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err = <-errCh:
		ta.Errorf("Got unexpected error: %v", err)
	case sig := <-ta.WatchSignals():
		ta.Infof("Got signal: %v", sig)
	}

	ta.Infof("Deregister application: %s with key: %s", ta.Name, serviceInstance.Key)
	if errd := ta.Deregister(serviceInstance); errd != nil {
		ta.Errorf("Deregister application failed: %v", errd)
	} else if err == nil {
		hook.BeforeServerStoping.Notify(nil)
		ta.Info("Graceful shutdown..")
		ta.server.Stop()
		return nil
	}
	return err
}

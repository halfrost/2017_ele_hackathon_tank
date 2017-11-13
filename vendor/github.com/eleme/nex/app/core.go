package app

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/eleme/common"
	"github.com/eleme/huskar/service"
	"github.com/eleme/nex/cmd/debug"
	"github.com/eleme/nex/consts"
	"github.com/eleme/nex/utils"
)

const (
	registerRetryCnt int    = 10
	pprofAddr        string = "0.0.0.0:4455"
)

var (
	expCounters = expvar.NewMap("counters")
)

// Core provides basic capabilities shared among different type of
// applications
type Core struct {
	service.Registrator
	watched int32
	sigCh   chan os.Signal
}

// WatchSignals watch termination signals and debug siginals.
func (c *Core) WatchSignals() <-chan os.Signal {
	if !atomic.CompareAndSwapInt32(&c.watched, 0, 1) {
		return c.sigCh
	}
	c.installDebugSignal()
	c.startHTTPDebuger()

	c.sigCh = make(chan os.Signal)
	signal.Notify(c.sigCh, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	return c.sigCh
}

func (c *Core) startHTTPDebuger() {
	pprofHandler := http.NewServeMux()
	pprofHandler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	pprofHandler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	pprofHandler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	pprofHandler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	pprofHandler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	expvarHandler := expvar.Handler()
	pprofHandler.Handle("/debug/vars", expvarHandler)
	pprofHandler.Handle("/debug/pprof/vars", expvarHandler) // alias

	server := &http.Server{Addr: pprofAddr, Handler: pprofHandler}
	go server.ListenAndServe()
}

// debug signal handler.
func (c *Core) installDebugSignal() {
	go func() {
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, syscall.SIGTTIN, syscall.SIGTTOU)
		var server *debug.Server

		for {
			s := <-sigCh
			switch s {
			case syscall.SIGTTIN:
				if server == nil {
					server = debug.NewServer()
					server.Start()
				}
			case syscall.SIGTTOU:
				if server != nil {
					server.Close()
					server = nil
				}
			}
		}
	}()
}

// Register registers this application with key.
func (c *Core) Register(instance *service.Instance) error {
	retry := func() error { return c.Registrator.Register("", "", instance) }
	return utils.Retry(retry, registerRetryCnt)
}

// Deregister deregisters this application.
func (c *Core) Deregister(instance *service.Instance) error {
	retry := func() error { return c.Registrator.Deregister("", "", instance.Key) }
	return utils.Retry(retry, registerRetryCnt)
}

// MakeHuskarServiceInstance creates a huskar service instance by address.
func (c *Core) MakeHuskarServiceInstance(bind string, version string) (*service.Instance, error) {
	addr, err := net.ResolveTCPAddr("tcp4", bind)
	if err != nil {
		return nil, err
	}
	ip := common.GetLocalIP()
	if ip == "" {
		return nil, fmt.Errorf("get local IP failed")
	}

	var key string
	if containerID := os.Getenv(consts.EnvDockerContainerID); containerID != "" {
		key = containerID
	} else {
		key = fmt.Sprintf("%s_%d", ip, addr.Port)
	}

	nCPU, err := strconv.Atoi(os.Getenv("GOMAXPROCS")) // appos(docker)
	if err != nil || nCPU < 1 {
		nCPU = runtime.NumCPU()
	}

	return &service.Instance{
		Key: key,
		Value: &service.StaticInfo{
			IP: ip,
			Port: map[string]int{
				"main": addr.Port,
			},
			State: service.StateUp,
			Meta: map[string]interface{}{
				"weight":     nCPU*2 + 1,
				"protocol":   "thrift",
				"soaFx":      "nex",
				"soaVersion": version,
			},
		},
	}, nil
}

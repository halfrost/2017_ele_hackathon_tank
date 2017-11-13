package etrace

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/eleme/etrace-go/client/thrift/gen-go/etrace"
)

// MockMessage is the message that MockServer acts on.
type MockMessage struct {
	Head    []byte
	Message []byte
}

// MockServer is an local etrace mock server.
type MockServer struct {
	sync.Mutex
	httpLn           net.Listener
	server           *http.Server
	thriftServer     *thrift.TSimpleServer
	thriftServerPort int
	msgCh            []chan MockMessage
}

// NewMockServer create a MockServer.
func NewMockServer() *MockServer {
	s := &MockServer{
		msgCh: make([]chan MockMessage, 0),
	}
	s.initHTTPServer()
	s.initThriftServer()
	return s
}

func (s *MockServer) initHTTPServer() {
	s.httpLn, _ = net.Listen("tcp", "127.0.0.1:0")
	mux := http.NewServeMux()
	mux.HandleFunc("/agent-config", s.agentConfigHandler)
	mux.HandleFunc("/collector", s.collectorsHandler)
	s.server = &http.Server{Handler: mux, Addr: ""}
}

func (s *MockServer) initThriftServer() {
	var transport *thrift.TServerSocket
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	s.thriftServerPort = getFreePort()
	transport, _ = thrift.NewTServerSocket(fmt.Sprintf("127.0.0.1:%d", s.thriftServerPort))
	processor := etrace.NewMessageServiceProcessor(s)
	s.thriftServer = thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

}

// SubMsg subscribe the messages that thrift server handler received.
func (s *MockServer) SubMsg(ch chan MockMessage) {
	s.Lock()
	defer s.Unlock()
	s.msgCh = append(s.msgCh, ch)
}

// Start will start the mock server.
func (s *MockServer) Start() {
	go s.server.Serve(s.httpLn)
	go s.thriftServer.Serve()
	time.Sleep(time.Second)
}

// Close will close the mock server.
func (s *MockServer) Close() {
	s.thriftServer.Stop()
	s.thriftServer.ServerTransport().Close()
	s.httpLn.Close()
}

// HTTPServerPort is the port that mock server http  service listens on.
func (s *MockServer) HTTPServerPort() int {
	return s.httpLn.Addr().(*net.TCPAddr).Port
}

// ThriftServerPort is the port that mock server thrift service listens on.
func (s *MockServer) ThriftServerPort() int {
	return s.thriftServerPort
}

type agentConfig struct {
	Enable       bool `json:"enable"`
	MessageCount int  `json:"messageCount"`
}

func (s *MockServer) agentConfigHandler(w http.ResponseWriter, r *http.Request) {
	buf, _ := json.Marshal(&agentConfig{
		Enable:       true,
		MessageCount: 1,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

type collector struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func (s *MockServer) collectorsHandler(w http.ResponseWriter, r *http.Request) {
	collectors := make([]collector, 0)
	collectors = append(collectors, collector{
		IP:   "127.0.0.1",
		Port: s.ThriftServerPort(),
	})
	buf, _ := json.Marshal(collectors)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

// Send is the handler when thrift service receives remote call.
func (s *MockServer) Send(head []byte, message []byte) error {
	s.Lock()
	defer s.Unlock()
	msg := MockMessage{
		Head:    head,
		Message: message,
	}
	for _, ch := range s.msgCh {
		select {
		case ch <- msg:
		default:
		}
	}
	return nil
}

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

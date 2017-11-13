package debug

import (
	"net"
	"net/rpc"
	"runtime"

	"github.com/eleme/log"
)

const (
	// RPCBind is bind address for debug server.
	RPCBind = ":5566"
)

// Server represents a debug server.
type Server struct {
	log.SimpleLogger
	*rpc.Server
	listen net.Listener
}

// NewServer returns a new debug server.
func NewServer() *Server {
	return &Server{
		SimpleLogger: log.New("Debugger"),
		Server:       rpc.NewServer(),
	}
}

// Start listen and accept RPC requests.
func (s *Server) Start() {
	s.Register(&Debugger{})
	l, err := net.Listen("tcp", RPCBind)
	if err != nil {
		s.Debugf("Debug server listen failed: %s", err)
		return
	}
	s.listen = l
	go s.Accept(l)
}

// Accept override the rpc.Server Accept method.
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		go s.ServeConn(conn)
	}
}

// Close close the debug server.
func (s *Server) Close() error {
	return s.listen.Close()
}

// Debugger represents an object of debug server will report.
type Debugger struct {
}

// Args represents the arguments of RPC method.
type Args struct {
}

// Reply represents the result of RPC method.
type Reply struct {
	Value string
}

// Ping just test network is ok.
func (d *Debugger) Ping(args *Args, reply *Reply) error {
	reply.Value = "Pong"
	return nil
}

// PrintStack return all goroutines stack info.
func (d *Debugger) PrintStack(args *Args, reply *Reply) error {
	stack := StackTrace(true)
	reply.Value = stack
	return nil
}

// StackTrace collect all goroutines stack info.
func StackTrace(all bool) string {
	buf := make([]byte, 10240)
	for {
		size := runtime.Stack(buf, all)
		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break
	}
	return string(buf)
}

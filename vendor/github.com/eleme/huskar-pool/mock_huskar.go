package huskarPool

import (
	"encoding/json"
	"net/http"
	"time"

	huskarService "github.com/eleme/huskar/service"
)

type mockServer struct {
	ins map[string]map[string]map[string]*huskarService.Instance
}

func newMock() *mockServer {
	return &mockServer{
		ins: make(map[string]map[string]map[string]*huskarService.Instance),
	}
}

func (m *mockServer) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handle)
	http.ListenAndServe(":8080", mux)
}

func (m *mockServer) handle(w http.ResponseWriter, r *http.Request) {
	buf, _ := json.Marshal(m.ins)
	w.Write(buf)
	time.Sleep(time.Minute)
}

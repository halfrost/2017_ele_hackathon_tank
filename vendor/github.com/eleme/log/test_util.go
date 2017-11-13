package log

import (
	"bytes"
	"sync"
)

type fakeWriter struct {
	writed chan bool
	buf    *bytes.Buffer
}

func (f *fakeWriter) Write(p []byte) (n int, err error) {
	f.buf.Write(p)
	f.writed <- true
	return
}

func (f *fakeWriter) String() string {
	return f.buf.String()
}

type appendWriter struct {
	lines []string
	l     *sync.Mutex
}

func (w *appendWriter) Write(p []byte) (n int, err error) {
	w.l.Lock()
	w.lines = append(w.lines, string(p))
	w.l.Unlock()

	return len(p), nil
}

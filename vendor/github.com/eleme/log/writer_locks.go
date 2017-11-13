package log

import (
	"io"
	"sync"
)

type writerLocker struct {
	m  map[io.Writer]*sync.Mutex
	mu sync.RWMutex
}

func (wl *writerLocker) Lock(w io.Writer) {
	wl.mu.RLock()
	if l, ok := wl.m[w]; ok {
		wl.mu.RUnlock()
		l.Lock()
	} else {
		wl.mu.RUnlock()
		// add new lock to map
		var newLock sync.Mutex
		wl.mu.Lock()
		wl.m[w] = &newLock
		wl.mu.Unlock()
		// lock it
		newLock.Lock()
	}
}

func (wl *writerLocker) Unlock(w io.Writer) {
	wl.mu.RLock()
	l, ok := wl.m[w]
	wl.mu.RUnlock()

	if ok {
		l.Unlock()
	}
}

func newWriterLocker() *writerLocker {
	return &writerLocker{
		m: make(map[io.Writer]*sync.Mutex),
	}
}

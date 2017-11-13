package log

import (
	"io"
	"sync"

	"github.com/tevino/abool"
)

type writerWorker struct {
	callbacks chan func()
	wg        sync.WaitGroup
	closing   chan bool
	closed    chan bool
}

func (w *writerWorker) Push(f func()) {
	w.wg.Add(1)
	select {
	case w.callbacks <- f:
	default:
		//throw message if full
		w.wg.Done()
	}
}

func (w *writerWorker) Start() {
	go func() {
		for fn := range w.callbacks {
			fn()
			w.wg.Done()
		}
		close(w.closed)
	}()
}

func (w *writerWorker) WaitClose() {
	close(w.closing)
	w.wg.Wait()
	close(w.callbacks) // Notify the worker to exit
	<-w.closed
}

type writerSupervisor struct {
	m          map[io.Writer]*writerWorker
	mu         sync.RWMutex
	closed     *abool.AtomicBool
	bufferSize int
}

func (ws *writerSupervisor) WaitClose() {
	if ws.closed.IsSet() {
		return
	}

	ws.closed.Set()

	ws.mu.RLock()
	defer ws.mu.RUnlock()
	for _, worker := range ws.m {
		worker.WaitClose()
	}
}

func (ws *writerSupervisor) Do(w io.Writer, f func()) {
	if ws.closed.IsSet() {
		return
	}

	ws.mu.RLock()
	worker, ok := ws.m[w]
	ws.mu.RUnlock()

	if !ok {
		worker = &writerWorker{
			callbacks: make(chan func(), ws.bufferSize),
			closing:   make(chan bool),
			closed:    make(chan bool),
		}

		ws.mu.Lock()
		if currentWorker, ok := ws.m[w]; ok {
			worker = currentWorker
		} else {
			ws.m[w] = worker
			worker.Start()
		}
		ws.mu.Unlock()
	}

	worker.Push(f)
}

func newWriterSupervisor(bufferSize int) *writerSupervisor {
	return &writerSupervisor{
		m:          make(map[io.Writer]*writerWorker),
		mu:         sync.RWMutex{},
		closed:     abool.New(),
		bufferSize: bufferSize,
	}
}

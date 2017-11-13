package log

import "io"

var (
	writerLocks       *writerLocker
	wSupervisor       *writerSupervisor
	defaultBufferSize = 1024
)

func init() {
	writerLocks = newWriterLocker()
	wSupervisor = newWriterSupervisor(defaultBufferSize)
}

// SetBufferSize sets the default buffer size for async logging
// Default to 1024
// NOTE: Call this before logging anything
func SetBufferSize(size int) {
	wSupervisor.bufferSize = size
}

//Wait wait for all writer worker close
func Wait() {
	wSupervisor.WaitClose()
}

// Handler represents a handler of Record
type Handler interface {
	Log(record Record)
	Writer() io.Writer
}

// StreamHandler is a Handler of Stream writer e.g. console
type StreamHandler struct {
	writer io.Writer
	Formatter
}

// NewStreamHandler creates a StreamHandler with given writer(usually os.Stdout)
// and format string, whether to color the output is determined by the type of
// writer
func NewStreamHandler(w io.Writer, f Formatter) *StreamHandler {
	h := new(StreamHandler)
	h.writer = w
	h.Formatter = f
	return h
}

// Colored enable or disable the color function of internal format, usually
// this is determined automatically
//
// When called with no argument, it returns the current state of color function
func (sw *StreamHandler) Colored(ok ...bool) bool {
	if len(ok) > 0 {
		sw.Formatter.SetColored(ok[0])
	}
	return sw.Formatter.Colored()
}

// Log print the Record to the internal writer
func (sw *StreamHandler) Log(record Record) {
	b := sw.Formatter.Format(record)
	writerLocks.Lock(sw.writer)
	defer writerLocks.Unlock(sw.writer)
	sw.writer.Write(b)
}

// Writer return the writer
func (sw *StreamHandler) Writer() io.Writer {
	return sw.writer
}

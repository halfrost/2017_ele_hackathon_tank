package statsd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var logger Logger

func init() {
	logger = log.New(os.Stderr, "", log.LstdFlags)
}

// Config is used to configure metrics settings
type Config struct {
	StatsdAddr           string        // Statsd udp address, (host:port)
	Prefix               []string      // Prefixed with keys to seperate services
	HostName             string        // Hostname to use. If not provided and EnableHostname, it will be os.Hostname
	EnableRuntimeMetrics bool          // Enables profiling of runtime metrics (GC, Goroutines, Memory)
	EnableCPUMetrics     bool          // Enables self cpu metrics (utime, stime)
	Interval             time.Duration // Collect metrics interval.
	ScheduleSend         bool          // Schedule(default) or Instant send.
	FlushInterval        int           // Flush metrics interval. (ms)
	StatsdMaxLen         int           // Maximum size of a packet to send to statsd.
}

// DefaultConfig provides a sane default configuration
func DefaultConfig(statsAddr string, prefix ...string) *Config {
	c := &Config{
		StatsdAddr:           statsAddr,
		Prefix:               prefix, // Use client provided service
		EnableRuntimeMetrics: true,   // Enable runtime profiling
		EnableCPUMetrics:     true,
		Interval:             time.Second, // Poll runtime every second
		FlushInterval:        100,
		StatsdMaxLen:         1400,
	}

	// Try to get the hostname
	name, err := Hostname()
	if err != nil {
		c.HostName = ""
	} else {
		c.HostName = name
	}
	return c
}

func Hostname() (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", err
	}
	name = strings.Replace(name, ".", "-", -1)
	return name, nil
}

type StatsdService struct {
	Config
	lastNumGC      uint32
	metricQueue    chan string
	done           chan bool
	running        int32
	runtimeMetrics map[string]struct{}
}

func NewStatsdService(conf *Config) *StatsdService {
	s := &StatsdService{
		Config:         *conf,
		metricQueue:    make(chan string, 4096),
		done:           make(chan bool),
		running:        1,
		runtimeMetrics: make(map[string]struct{}),
	}

	if conf.EnableRuntimeMetrics {
		for _, metric := range defaultRuntimeMetrics {
			s.runtimeMetrics[metric] = struct{}{}
		}
	}
	return s
}

func (s *StatsdService) Start() {
	go s.flushMetrics()
	go s.collectStats()
	atomic.StoreInt32(&s.running, 1)
}

// Periodically collects stats to publish
func (s *StatsdService) collectStats() {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			if s.EnableRuntimeMetrics {
				s.emitRuntimeStats()
			}
			if s.EnableCPUMetrics {
				s.emitCPUStats()
			}
		}
	}
}

func (s *StatsdService) EnableHostname() {
	if s.HostName == "" {
		s.HostName, _ = Hostname()
	}
}

func (s *StatsdService) DisableHostname() {
	s.HostName = ""
}

func (s *StatsdService) Stop() error {
	atomic.StoreInt32(&s.running, 0)
	close(s.metricQueue)
	select {
	case <-s.done:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("statsd stop timeout")
	}
}

func (s *StatsdService) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// Flushes metrics
func (s *StatsdService) flushMetrics() {
	var sock net.Conn
	var err error
	var wait <-chan time.Time
	var ticker *time.Ticker
	if s.FlushInterval != 0 {
		ticker = time.NewTicker(time.Duration(s.FlushInterval) * time.Millisecond)
		defer ticker.Stop()
	}

CONNECT:
	// Create a buffer
	buf := bytes.NewBuffer(nil)

	// Attempt to connect
	sock, err = net.Dial("udp", s.StatsdAddr)
	if err != nil {
		logger.Printf("Error connecting to statsd! Err: %s", err)
		goto WAIT
	}

	if s.FlushInterval != 0 {
		for {
			select {
			case metric, ok := <-s.metricQueue:
				// Get a metric from the queue
				if !ok {
					goto QUIT
				}

				// Check if this would overflow the packet size
				if len(metric)+buf.Len() > s.StatsdMaxLen {
					_, err := sock.Write(buf.Bytes())
					buf.Reset()
					if err != nil {
						logger.Printf("Error writing to statsd! Err: %s", err)
						goto WAIT
					}
				}

				// Append to the buffer
				buf.WriteString(metric)

			case <-ticker.C:
				if buf.Len() == 0 {
					continue
				}

				_, err := sock.Write(buf.Bytes())
				buf.Reset()
				if err != nil {
					logger.Printf("Error flushing to statsd! Err: %s", err)
					goto WAIT
				}
			}
		}
	} else {
		for metric := range s.metricQueue {
			_, err := sock.Write([]byte(metric))
			if err != nil {
				logger.Printf("Error writing to statsd! Err: %s", err)
				goto WAIT
			}
		}
		close(s.done)
	}

WAIT:
	// Wait for a while
	wait = time.After(time.Duration(5) * time.Second)
	for {
		select {
		// Dequeue the messages to avoid backlog
		case _, ok := <-s.metricQueue:
			if !ok {
				goto QUIT
			}
		case <-wait:
			goto CONNECT
		}
	}
QUIT:
	if buf.Len() > 0 {
		_, err := sock.Write(buf.Bytes())
		if err != nil {
			logger.Printf("Error flushing to statsd! Err: %s", err)
		}
	}
	close(s.done)
}

// Inserts a string value at an index into the slice
func insert(i int, v string, s []string) []string {
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

/// Gauges ///

func (s *StatsdService) SetGaugeFloat64(key []string, val float64, isHostOn bool) {
	flatKey := s.mergeKey(key, isHostOn)
	s.pushMetric(fmt.Sprintf("%s:%f|g\n", flatKey, val))
}

func (s *StatsdService) SetGaugeInt(key []string, val int, isHostOn bool) {
	flatKey := s.mergeKey(key, isHostOn)
	s.pushMetric(fmt.Sprintf("%s:%d|g\n", flatKey, val))
}

func (s *StatsdService) SetGaugeUInt64(key []string, val uint64, isHostOn bool) {
	flatKey := s.mergeKey(key, isHostOn)
	s.pushMetric(fmt.Sprintf("%s:%d|g\n", flatKey, val))
}

/// Counters ///

func (s *StatsdService) IncrCounter(key []string, val int, isHostOn bool) {
	flatKey := s.mergeKey(key, isHostOn)
	s.pushMetric(fmt.Sprintf("%s:%d|c\n", flatKey, val))
}

/// Samples ///

func (s *StatsdService) AddSample(key []string, val float64, isHostOn bool) {
	flatKey := s.mergeKey(key, isHostOn)
	s.pushMetric(fmt.Sprintf("%s:%f|ms\n", flatKey, val))
}

func (s *StatsdService) MeasureSince(key []string, start time.Time, isHostOn bool) time.Duration {
	now := time.Now()
	elapsed := now.Sub(start)
	msec := float64(elapsed.Nanoseconds()) / float64(time.Millisecond)
	s.AddSample(key, msec, isHostOn)
	return elapsed
}

func (s *StatsdService) mergeKey(key []string, isHostOn bool) string {
	if s.HostName != "" && isHostOn {
		key = insert(0, s.HostName, key)
	}

	if len(s.Prefix) != 0 {
		key = insert(0, s.flattenKey(s.Prefix), key)
	}
	flatKey := s.flattenKey(key)
	return flatKey
}

// Flattens the key for formatting, removes spaces
func (s *StatsdService) flattenKey(parts []string) string {
	joined := strings.Join(parts, ".")
	return strings.Map(func(r rune) rune {
		switch r {
		case ':':
			fallthrough
		case ' ':
			return '_'
		default:
			return r
		}
	}, joined)
}

// Does a non-blocking push to the metrics queue
func (s *StatsdService) pushMetric(m string) {
	select {
	case s.metricQueue <- m:
	default:
	}
}

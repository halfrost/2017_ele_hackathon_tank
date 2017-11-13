package statsd

import (
	"runtime"
	"time"
)

var (
	defaultRuntimeMetrics = []string{"num_goroutines", "gc_pause_ms", "gc_pause_total_ms", "alloc_bytes",
		"total_alloc_bytes", "sys_bytes", "heap_alloc_bytes", "heap_sys_bytes", "heap_idle_bytes",
		"heap_inuse_bytes", "heap_released_bytes", "heap_objects", "stack_inuse_bytes", "stack_sys_bytes",
		"num_gc", "lookups", "mallocs", "frees"}
)

// Emits various runtime statsitics
func (s *StatsdService) emitRuntimeStats() {
	// Export memory stats
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	// Export info about the last few GC runs
	num := stats.NumGC

	for metric := range s.runtimeMetrics {
		switch metric {
		case "num_goroutines":
			// Export number of Goroutines
			numRoutines := runtime.NumGoroutine()
			s.SetGaugeInt([]string{"runtime", "num_goroutines"}, numRoutines, true)
		case "gc_pause_ms":
			// Handle wrap around
			if num < s.lastNumGC {
				s.lastNumGC = 0
			}
			// Ensure we don't scan more than 256
			if num-s.lastNumGC >= 256 {
				s.lastNumGC = num - 255
			}
			for i := s.lastNumGC; i < num; i++ {
				pause := stats.PauseNs[i%256]
				s.AddSample([]string{"runtime", "gc_pause_ms"}, float64(pause)/float64(time.Millisecond), true)
			}
			s.lastNumGC = num
		case "gc_pause_total_ms":
			s.SetGaugeUInt64([]string{"runtime", "gc_pause_total_ms"}, stats.PauseTotalNs/uint64(time.Millisecond), true)
		case "alloc_bytes":
			s.SetGaugeUInt64([]string{"runtime", "alloc_bytes"}, stats.Alloc, true)
		case "total_alloc_bytes":
			s.SetGaugeUInt64([]string{"runtime", "total_alloc_bytes"}, stats.TotalAlloc, true)
		case "sys_bytes":
			s.SetGaugeUInt64([]string{"runtime", "sys_bytes"}, stats.Sys, true)
		case "heap_alloc_bytes":
			s.SetGaugeUInt64([]string{"runtime", "heap_alloc_bytes"}, stats.HeapAlloc, true)
		case "heap_sys_bytes":
			s.SetGaugeUInt64([]string{"runtime", "heap_sys_bytes"}, stats.HeapSys, true)
		case "heap_idle_bytes":
			s.SetGaugeUInt64([]string{"runtime", "heap_idle_bytes"}, stats.HeapIdle, true)
		case "heap_inuse_bytes":
			s.SetGaugeUInt64([]string{"runtime", "heap_inuse_bytes"}, stats.HeapInuse, true)
		case "heap_released_bytes":
			s.SetGaugeUInt64([]string{"runtime", "heap_released_bytes"}, stats.HeapReleased, true)
		case "heap_objects":
			s.SetGaugeUInt64([]string{"runtime", "heap_objects"}, stats.HeapObjects, true)
		case "stack_inuse_bytes":
			s.SetGaugeUInt64([]string{"runtime", "stack_inuse_bytes"}, stats.StackInuse, true)
		case "stack_sys_bytes":
			s.SetGaugeUInt64([]string{"runtime", "stack_sys_bytes"}, stats.StackSys, true)
		case "num_gc":
			s.SetGaugeUInt64([]string{"runtime", "num_gc"}, uint64(num), true)
		case "lookups":
			s.SetGaugeUInt64([]string{"runtime", "lookups"}, stats.Lookups, true)
		case "mallocs":
			s.SetGaugeUInt64([]string{"runtime", "mallocs"}, stats.Mallocs, true)
		case "frees":
			s.SetGaugeUInt64([]string{"runtime", "frees"}, stats.Frees, true)
		}
	}
}

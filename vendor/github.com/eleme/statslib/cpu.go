package statsd

import (
	"syscall"
)

func (s *StatsdService) emitCPUStats() {
	usage := new(syscall.Rusage)
	syscall.Getrusage(syscall.RUSAGE_SELF, usage)
	stime := usage.Stime.Sec + int64(usage.Stime.Usec)/1e6
	utime := usage.Utime.Sec + int64(usage.Utime.Usec)/1e6
	s.SetGaugeUInt64([]string{"cpu_utime"}, uint64(stime), true)
	s.SetGaugeUInt64([]string{"cpu_stime"}, uint64(utime), true)
}

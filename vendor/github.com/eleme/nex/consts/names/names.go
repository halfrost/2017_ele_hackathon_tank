// Package names defines common use names for statsd, log...
package names

const (
	// OthErr defines a common error name in statsd.
	OthErr string = "err"
	// UserErr defines user exception name in statsd.
	UserErr string = "user_exc"
	// SysErr defines system exception name in statsd.
	SysErr string = "sys_exc"
	// UnkwnErr defines unknwon exception name in statsd.
	UnkwnErr string = "crit"
	// TimeoutErr defines timeout exception name in statsd.
	TimeoutErr string = "timeout"
	// NotHealthyErr is name for api not healthy in statsd.
	NotHealthyErr string = "sick"
)

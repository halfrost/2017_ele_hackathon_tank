package etrace

import (
	"fmt"
	"strings"
	"sync"
	"time"

	etrace "github.com/eleme/etrace-go"
)

// Ref: http://wiki.ele.to:8090/pages/viewpage.action?pageId=42743080

type redisTransaction struct {
	sync.Mutex
	etrace.Transactioner
	stats map[string]*redisStats
}

func newRedisTransaction(trans etrace.Transactioner) *redisTransaction {
	return &redisTransaction{
		Transactioner: trans,
		stats:         make(map[string]*redisStats),
	}
}

func (rs *redisTransaction) Merge(url, cmd string, success bool, duration time.Duration) {
	rs.Lock()
	if csm, ok := rs.stats[url]; ok {
		csm.merge(cmd, success, duration)
	} else {
		csm := newRedisStats(url, cmd, success, duration)
		rs.stats[url] = csm
	}
	rs.Unlock()
}

func (rs *redisTransaction) TagStatsAndSum() time.Duration {
	rs.Lock()
	var totalDuration int64
	for url, stats := range rs.stats {
		totalDuration += stats.totalDuration
		rs.AddTag(url, stats.jsonString())
	}
	rs.Unlock()
	return time.Duration(totalDuration)
}

type redisStats struct {
	url           string
	timestamp     int64
	cmds          map[string]*redisCmdStats
	totalDuration int64
}

func newRedisStats(url string, cmd string, success bool, duration time.Duration) *redisStats {
	rus := &redisStats{
		url:       url,
		timestamp: time.Now().UnixNano() / 1e6,
		cmds:      make(map[string]*redisCmdStats),
	}
	rus.merge(cmd, success, duration)
	return rus
}

func (rus *redisStats) merge(cmd string, success bool, duration time.Duration) {
	ns := duration.Nanoseconds()
	rus.totalDuration += ns
	if rcs, ok := rus.cmds[cmd]; ok {
		rcs.merge(success, ns)
	} else {
		rcs := newRedisCmdStats(cmd, success, ns)
		rus.cmds[cmd] = rcs
	}
}

func (rus *redisStats) jsonString() string {
	cmds := []string{}
	for _, cmd := range rus.cmds {
		cmds = append(cmds, cmd.jsonString())
	}
	return fmt.Sprintf(`{"url":"%v","timestamp":%v,"commands":[%v]}`, rus.url, rus.timestamp, strings.Join(cmds, ","))
}

type redisCmdStats struct {
	Command            string
	SucceedCount       int64
	FailCount          int64
	DurationSucceedSum int64 // ns
	DurationFailSum    int64 // ns
	MaxDuration        int64 // ns
	MinDuration        int64 // ns
	ResponseCount      int64
	HitCount           int64
	ResponseSizeSum    int64
	MaxResponseSize    int64
	MinResponseSize    int64
}

func newRedisCmdStats(cmd string, success bool, ns int64) *redisCmdStats {
	rcs := &redisCmdStats{
		Command:         cmd,
		ResponseCount:   -1,
		HitCount:        -1,
		ResponseSizeSum: -1,
		MaxResponseSize: -1,
		MinResponseSize: -1,
	}
	rcs.merge(success, ns)
	return rcs
}

func (rcs *redisCmdStats) merge(success bool, ns int64) {
	if success {
		rcs.SucceedCount++
		rcs.DurationSucceedSum += ns
	} else {
		rcs.FailCount++
		rcs.DurationFailSum += ns
	}
	if ns > rcs.MaxDuration {
		rcs.MaxDuration = ns
	}
	if ns < rcs.MinDuration {
		rcs.MinDuration = ns
	}
}

func (rcs *redisCmdStats) jsonString() string {
	return fmt.Sprintf(`{"command":"%v","succeedCount":%v,"failCount":%v,"durationSucceedSum":%v,"durationFailSum":%v,"maxDuration":%v,"minDuration":%v,"responseCount":%v,"hitCount":%v,"responseSizeSum":%v,"maxResponseSize":%v,"minResponseSize":%v}`,
		rcs.Command, rcs.SucceedCount, rcs.FailCount, rcs.DurationSucceedSum/1e6, rcs.DurationFailSum/1e6, rcs.MaxDuration/1e6, rcs.MinDuration/1e6, rcs.ResponseCount, rcs.HitCount, rcs.ResponseSizeSum, rcs.MaxResponseSize, rcs.MinResponseSize)
}

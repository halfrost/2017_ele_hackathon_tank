// Copyright 2016 Eleme Inc. All rights reserved.

/*

Package metric implements the in-memory metric counters for circuitbreaker.

Currently only support counters, which are implemented in RollingNumber, a
rolling number is like a sliding window on timestamp sequence.

And it is goroutine safe.

*/
package metric

import (
	"sync"
	"time"
)

/*

RollingNumber behaves like a FIFO queue with fixed length, or a sliding window
on timestamp sequence.

	1 2 0 3 [4 5 1 2 4 2] 3 4 ...   (<= time passing)
	        +--- 18  ---+

A rolling number's value is the sum of the queue items, the last item's value
will roll into previous position once the clock passed 1 granularity (default
1s).

Rolling number dosen't use thde golang timer to roll numbers on time goes on,
it uses passive clock checking instead. All read/write operations will shift
current rolling number to align its internel clock with timestamp now. The
shift will pop items on the left and fill zeros on the right, so if there is
a long time no data incoming, the rolling number will change to a all zero
queue with a sum value 0.

*/
type RollingNumber struct {
	size        int        // sliding window size
	granularity int        // time granularity in seconds
	clock       time.Time  // internel clock to align with
	values      []int      // values on the window
	lock        sync.Mutex // protects clock and values
}

// NewRollingNumber creates a new RollingNumber.
func NewRollingNumber(size, granularity int) *RollingNumber {
	return &RollingNumber{
		size:        size,
		granularity: granularity,
		clock:       time.Now(),
		values:      make([]int, size),
	}
}

// clear the rolling number to all zeros.
func (r *RollingNumber) clear() {
	r.values = make([]int, r.size)
}

// shiftOnClockChanges shifts the rolling number if its clock is behind
// the timestamp now by at least 1 timestamp granularity and synchronous
// its clock to now.
func (r *RollingNumber) shiftOnClockChanges() {
	now := time.Now()
	length := int(int(now.Sub(r.clock).Seconds()) / r.granularity)
	if length > 0 {
		r.shift(length)
		r.clock = now
	}
}

// shift the rolling number to the right by length, will pop items on the
// left and fill zeros on the right.
func (r *RollingNumber) shift(length int) {
	if length <= 0 {
		return
	}
	if length > r.size {
		r.clear()
		return
	}
	end := make([]int, length)
	r.values = append(r.values[length:], end...)
}

// Value returns the value this rolling number presents, actually the
// sum value of the queue.
func (r *RollingNumber) Value() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.shiftOnClockChanges()
	var sum int
	for i := 0; i < len(r.values); i++ {
		sum += r.values[i]
	}
	return sum
}

// Increment this number by given delta.
func (r *RollingNumber) Increment(delta int) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.shiftOnClockChanges()
	r.values[len(r.values)-1] += delta
}

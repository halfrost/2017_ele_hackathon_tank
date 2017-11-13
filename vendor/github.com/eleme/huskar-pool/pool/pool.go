/*
Origin code: https://github.com/youtube/vitess/blob/master/go/pools/resource_pool.go
Licensed under the Apache License, Version 2.0 (the "License").
*/

// Package pool provides functionality to manage and reuse resources
// like connections.
package pool

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrClosed is returned if ResourcePool is used when it's closed.
	ErrClosed = errors.New("resource pool is closed")
)

// Pooler is a abstraction for pool.
type Pooler interface {
	Get(ctx context.Context) (Resource, error)
	Put(resource Resource)
	Close()
	IsClosed() bool
	SetCapacity(capacity int) error
	SetIdleTimeout(idleTimeout time.Duration)
	StatsJSON() string
	StatsReadableJSON() string
	Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration)
}

// Factory is a function that can be used to create a resource.
type Factory func() (Resource, error)

// Resource defines the interface that every resource must provide.
// Thread synchronization between Close() and IsClosed()
// is the responsibility of the caller.
type Resource interface {
	Close() error
}

// ResourcePool allows you to use a pool of resources.
type ResourcePool struct {
	resources   chan resourceWrapper
	factory     Factory
	capacity    AtomicInt64
	idleTimeout AtomicDuration

	// stats
	waitCount AtomicInt64
	waitTime  AtomicDuration
}

type resourceWrapper struct {
	resource Resource
	timeUsed time.Time
}

// NewResourcePool creates a new ResourcePool pool.
// capacity is the number of active resources in the pool:
// there can be up to 'capacity' of these at a given time.
// maxCap specifies the extent to which the pool can be resized
// in the future through the SetCapacity function.
// You cannot resize the pool beyond maxCap.
// If a resource is unused beyond idleTimeout, it's discarded.
// An idleTimeout of 0 means that there is no timeout.
func NewResourcePool(factory Factory, capacity, maxCap int, idleTimeout time.Duration) *ResourcePool {
	if capacity <= 0 || maxCap <= 0 || capacity > maxCap {
		panic(errors.New("invalid/out of range capacity"))
	}
	rp := &ResourcePool{
		resources:   make(chan resourceWrapper, maxCap),
		factory:     factory,
		capacity:    NewAtomicInt64(int64(capacity)),
		idleTimeout: NewAtomicDuration(idleTimeout),
	}
	for i := 0; i < capacity; i++ {
		rp.resources <- resourceWrapper{}
	}
	return rp
}

// Close empties the pool calling Close on all its resources.
// You can call Close while there are outstanding resources.
// It waits for all resources to be returned (Put).
// After a Close, Get is not allowed.
func (rp *ResourcePool) Close() {
	_ = rp.SetCapacity(0)
}

// IsClosed returns true if the resource pool is closed.
func (rp *ResourcePool) IsClosed() (closed bool) {
	return rp.capacity.Get() == 0
}

// Get will return the next available resource. If capacity
// has not been reached, it will create a new one using the factory. Otherwise,
// it will wait till the next resource becomes available or a timeout.
// A timeout of 0 is an indefinite wait.
func (rp *ResourcePool) Get(ctx context.Context) (resource Resource, err error) {
	return rp.get(ctx, true)
}

func (rp *ResourcePool) get(ctx context.Context, wait bool) (resource Resource, err error) {
	// If ctx has already expired, avoid racing with rp's resource channel.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Fetch
	var wrapper resourceWrapper
	var ok bool
	select {
	case wrapper, ok = <-rp.resources:
	default:
		if !wait {
			return nil, nil
		}
		startTime := time.Now()
		select {
		case wrapper, ok = <-rp.resources:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		rp.recordWait(startTime)
	}
	if !ok {
		return nil, ErrClosed
	}

	// Unwrap
	idleTimeout := rp.idleTimeout.Get()
	if wrapper.resource != nil && idleTimeout > 0 && wrapper.timeUsed.Add(idleTimeout).Sub(time.Now()) < 0 {
		wrapper.resource.Close()
		wrapper.resource = nil
	}
	if wrapper.resource == nil {
		wrapper.resource, err = rp.factory()
		if err != nil {
			rp.resources <- resourceWrapper{}
		}
	}
	return wrapper.resource, err
}

// Put will return a resource to the pool. For every successful Get,
// a corresponding Put is required. If you no longer need a resource,
// you will need to call Put(nil) instead of returning the closed resource.
// The will eventually cause a new resource to be created in its place.
func (rp *ResourcePool) Put(resource Resource) {
	var wrapper resourceWrapper
	if resource != nil {
		wrapper = resourceWrapper{resource, time.Now()}
	}
	select {
	case rp.resources <- wrapper:
	default:
		panic(errors.New("attempt to Put into a full ResourcePool"))
	}
}

// SetCapacity changes the capacity of the pool.
// You can use it to shrink or expand, but not beyond
// the max capacity. If the change requires the pool
// to be shrunk, SetCapacity waits till the necessary
// number of resources are returned to the pool.
// A SetCapacity of 0 is equivalent to closing the ResourcePool.
func (rp *ResourcePool) SetCapacity(capacity int) error {
	if capacity < 0 || capacity > cap(rp.resources) {
		return fmt.Errorf("capacity %d is out of range", capacity)
	}

	// Atomically swap new capacity with old, but only
	// if old capacity is non-zero.
	var oldcap int
	for {
		oldcap = int(rp.capacity.Get())
		if oldcap == 0 {
			return ErrClosed
		}
		if oldcap == capacity {
			return nil
		}
		if rp.capacity.CompareAndSwap(int64(oldcap), int64(capacity)) {
			break
		}
	}

	if capacity < oldcap {
		for i := 0; i < oldcap-capacity; i++ {
			wrapper := <-rp.resources
			if wrapper.resource != nil {
				wrapper.resource.Close()
			}
		}
	} else {
		for i := 0; i < capacity-oldcap; i++ {
			rp.resources <- resourceWrapper{}
		}
	}
	if capacity == 0 {
		close(rp.resources)
	}
	return nil
}

func (rp *ResourcePool) recordWait(start time.Time) {
	rp.waitCount.Add(1)
	rp.waitTime.Add(time.Now().Sub(start))
}

// SetIdleTimeout sets the idle timeout.
func (rp *ResourcePool) SetIdleTimeout(idleTimeout time.Duration) {
	rp.idleTimeout.Set(idleTimeout)
}

// StatsJSON returns the stats in JSON format.
func (rp *ResourcePool) StatsJSON() string {
	c, a, mx, wc, wt, it := rp.Stats()
	return fmt.Sprintf(`{"Capacity": %v, "Available": %v, "MaxCapacity": %v, "WaitCount": %v, "WaitTime": %v, "IdleTimeout": %v}`, c, a, mx, wc, int64(wt), int64(it))
}

// StatsReadableJSON returns the stats in JSON format, but convert time.Duration into human readable string.
func (rp *ResourcePool) StatsReadableJSON() string {
	c, a, mx, wc, wt, it := rp.Stats()
	return fmt.Sprintf(`{"Capacity": %v, "Available": %v, "MaxCapacity": %v, "WaitCount": %v, "WaitTime": "%v", "IdleTimeout": "%v"}`, c, a, mx, wc, wt, it)
}

// Stats returns the stats.
func (rp *ResourcePool) Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration) {
	return rp.Capacity(), rp.Available(), rp.MaxCap(), rp.WaitCount(), rp.WaitTime(), rp.IdleTimeout()
}

// Capacity returns the capacity.
func (rp *ResourcePool) Capacity() int64 {
	return rp.capacity.Get()
}

// Available returns the number of currently unused resources.
func (rp *ResourcePool) Available() int64 {
	return int64(len(rp.resources))
}

// MaxCap returns the max capacity.
func (rp *ResourcePool) MaxCap() int64 {
	return int64(cap(rp.resources))
}

// WaitCount returns the total number of waits.
func (rp *ResourcePool) WaitCount() int64 {
	return rp.waitCount.Get()
}

// WaitTime returns the total wait time.
func (rp *ResourcePool) WaitTime() time.Duration {
	return rp.waitTime.Get()
}

// IdleTimeout returns the idle timeout.
func (rp *ResourcePool) IdleTimeout() time.Duration {
	return rp.idleTimeout.Get()
}

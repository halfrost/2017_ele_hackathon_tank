// Copyright 2016 Eleme Inc. All rights reserved.

package internal

import (
	"sync"
)

// Registry store
type Registry struct {
	sync.RWMutex
	topics    map[string]map[chan interface{}]bool
	capacity  int
	genTopics map[chan interface{}]bool
}

// NewRegistry create a registry.
func NewRegistry(capacity int) *Registry {
	return &Registry{
		topics:    make(map[string]map[chan interface{}]bool),
		capacity:  capacity,
		genTopics: make(map[chan interface{}]bool),
	}
}

// Pub publish the msg on topic.
func (reg *Registry) Pub(msg interface{}, topic string) {
	reg.send(topic, msg)
}

// Sub returns a channel subscribed to topic.
func (reg *Registry) Sub(topic string) chan interface{} {
	msgC := make(chan interface{}, reg.capacity)
	reg.add(topic, msgC)
	return msgC
}

// SubAll returns a channel subscribed to all topic.
func (reg *Registry) SubAll() chan interface{} {
	msgC := make(chan interface{}, reg.capacity)
	reg.addAll(msgC)
	return msgC
}

// Unsub unsubscribes the ch on topic.
func (reg *Registry) Unsub(topic string, ch chan interface{}) {
	reg.remove(topic, ch)
}

// UnsubAll unsubscribes the ch on all topic.
func (reg *Registry) UnsubAll(ch chan interface{}) {
	reg.Lock()
	defer reg.Unlock()
	delete(reg.genTopics, ch)
	close(ch)
}

// Clean clean all topics.
func (reg *Registry) Clean() {
	for topic, chans := range reg.topics {
		for ch := range chans {
			reg.remove(topic, ch)
		}
	}
}

// for sub
func (reg *Registry) add(topic string, ch chan interface{}) {
	reg.Lock()
	defer reg.Unlock()
	chans, ok := reg.topics[topic]
	if !ok {
		chans = make(map[chan interface{}]bool)
	}
	reg.topics[topic] = chans
	chans[ch] = true
}

// for subAll
func (reg *Registry) addAll(ch chan interface{}) {
	reg.Lock()
	defer reg.Unlock()
	reg.genTopics[ch] = true
}

// for pub
func (reg *Registry) send(topic string, msg interface{}) {
	reg.RLock()
	chans := []chan interface{}{}
	for ch := range reg.topics[topic] {
		chans = append(chans, ch)
	}
	for ch := range reg.genTopics {
		chans = append(chans, ch)
	}
	reg.RUnlock()

	// we make a copy of channels to avoid dead lock
	for _, ch := range chans {
		ch <- msg
	}
}

func (reg *Registry) remove(topic string, ch chan interface{}) {
	reg.Lock()
	defer reg.Unlock()
	if _, ok := reg.topics[topic]; !ok {
		return
	}
	if _, ok := reg.topics[topic][ch]; !ok {
		return
	}

	delete(reg.topics[topic], ch)
	if len(reg.topics[topic]) == 0 {
		delete(reg.topics, topic)
	}
	close(ch)
}

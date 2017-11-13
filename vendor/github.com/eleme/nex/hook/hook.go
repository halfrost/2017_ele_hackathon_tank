package hook

import "sync"

// Observer defines a standard interface for instances that wish to listen for
// the occurrence of a specific event.
type Observer interface {
	// OnNotify allows an event to be "published" to interface implementations.
	OnNotify(evt interface{})
}

// Notifier is used to notify observers.
type Notifier struct {
	sync.RWMutex
	name      string
	observers map[Observer]struct{}
}

// NewNotifier creates a new Notifier.
func NewNotifier(name string) *Notifier {
	return &Notifier{
		name:      name,
		observers: map[Observer]struct{}{},
	}
}

// Name returns the identify of the Notifier.
func (n *Notifier) Name() string {
	return n.name
}

// Register allows an Observer to register itself to observe events.
func (n *Notifier) Register(o Observer) {
	n.Lock()
	n.observers[o] = struct{}{}
	n.Unlock()
}

// Deregister allows an Observer to remove itself from the collection of observers.
func (n *Notifier) Deregister(o Observer) {
	n.Lock()
	delete(n.observers, o)
	n.Unlock()
}

// Notify publishes new events to observers.
func (n *Notifier) Notify(evt interface{}) {
	n.RLock()
	for o := range n.observers {
		o.OnNotify(evt)
	}
	n.RUnlock()
}

// Copyright 2016 Eleme Inc. All rights reserved.

package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/eleme/huskar/cache"
	"github.com/eleme/huskar/internal"
	"github.com/eleme/huskar/structs"
)

// Registrator used to get, watch and register service.
type Registrator interface {
	// EnableCache enable disk cache.
	EnableCache(fpath string) error
	// Register is used to add or update one service’s static data or runtime data.
	Register(service string, cluster string, instance *Instance) error
	// Deregister is used to deregister a service with service, cluster and key.
	Deregister(service string, cluster string, key string) error
	// Get is used to get single instance of service by specific service, cluster and key.
	Get(service string, cluster string, key string) (*Instance, error)
	// GetAll is used to get all instances of service by specific service and cluster.
	GetAll(service string, cluster string) ([]*Instance, error)
	// GetClusters is used to get all clusters of specified service.
	GetClusters(service string) ([]*structs.Cluster, error)
	// Watch specified service and clusters until http long_pool breaked or call stopWatch.
	Watch(service string, clusters ...string) (eventC <-chan *Event, stopWatch func(), err error)
	// SetLogger sets the logger to be used for printing errors.
	SetLogger(l structs.Logger)
	// DeleteCluster deletes an empty cluster.
	DeleteCluster(service string, cluster string) error
}

// Event represents the event of service.
type Event struct {
	Type      structs.EventType
	Err       error
	Instances []*Instance
}

// Register implements the Registrator interface.
type Register struct {
	logger        structs.Logger
	client        *internal.Client
	rwmu          sync.RWMutex
	store         map[string]map[string]*Instance // key: appid/clusters
	wmu           sync.Mutex
	watchedTopics map[string]chan struct{}
	cache         *cache.Cache
	registry      *internal.Registry
}

// New create a Register with config.
func New(config structs.Config) (Registrator, error) {
	client, err := internal.NewClient(config)
	if err != nil {
		return nil, err
	}
	return NewWithClient(client), nil
}

// NewWithClient create a Configer with client.
func NewWithClient(client *internal.Client) *Register {
	return &Register{
		logger:        structs.DefaultLogger{},
		client:        client,
		store:         map[string]map[string]*Instance{},
		watchedTopics: map[string]chan struct{}{},
		registry:      internal.NewRegistry(5),
	}
}

// SetLogger sets the logger to be used for printing errors.
func (r *Register) SetLogger(l structs.Logger) {
	r.logger = l
}

type options struct {
	service  string
	cluster  string
	key      string
	formData map[string]string
}

// GetClusters is used to get all clusters of specified service
func (r *Register) GetClusters(service string) ([]*structs.Cluster, error) {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	opts := options{service: service}
	content, err := r.doMethod(http.MethodGet, opts)
	if err != nil {
		return nil, err
	}
	resultClusters, err := structs.NewResultClusters(content)
	if err != nil {
		return nil, err
	}
	return resultClusters.Data, nil
}

// Register is used to add or update one service’s static data or runtime data.
func (r *Register) Register(service, cluster string, instance *Instance) error {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	if cluster == "" {
		cluster = config.Cluster
	}

	if instance == nil {
		return nil
	}

	if err := instance.validate(); err != nil {
		return fmt.Errorf("instance validate failed: %s", err)
	}
	formData, err := instance.toMap()
	if err != nil {
		return fmt.Errorf("instance to form data failed: %s", err)
	}

	opts := options{
		service:  service,
		cluster:  cluster,
		formData: formData,
	}
	if _, err := r.doMethod(http.MethodPost, opts); err != nil {
		return err
	}
	return nil
}

// Deregister is used to deregister a service with service, cluster and key.
func (r *Register) Deregister(service string, cluster string, key string) error {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	if cluster == "" {
		cluster = config.Cluster
	}
	opts := options{
		service: service,
		cluster: cluster,
		key:     key,
	}
	if _, err := r.doMethod(http.MethodDelete, opts); err != nil {
		return err
	}
	return nil
}

// DeleteCluster deletes an empty cluster.
func (r *Register) DeleteCluster(service string, cluster string) error {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	if cluster == "" {
		cluster = config.Cluster
	}
	opts := options{
		service: service,
		formData: map[string]string{
			"cluster": cluster,
		},
	}
	if _, err := r.doMethod(http.MethodDelete, opts); err != nil {
		return err
	}
	return nil
}

func (r *Register) doMethod(method string, opts options) ([]byte, error) {
	token := r.client.Config().Token
	url := path.Join("/api/service", opts.service, opts.cluster)
	if opts.key != "" {
		url = fmt.Sprintf("%s?key=%s", url, opts.key)
	}

	doOptions := internal.DoOptions{
		Headers: map[string]string{internal.HeaderAuth: token},
	}
	if opts.formData != nil {
		doOptions.FormData = opts.formData
	}

	resp, err := r.client.Do(method, url, doOptions)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// EnableCache enable disk cache, call it ASAP.
func (r *Register) EnableCache(cachePath string) error {
	cache, err := cache.New(cachePath)
	if err != nil {
		return err
	}

	r.rwmu.Lock()
	defer r.rwmu.Unlock()
	if len(r.store) > 0 {
		err = cache.Save(&r.store)
	} else {
		err = cache.Load(&r.store)
	}
	if err != nil {
		return err
	}
	r.cache = cache
	return nil
}

// Get is used to get single instance of service by specific service, cluster and key.
func (r *Register) Get(service string, cluster string, key string) (*Instance, error) {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	if cluster == "" {
		cluster = config.Cluster
	}
	if key == "" {
		return nil, structs.ErrKeyEmpty
	}

	topic := path.Join(service, cluster)
	connectedCh := r.watch(topic, service, cluster)
	// XXX: wait first?
	err := waitCh(connectedCh, config.WaitTimeout)
	ins := r.getFromCache(topic, key)
	if ins != nil {
		return ins, nil
	}
	if err == nil {
		err = structs.ErrKeyNotExists
	}
	return nil, err
}

// GetAll is used to get all instances of service by specific service and cluster.
func (r *Register) GetAll(service string, cluster string) ([]*Instance, error) {
	config := r.client.Config()
	if service == "" {
		service = config.Service
	}
	if cluster == "" {
		cluster = config.Cluster
	}

	topic := path.Join(service, cluster)
	connectedCh := r.watch(topic, service, cluster)
	// XXX: wait first?
	err := waitCh(connectedCh, config.WaitTimeout)
	instances := r.getAllFromCache(topic)
	if instances != nil && len(instances) > 0 {
		return instances, nil
	}
	if err == nil {
		err = structs.ErrKeyNotExists
	}
	return nil, err
}

func (r *Register) getFromCache(topic, key string) *Instance {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()
	if instances, in := r.store[topic]; in {
		if ins, in := instances[key]; in {
			return ins
		}
	}
	return nil
}

func (r *Register) getAllFromCache(topic string) []*Instance {
	r.rwmu.RLock()
	defer r.rwmu.RUnlock()
	instances := []*Instance{}
	for _, ins := range r.store[topic] {
		instances = append(instances, ins)
	}
	return instances
}

func (r *Register) cacheInstances(topic string, event *Event) {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	switch event.Type {
	case structs.EventAll:
		instances := map[string]*Instance{}
		for _, ins := range event.Instances {
			instances[ins.Key] = ins
		}
		r.store[topic] = instances
	case structs.EventUpdate:
		instances, in := r.store[topic]
		if !in {
			instances = map[string]*Instance{}
			r.store[topic] = instances
		}
		for _, ins := range event.Instances {
			instances[ins.Key] = ins
		}
	case structs.EventDelete:
		if instances, in := r.store[topic]; in {
			for _, ins := range event.Instances {
				delete(instances, ins.Key)
			}
		}
	}

	if r.cache != nil {
		r.cache.Save(&r.store)
	}
}

// Watch specified service and clusters until http long_pool breaked.
func (r *Register) Watch(service string, clusters ...string) (<-chan *Event, func(), error) {
	topic := path.Join(service, path.Join(clusters...))

	eventC := make(chan *Event)
	msgC := r.registry.Sub(topic)
	go func() {
		if instances := r.getAllFromCache(topic); len(instances) > 0 {
			event := &Event{Instances: instances, Type: structs.EventAll}
			eventC <- event // notify if it exists in cache
		}
		for msg := range msgC {
			eventC <- msg.(*Event)
		}
		close(eventC)
	}()

	r.watch(topic, service, clusters...)

	return eventC, func() {
		r.registry.Unsub(topic, msgC)
	}, nil
}

func (r *Register) watch(topic, service string, clusters ...string) <-chan struct{} {
	r.wmu.Lock()
	ch, in := r.watchedTopics[topic]
	if !in {
		services := internal.NewServices()
		services.AddClusters(service, clusters...)
		ch = make(chan struct{})
		r.watchedTopics[topic] = ch
		go r.loop(topic, services, ch)
	}
	r.wmu.Unlock()
	return ch
}

func (r *Register) loop(topic string, services internal.Services, connectedCh chan struct{}) {
	entryC := make(chan *structs.Entry)
	done := make(chan bool)
	watcher := internal.NewWatcher(r.client, structs.WatchTypeService)
	errCh := watcher.Watch(services, internal.WatchOptions{EntryC: entryC, Done: done})
Loop:
	for {
		select {
		case entry, ok := <-entryC:
			if !ok {
				break Loop
			}

			event := &Event{Instances: make([]*Instance, 0)}

			// service link changed WILL trigger all keys pull event.
			switch entry.Message {
			case structs.MessageTypePing:
				continue
			case structs.MessageTypeAll:
				event.Type = structs.EventAll
			case structs.MessageTypeUpdate:
				event.Type = structs.EventUpdate
			case structs.MessageTypeDelete:
				event.Type = structs.EventDelete
			default:
				event.Type = structs.EventUnknown
			}

			nodesMap := entry.NodesMap()
			for _, node := range nodesMap {
				instance, err := nodeToInstance(node)
				event.Instances = append(event.Instances, instance)
				if err != nil {
					event.Type = structs.EventError
					event.Err = err
					break
				}
			}
			r.cacheInstances(topic, event)
			select {
			case <-connectedCh:
			default:
				close(connectedCh)
			}
			r.registry.Pub(event, topic)
		}
	}

	err := <-errCh
	if err != nil {
		r.logger.Printf("%s, will retry after: %v", err, r.client.Config().RetryDelay)
		time.Sleep(r.client.Config().RetryDelay)
		go r.loop(topic, services, connectedCh)
	}
	close(done)
}

func nodeToInstance(node *structs.Node) (*Instance, error) {
	instance := &Instance{
		Key:   node.Key,
		Value: &StaticInfo{},
	}
	if node.Value != "" {
		if err := json.Unmarshal([]byte(node.Value), instance.Value); err != nil {
			return instance, fmt.Errorf("unmarshal static value failed: %s", err)
		}
	}
	if node.Runtime != "" {
		if err := json.Unmarshal([]byte(node.Runtime), &instance.Runtime); err != nil {
			return instance, fmt.Errorf("unmarshal runtime value failed: %s", err)
		}
	}
	return instance, nil
}

func waitCh(ch <-chan struct{}, d time.Duration) error {
	select {
	case <-ch:
	default:
		select {
		case <-ch:
		case <-time.After(d):
			return structs.ErrWaitTimeout
		}
	}
	return nil
}

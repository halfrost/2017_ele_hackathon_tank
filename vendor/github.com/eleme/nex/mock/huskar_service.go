// TODO: Actually implement this

package mock

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/eleme/huskar/service"
	"github.com/eleme/huskar/structs"
	json "github.com/json-iterator/go"
)

type fileInstance struct {
	Key     string         `json:"key"`
	App     string         `json:"application"`
	Cluster string         `json:"cluster"`
	State   string         `json:"state"`
	IP      string         `json:"ip"`
	Port    map[string]int `json:"port"`
}

// HuskarFileRegistrator is a mock.
type HuskarFileRegistrator struct {
	mu        *sync.RWMutex
	fileName  string
	instances map[string]map[string]*service.Instance
}

// NewHuskarFileReisgrator creats a file register.
// The file(service.json) must place in top level directory(same as app.yaml),
// it will check service.json every 3 seconds, so the file changing can be detected.
// the file content format:
//   [
//     {
//       "key": "111.111.111.111_9999",
//       "application": "arch.note",
//       "cluster": "Common",
//       "state": "up",
//       "ip": "111.111.111.111",
//       "port": {"main": 8010, "back": 8011}
//     },
//     ...
//   ]
func NewHuskarFileReisgrator(fileName string) *HuskarFileRegistrator {
	return &HuskarFileRegistrator{
		mu:        &sync.RWMutex{},
		fileName:  fileName,
		instances: make(map[string]map[string]*service.Instance),
	}
}

func topic(service string, clusters ...string) string {
	return fmt.Sprintf("%s@%s", service, strings.Join(clusters, ""))
}

// EnableCache mock.
func (register *HuskarFileRegistrator) EnableCache(_ string) error { return nil }

// Register is used to add or update one service’s static data or runtime data.
func (register *HuskarFileRegistrator) Register(service string, cluster string, instance *service.Instance) error {
	return nil
}

// Deregister is used to deregister a service with service, cluster and key.
func (register *HuskarFileRegistrator) Deregister(service string, cluster string, key string) error {
	return nil
}

// Get is used to get single instance of service by specific service, cluster and key.
func (register *HuskarFileRegistrator) Get(service string, cluster string, key string) (*service.Instance, error) {
	if key == "" {
		return nil, structs.ErrKeyEmpty
	}
	register.mu.RLock()
	defer register.mu.RUnlock()
	if instances, in := register.instances[topic(service, cluster)]; in {
		if ins, in := instances[key]; in {
			return ins, nil
		}
	}
	return nil, fmt.Errorf("No service instance found for '%v'", key)
}

// GetAll is used to get all instances of service by specific service and cluster.
func (register *HuskarFileRegistrator) GetAll(serviceName string, cluster string) ([]*service.Instance, error) {
	register.mu.RLock()
	defer register.mu.RUnlock()
	if instances, in := register.instances[topic(serviceName, cluster)]; in {
		lins := []*service.Instance{}
		for _, ins := range instances {
			lins = append(lins, ins)
		}
		return lins, nil
	}
	return nil, fmt.Errorf("No service instances found for '%v@%v'", serviceName, cluster)
}

// GetClusters is used to get all clusters of specified service.
func (register *HuskarFileRegistrator) GetClusters(service string) ([]*structs.Cluster, error) {
	return nil, errNotSupported
}

// Watch specified service and clusters until http long_pool breaked or call stopWatch.
func (register *HuskarFileRegistrator) Watch(serviceName string, clusters ...string) (<-chan *service.Event, func(), error) {
	t := topic(serviceName, clusters...)
	ticker := time.Tick(3 * time.Second)
	eventC := make(chan *service.Event)
	stop := make(chan struct{})

	go func() {
		register.checkAndNotify(t, eventC)
	LOOP:
		for {
			select {
			case <-stop:
				break LOOP
			case <-ticker:
				register.checkAndNotify(t, eventC)
			}
		}
		close(eventC)
	}()
	return eventC, func() { close(stop) }, nil
}

func (register *HuskarFileRegistrator) checkAndNotify(topicName string, eventC chan<- *service.Event) {
	data, err := ioutil.ReadFile(register.fileName)
	if err != nil {
		return // ignore
	}
	var contents []*fileInstance
	if err := json.Unmarshal(data, &contents); err != nil {
		return // ignore
	}

	register.mu.Lock()
	defer register.mu.Unlock()

	added := []*service.Instance{}
	deleted := []*service.Instance{}
	instances := map[string]*service.Instance{}
	oldInstances, in := register.instances[topicName]
	if !in {
		oldInstances = map[string]*service.Instance{}
	}

	for _, content := range contents {
		if topicName != topic(content.App, content.Cluster) {
			continue
		}
		instance := &service.Instance{
			Key: content.Key,
			Value: &service.StaticInfo{
				IP:    content.IP,
				Port:  content.Port,
				State: service.State(content.State),
			},
		}
		instances[content.Key] = instance
		old, in := oldInstances[content.Key]
		if !in || old.Value.State != instance.Value.State {
			added = append(added, instance)
		}
	}

	register.instances[topicName] = instances

	for key, instance := range oldInstances {
		if _, in := instances[key]; !in {
			deleted = append(deleted, instance)
		}
	}

	if len(added) > 0 {
		event := &service.Event{Type: structs.EventUpdate, Instances: added}
		eventC <- event
	}
	if len(deleted) > 0 {
		event := &service.Event{Type: structs.EventDelete, Instances: deleted}
		eventC <- event
	}
}

// SetLogger sets the logger to be used for printing errors.
func (register *HuskarFileRegistrator) SetLogger(_ structs.Logger) {}

// DeleteCluster deletes an empty cluster.
func (register *HuskarFileRegistrator) DeleteCluster(service string, cluster string) error {
	return errNotSupported
}

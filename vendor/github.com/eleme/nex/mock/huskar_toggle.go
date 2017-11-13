package mock

import (
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/eleme/huskar/structs"
	json "github.com/json-iterator/go"
)

// HuskarFileToggler  is a mock
type HuskarFileToggler struct {
	content map[string]float64
}

// NewHuskarFileToggler creats a file toggler,
// The file(toggle.json) must place in top level directory(same as app.yaml),
// the file content format:
//   {
//     "key": value(number),
//     ...
//   }
func NewHuskarFileToggler(filename string) (*HuskarFileToggler, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %s", filename, err)
	}

	var data map[string]float64
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	return NewHuskarFileTogglerWith(data), nil
}

// NewHuskarFileTogglerWith creates a new HuskarFileToggler with a map of data.
func NewHuskarFileTogglerWith(data map[string]float64) *HuskarFileToggler {
	return &HuskarFileToggler{
		content: data,
	}
}

// EnableCache mock.
func (tg *HuskarFileToggler) EnableCache(_ string) error { return nil }

// IsOn get the current state of toggle by key name. True is ON, false is OFF.
func (tg *HuskarFileToggler) IsOn(key string) (bool, error) {
	v, in := tg.content[key]
	if !in {
		return false, fmt.Errorf("key %s not exist", key)
	}
	if structs.IsEqual(100.0, v) {
		return true, nil
	} else if structs.IsEqual(v, 0.0) {
		return false, nil
	}
	x := rand.Float64() * 100.0
	return x < v, nil
}

// Rate get the current rate of toggle by key name.
func (tg *HuskarFileToggler) Rate(key string) (float32, error) {
	v, in := tg.content[key]
	if !in {
		return 0.0, fmt.Errorf("key %s not exist", key)
	}
	return float32(v), nil
}

// Watch used to watch specified switch by key, and return the stop watch function.
func (tg *HuskarFileToggler) Watch(key string) (eventC <-chan *structs.Event, stopWatch func(), err error) {
	return nil, nil, errNotSupported
}

// WatchAll the value change of all key, and return the stop watch function.
func (tg *HuskarFileToggler) WatchAll() (nodeC <-chan *structs.Event, stopWatch func(), err error) {
	return nil, nil, errNotSupported
}

// GetAll return all key-value pair.
func (tg *HuskarFileToggler) GetAll() (map[string]string, error) {
	all := make(map[string]string)
	for k, v := range tg.content {
		all[k] = fmt.Sprintf("%v", v)
	}
	return all, nil
}

// IsOnOr get the current state of toggle by key name, and will return toggleDefault
// either if the desired toggle doesn't exists or there's an error getting it.
func (tg *HuskarFileToggler) IsOnOr(key string, toggleDefault bool) bool {
	isOn, err := tg.IsOn(key)
	if err != nil {
		return toggleDefault
	}
	return isOn
}

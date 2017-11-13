package mock

import (
	"fmt"
	"io/ioutil"

	huskarConfig "github.com/eleme/huskar/config"
	"github.com/eleme/huskar/structs"
	json "github.com/json-iterator/go"
)

// HuskarFileConfiger implements github.com/eleme/huskar/config/Configer,
// it can be used when local file is desired instead of huskar
type HuskarFileConfiger struct {
	content map[string]string
	logger  structs.Logger
	codec   huskarConfig.Codecer
}

// NewHuskarFileConfiger creates a file configer,
// The file(config.json) must place in top level directory(same as app.yaml),
// the file content format:
//   {
//     "key": "value",
//     ...
//   }
func NewHuskarFileConfiger(filename string) (*HuskarFileConfiger, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %s", filename, err)
	}

	var data map[string]string
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	return NewHuskarFileConfigerWith(data), nil
}

// NewHuskarFileConfigerWith creates a new HuskarFileConfiger with a map of data.
func NewHuskarFileConfigerWith(data map[string]string) *HuskarFileConfiger {
	return &HuskarFileConfiger{
		logger:  structs.DefaultLogger{},
		codec:   new(huskarConfig.Codec),
		content: data,
	}
}

// EnableCache mock.
func (fileConfiger *HuskarFileConfiger) EnableCache(_ string) error { return nil }

// Get config value by key.
func (fileConfiger *HuskarFileConfiger) Get(key string) (string, error) {
	if result, in := fileConfiger.content[key]; in {
		return result, nil
	}
	return "", fmt.Errorf("config: %s not in file config", key)
}

// GetAll return all key-value pair.
func (fileConfiger *HuskarFileConfiger) GetAll() (map[string]string, error) {
	return fileConfiger.content, nil
}

// UnmarshalAll is used to unmarshal all config elements.
func (fileConfiger *HuskarFileConfiger) UnmarshalAll(i interface{}) error {
	return fileConfiger.codec.Unmarshal(fileConfiger.content, i)
}

// Watch the value change of specified key, and return the stop watch function.
func (fileConfiger *HuskarFileConfiger) Watch(key string) (nodeC <-chan *structs.Event, stopWatch func(), err error) {
	return nil, nil, errNotSupported
}

// WatchAll the value change of all key, and return the stop watch function.
func (fileConfiger *HuskarFileConfiger) WatchAll() (<-chan *structs.Event, func(), error) {
	return nil, nil, errNotSupported
}

// Set used to add a new config.
func (fileConfiger *HuskarFileConfiger) Set(key string, value []byte, comment []byte) error {
	fileConfiger.content[key] = string(value)
	return nil
}

// Update used to update a config. If comment given nil, the comment of the key will not be updated.
func (fileConfiger *HuskarFileConfiger) Update(key string, value []byte, comment []byte) error {
	return errNotSupported
}

// Delete used to delete a config with key.
func (fileConfiger *HuskarFileConfiger) Delete(key string) error {
	return errNotSupported
}

// SetLogger sets the logger to be used for printing errors.
func (fileConfiger *HuskarFileConfiger) SetLogger(l structs.Logger) {
	fileConfiger.logger = l
}

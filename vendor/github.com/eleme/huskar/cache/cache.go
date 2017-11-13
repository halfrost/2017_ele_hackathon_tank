package cache

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Cache saves JSON data on disk.
type Cache struct {
	fpath string
}

// New creates a new Cache.
func New(fpath string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &Cache{fpath: fpath}, nil
}

// Load loads data from the disk.
func (c *Cache) Load(v interface{}) error {
	data, err := ioutil.ReadFile(c.fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, v)
}

// Save saves data to disk.
func (c *Cache) Save(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.fpath, data, 0644)
}

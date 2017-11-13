// Copyright 2016 Eleme Inc. All rights reserved.

package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eleme/huskar/structs"
)

// Services is a map of name to clusters.
type Services map[string][]string

// NewServices create a Services.
func NewServices() Services {
	return make(Services)
}

// AddClusters add clusters in specific service.
func (ss Services) AddClusters(service string, clusters ...string) {
	if _, ok := ss[service]; !ok && len(clusters) == 0 {
		ss[service] = []string{}
	}
	ss[service] = append(ss[service], clusters...)
}

// Watcher used to watch specific services.
type Watcher struct {
	c         *Client
	watchType structs.WatchType
}

// NewWatcher create a Watcher with client and watchType.
func NewWatcher(client *Client, watchType structs.WatchType) *Watcher {
	return &Watcher{
		c:         client,
		watchType: watchType,
	}
}

// Watch sends entry for the given services to the entry channel.
func (w *Watcher) Watch(services Services, opts WatchOptions) chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- w.watch(services, opts)
		close(errCh)
	}()
	return errCh
}

// WatchOptions wraps watch options.
type WatchOptions struct {
	EntryC chan<- *structs.Entry
	Done   chan bool
}

// watch can be stoped by opts.Done.
func (w *Watcher) watch(services Services, opts WatchOptions) (retErr error) {
	errCh := make(chan error, 1)
	readCloser, writeCloser := io.Pipe()
	defer func() {
		close(opts.EntryC)

		select {
		case err := <-errCh:
			if err != nil && retErr == nil {
				retErr = err
			}
		default:
			// No errors
		}
		if err := readCloser.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	out, err := json.Marshal(services)
	if err != nil {
		return fmt.Errorf("marshal services failed: %s", err)
	}
	postData := fmt.Sprintf(`{"%s":%s}`, w.watchType, string(out))
	streamOptions := StreamOptions{
		Headers: map[string]string{HeaderAuth: w.c.config.Token},
		In:      bytes.NewBufferString(postData),
		Stdout:  writeCloser,
	}

	go func() {
		err := w.c.Stream(http.MethodPost, "/api/data/long_poll", streamOptions)
		if closeErr := writeCloser.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		errCh <- err
		close(errCh)
	}()

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		// block here waiting for the signal to stop function
		select {
		case <-opts.Done:
			readCloser.Close()
		case <-quit:
			return
		}
	}()

	decoder := json.NewDecoder(readCloser)
	entry := new(structs.Entry)
	for err := decoder.Decode(entry); err != io.EOF; err = decoder.Decode(entry) {
		if err != nil {
			return err
		}
		if entry.Message == "" {
			continue
		}
		entry.WatchType = w.watchType

		opts.EntryC <- entry
		entry = new(structs.Entry)
	}
	return
}

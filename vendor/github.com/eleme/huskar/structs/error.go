// Copyright 2016 Eleme Inc. All rights reserved.

package structs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Error is the error type usually returned by functions.
// It describes the HTTP status code, status, and message of an error.
type Error struct {
	StatusCode int         `json:"-"`
	Status     string      `json:"status,omitempty"`
	Message    interface{} `json:"message,omitempty"`
}

// NewError create an Error.
func NewError(resp *http.Response) *Error {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Error{StatusCode: resp.StatusCode, Message: fmt.Sprintf("cannot read body, err: %v", err)}
	}
	e := &Error{}
	if err := json.Unmarshal(data, e); err != nil {
		return &Error{StatusCode: resp.StatusCode, Message: fmt.Sprintf("cannot unmarshal body, err: %v", err)}
	}
	e.StatusCode = resp.StatusCode
	return e
}

// Error returns an error string formatted as follows:
// API error ($StatusCode): Status: $Status, Message: $message
func (e *Error) Error() string {
	var message string
	switch msg := e.Message.(type) {
	case string:
		message = msg
	case map[string]interface{}:
		for k, v := range msg {
			message += fmt.Sprintf("%s: %s", k, v)
		}
	default:
		message = fmt.Sprintf("%s", msg)
	}

	return fmt.Sprintf("API error (%d): Status: %s, Message: %s", e.StatusCode, e.Status, message)
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	codecLog     = false
	durationType = reflect.TypeOf(time.Duration(0))
)

func init() {
	e := os.Getenv("GODEBUG")
	if strings.Contains(e, "huskardebug=1") {
		codecLog = true
	}
}

// Codecer is used to marshal and unmarshal config elements.
type Codecer interface {
	// Unmarshal parses the given values and store the result
	// in the value pointed to i.
	Unmarshal(values map[string]string, i interface{}) error
}

// Codec implements the Codecer interface
type Codec struct{}

// Unmarshal parses the given values and store the result
// in the value pointed to i.
func (c *Codec) Unmarshal(values map[string]string, i interface{}) error {
	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return errors.New("must provide a ptr value")
	}
	v := reflect.Indirect(reflect.ValueOf(i))
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("must provide a struct ptr")
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("huskar")
		if tag == "" {
			continue
		}

		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}

		value, ok := values[tag]
		if !ok {
			continue
		}

		var (
			err  error
			kind = fv.Kind()
		)
		switch kind {
		case reflect.Bool:
			err = c.unmarshalBool(fv, value)
		case reflect.String:
			err = c.unmarshalString(fv, value)

		case reflect.Int:
			err = c.unmarshalInt(fv, value, 0)
		case reflect.Int8:
			err = c.unmarshalInt(fv, value, 8)
		case reflect.Int16:
			err = c.unmarshalInt(fv, value, 16)
		case reflect.Int32:
			err = c.unmarshalInt(fv, value, 32)
		case reflect.Int64:
			if fv.Type() == durationType {
				err = c.unmarshalDuration(fv, value)
			} else {
				err = c.unmarshalInt(fv, value, 64)
			}

		case reflect.Float32:
			err = c.unmarshalFloat(fv, value, 32)
		case reflect.Float64:
			err = c.unmarshalFloat(fv, value, 64)

		case reflect.Struct:
			err = c.unmarshalStruct(fv, value)
		case reflect.Slice:
			err = c.unmarshalSlice(fv, value)
		case reflect.Ptr:
			err = c.unmarshalPtr(fv, value)
		default:
			err = fmt.Errorf("unsupport filed type: %s", kind)
		}
		if codecLog && err != nil {
			log.Printf("unmarshal value '%s' to the field pointed to %s tag failed: %s\n", value, tag, err)
		}
	}
	return nil
}

func (c *Codec) unmarshalBool(fv reflect.Value, hv string) error {
	if hv == "1" || hv == "true" || hv == "True" {
		fv.SetBool(true)
	} else {
		fv.SetBool(false)
	}
	return nil
}

func (c *Codec) unmarshalString(fv reflect.Value, hv string) error {
	fv.SetString(hv)
	return nil
}

func (c *Codec) unmarshalInt(fv reflect.Value, hv string, bitSize int) error {
	v, err := strconv.ParseInt(hv, 10, bitSize)
	if err != nil {
		return err
	}
	fv.SetInt(v)
	return nil
}

func (c *Codec) unmarshalDuration(fv reflect.Value, hv string) error {
	d, err := time.ParseDuration(hv)
	if err != nil {
		return err
	}
	fv.SetInt(int64(d))
	return nil
}

func (c *Codec) unmarshalFloat(fv reflect.Value, hv string, bitSize int) error {
	v, err := strconv.ParseFloat(hv, bitSize)
	if err != nil {
		return err
	}
	fv.SetFloat(v)
	return nil
}

func (c *Codec) unmarshalStruct(fv reflect.Value, hv string) error {
	ft := fv.Type()
	e := reflect.New(ft)
	ei := e.Interface()
	if err := json.Unmarshal([]byte(hv), &ei); err != nil {
		return err
	}
	fv.Set(e.Elem())
	return nil
}

func (c *Codec) unmarshalSlice(fv reflect.Value, hv string) error {
	var is []interface{}
	if err := json.Unmarshal([]byte(hv), &is); err != nil {
		return err
	}

	ft := fv.Type()
	fet := ft.Elem()
	s := reflect.MakeSlice(ft, 0, 1)
	e := reflect.New(fet)
	ei := e.Interface()
	for _, i := range is {
		bytes, _ := json.Marshal(i)
		if err := json.Unmarshal(bytes, &ei); err != nil {
			continue
		}
		s = reflect.Append(s, e.Elem())
	}
	fv.Set(s)
	return nil
}

func (c *Codec) unmarshalPtr(fv reflect.Value, hv string) error {
	switch fv.Type().Elem().Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fv.Set(reflect.ValueOf(&hv))
	case reflect.Struct:
		e := reflect.New(fv.Type())
		ei := e.Interface()
		if err := json.Unmarshal([]byte(hv), &ei); err != nil {
			return err
		}
		fv.Set(e.Elem())
	}
	return nil
}

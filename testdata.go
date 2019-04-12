package got

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
)

const tagName = "testdata"

// TestData extracts the contents of a directory into an annotated struct, using
// the "testdata" struct tag for configuration.

// The struct tag currently only supports passing a filename, but this will
// likely be expanded on in future versions.
func TestData(dir string, out interface{}) error {
	if out == nil {
		return nil
	}

	if k := reflect.TypeOf(out).Kind(); k != reflect.Ptr {
		return fmt.Errorf("pointer value required, got %s", k)
	}

	t := reflect.TypeOf(out).Elem()
	v := reflect.ValueOf(out).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)

		if tag == "" || tag == "-" {
			continue
		}

		file := filepath.Join(dir, tag)

		switch k := field.Type.Kind(); k {
		case reflect.String:
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			v.Field(i).SetString(string(data))
		case reflect.Slice:
			switch j := field.Type.Elem().Kind(); j {
			case reflect.Uint8:
				data, err := ioutil.ReadFile(file)
				if err != nil {
					return err
				}
				v.Field(i).Set(reflect.ValueOf(data))
			default:
				return fmt.Errorf("unsupported field slice kind %s", j)
			}
		case reflect.Map:
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			m := reflect.MakeMap(field.Type).Interface()
			if err := json.Unmarshal(data, &m); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(m))
		case reflect.Interface:
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			var x interface{}
			if err := json.Unmarshal(data, &x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x))
		default:
			return fmt.Errorf("unsupported field kind %s", k)
		}
	}

	return nil
}

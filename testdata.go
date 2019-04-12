package got

import (
	"encoding/json"
	"errors"
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
func TestData(dir string, out interface{}) (err error) {
	if out == nil {
		return errors.New("cannot use nil")
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

		data, err := ioutil.ReadFile(filepath.Join(dir, tag))
		if err != nil {
			return err
		}

		if field.Type.Kind() == reflect.String {
			v.Field(i).SetString(string(data))
			continue
		}

		switch filepath.Ext(tag) {
		case ".json":
			x := reflect.New(field.Type).Interface()
			if err := json.Unmarshal(data, x); err != nil {
				return err
			}
			v.Field(i).Set(reflect.ValueOf(x).Elem())
		default:
			v.Field(i).Set(reflect.ValueOf(data))
		}
	}

	return nil
}

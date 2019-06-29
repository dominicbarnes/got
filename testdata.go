package got

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
)

const tagName = "testdata"

// TestData extracts the contents of a directory into an annotated struct, using
// the "testdata" struct tag for configuration.
//
// The struct tag currently only supports passing a filename, but this will
// likely be expanded on in future versions.
func TestData(t TestingT, dir string, out interface{}) {
	t.Helper()

	if out == nil {
		t.Fatal("output cannot be nil")
		return
	}

	if k := reflect.TypeOf(out).Kind(); k != reflect.Ptr {
		t.Fatalf("output must be pointer value, instead got %s", k)
		return
	}

	typ := reflect.TypeOf(out).Elem()
	val := reflect.ValueOf(out).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get(tagName)

		if tag == "" || tag == "-" {
			continue
		}

		file := filepath.Join(dir, tag)
		t.Logf("%s: reading file %s", field.Name, file)
		data, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("%s: failed to read file: %s", field.Name, err.Error())
			return
		}

		if field.Type.Kind() == reflect.String {
			val.Field(i).SetString(string(data))
			continue
		}

		switch filepath.Ext(tag) {
		case ".json":
			x := reflect.New(field.Type).Interface()
			if err := json.Unmarshal(data, x); err != nil {
				t.Fatalf("%s: failed to parse %s as JSON: %s", field.Name, file, err.Error())
				return
			}
			val.Field(i).Set(reflect.ValueOf(x).Elem())
		default:
			val.Field(i).Set(reflect.ValueOf(data))
		}
	}
}

package got

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fatih/structtag"
)

const tagName = "testdata"

var osFileType = reflect.TypeOf(new(os.File))

// LoadTestData extracts the contents of a directory into an annotated struct,
// using the "testdata" struct tag for configuration.
//
// The primary argument for the struct tag is a path to a file. Depending on the
// type of the struct property, the file will be loaded from disk and decoded
// via reflection.
//
// By default, all the struct members with a testdata annotation are required
// and the test will fail if the file is not found or otherwise fails to load.
// You can specify "optional" on the struct tag to suppress this behavior.
//
// The easiest examples include using types []byte or string which a decently
// low-level. However, you can go even lower by using *os.File to get a raw file
// reference.
//
// If your file is a JSON file, you can use any arbitrary type and the file body
// will be unmarshalled into that type, allowing you to still get type-safety.
func LoadTestData(t TestingT, dir string, output interface{}) {
	t.Helper()

	if output == nil {
		t.Fatal("output cannot be nil")
		return
	}

	if k := reflect.TypeOf(output).Kind(); k != reflect.Ptr {
		t.Fatalf("output must be pointer value, instead got %s", k)
		return
	}

	typ := reflect.TypeOf(output).Elem()
	val := reflect.ValueOf(output).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			t.Logf("failed to parse struct tags: %s", err.Error())
			continue
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			continue
		}

		if tag.Name == "" || tag.Name == "-" {
			continue
		}

		file := filepath.Join(dir, tag.Name)
		t.Logf("%s: reading file %s", field.Name, file)

		if field.Type == osFileType {
			f, err := os.Open(file)
			if err != nil {
				t.Fatalf("%s: failed to open file: %s", field.Name, err.Error())
				return
			}
			val.Field(i).Set(reflect.ValueOf(f))
		} else {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				if tag.HasOption("optional") {
					if os.IsNotExist(err) {
						t.Logf("%s: optional file not found", field.Name)
					} else {
						t.Logf("%s: failed to read optional file: %s", field.Name, err.Error())
					}
					continue
				} else {
					t.Fatalf("%s: failed to read file: %s", field.Name, err.Error())
					return
				}
			}

			if field.Type.Kind() == reflect.String {
				val.Field(i).SetString(string(data))
				continue
			}

			switch filepath.Ext(file) {
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
}

// SaveGoldenTestData takes the data embedded in the input struct for properties
// with the "golden" parameter in their struct tag and saves it to disk.
//
// A common pattern for defining test cases is "golden files", which are
// basically test fixtures that are generally automatically generated when code
// is known to be working in a specific way. Future tests are run and the output
// is compared against these test fixtures to detect unintended differences.
func SaveGoldenTestData(t TestingT, input interface{}, dir string) {
	t.Helper()

	if input == nil {
		t.Fatal("output cannot be nil")
		return
	}

	if k := reflect.TypeOf(input).Kind(); k != reflect.Ptr {
		t.Fatalf("output must be pointer value, instead got %s", k)
		return
	}

	typ := reflect.TypeOf(input).Elem()
	val := reflect.ValueOf(input).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			t.Logf("failed to parse struct tags: %s", err.Error())
			continue
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			continue
		}

		if tag.Name == "" || tag.Name == "-" {
			continue
		}

		if !tag.HasOption("golden") {
			continue
		}

		file := filepath.Join(dir, tag.Name)
		t.Logf("%s: writing file %s", field.Name, file)

		data, err := encode(t, file, field, val.Field(i))
		if err != nil {
			t.Fatalf("%s: failed to write file %s: %s", field.Name, file, err)
			return
		}

		if len(data) > 0 {
			if err := ioutil.WriteFile(file, data, 0644); err != nil {
				t.Fatalf("%s: failed to write file %s: %s", field.Name, file, err)
				return
			}
		} else {
			if err := os.Remove(file); err != nil {
				if !os.IsNotExist(err) {
					t.Fatalf("%s: failed to delete file %s: %s", field.Name, file, err)
					return
				}
			}
		}
	}
}

func encode(t TestingT, file string, field reflect.StructField, val reflect.Value) ([]byte, error) {
	if field.Type.Kind() == reflect.String {
		return []byte(val.String()), nil
	} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Uint8 {
		return val.Bytes(), nil
	} else {
		switch filepath.Ext(file) {
		case ".json":
			return json.Marshal(val.Interface())
		}
	}

	return nil, nil
}

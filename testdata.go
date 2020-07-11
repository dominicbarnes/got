package got

import (
	"encoding/json"
	"errors"
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
// Raw Types
//
// The most low-level type is *os.File, which will open the file and put the
// reference into the struct. Use this if you need the most control over what
// happens after load.
//
// The next level up will be string and []byte which will be parsed directly
// from the file contents, with no additional transformation.
//
// Structured Types
//
// The next level up from that will be more complex data types like structs,
// maps and slices. In order to take advantage of these, you'll need to use a
// supported decoder.
//
// Currently, only JSON is supported, and is activated when the file has a .json
// file extension. In this case, the file contents are run through
// json.Unmarshal into the target type.
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

		f, err := openTagFile(file, tag)
		if err != nil {
			t.Fatalf("%s: failed to open file: %s", field.Name, err.Error())
			return
		} else if f == nil {
			t.Logf("%s: failed to open optional file", field.Name)
			continue
		}

		// if the target type is an *os.File, attempt no further transformation
		if field.Type == osFileType {
			val.Field(i).Set(reflect.ValueOf(f))
			continue
		}

		// read the file contents
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("%s: failed to read file: %s", field.Name, err.Error())
			return
		}

		// raw types
		if isBytes(field.Type) {
			val.Field(i).SetBytes(data)
			continue
		} else if isString(field.Type) {
			val.Field(i).SetString(string(data))
			continue
		}

		// structured types
		if filepath.Ext(file) == ".json" {
			v := reflect.New(field.Type).Interface()
			if err := json.Unmarshal(data, v); err != nil {
				t.Fatalf("%s: failed to parse as JSON: %s", field.Name, err.Error())
				return
			}
			val.Field(i).Set(reflect.ValueOf(v).Elem())
			continue
		}

		// add this log as a cue that this particular decode
		t.Fatalf("%s: file opened, but not decoded", field.Name)
	}
}

// SaveGoldenTestData takes struct data with the "golden" parameter in their
// struct tag and saves it to disk, making it the reverse of LoadTestData.
//
// A common pattern for defining test cases is using "golden files", which are
// test fixtures that are automatically generated when code is known to be
// working in a specific way. Future tests are run and the output is compared
// against these test fixtures to detect unintended differences.
//
// Raw Types
//
// Currently, *os.File is not supported here due to the ambiguity of writing
// back the same file reference.
//
// The types string and []byte are written to disk as-is.
//
// Structured Types
//
// Currently, only JSON is supported with any other types, and the value is
// marshalled (with indent) and written to disk.
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

		data, err := encode(file, field, val.Field(i))
		if err != nil {
			t.Fatalf("%s: failed to write file %s: %s", field.Name, file, err)
			return
		}

		if len(data) > 0 {
			if err := ioutil.WriteFile(file, data, 0644); err != nil {
				t.Fatalf("%s: failed to write file %s: %s", field.Name, file, err)
				return
			}
		} else if tag.HasOption("omitempty") {
			if err := os.Remove(file); err != nil {
				if !os.IsNotExist(err) {
					t.Fatalf("%s: failed to delete file %s: %s", field.Name, file, err)
					return
				}
			}
		}
	}
}

func openTagFile(file string, tag *structtag.Tag) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		if tag.HasOption("optional") && os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func isString(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.String
}

func isBytes(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Slice && targetType.Elem().Kind() == reflect.Uint8
}

func encode(file string, field reflect.StructField, val reflect.Value) ([]byte, error) {
	if field.Type == osFileType {
		return nil, errors.New("saving *os.File is not currently supported")
	}

	if isBytes(field.Type) {
		return val.Bytes(), nil
	} else if isString(field.Type) {
		return []byte(val.String()), nil
	}

	switch filepath.Ext(file) {
	case ".json":
		return json.MarshalIndent(val.Interface(), "", "  ")
	}

	return nil, nil
}

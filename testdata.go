package got

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/structtag"
)

const tagName = "testdata"

// LoadTestData extracts the contents of a directory into a annotated structs,
// using the "testdata" struct tag for configuration.
//
// Struct Tag Format
//
// The primary argument for the struct tag is a path to a file. Depending on the
// type of the struct property, the file will be loaded from disk and decoded
// via reflection.
//
//   testdata:"input.txt"
//
// By default, all the struct members with a testdata annotation are required
// and the test will fail if the file is not found or otherwise fails to load.
// You can specify "optional" on the struct tag to suppress this behavior.
//
//   testdata:"state.json,optional"
//
// There are other struct tag options, but do not apply for the purposes of this
// function. Look to SaveTestData for additional struct tag options.
//
// Raw Types
//
// The types string and []byte which will be parsed directly from the file
// contents, with no additional transformation.
//
// Struct Types
//
// When using structs, the contents of the file will be decoded depending on the
// file extension.
//
// Currently, only JSON is supported, and is activated when the file has a .json
// file extension. In this case, the file contents are run through
// json.Unmarshal into the target type.
//
// Map Types
//
// Maps with string keys are given special treatment. When used, the filename
// can be treated as a glob pattern, provided it includes a "*" character.
//
// When a glob pattern is detected, the key is treated as the filename (relative
// to dir) as the key and the value will be decoded as described above.
//
// If a pattern is not detected, then the value will be decoded directly like
// other structs.
func LoadTestData(t *testing.T, dir string, output ...interface{}) {
	t.Helper()

	for _, v := range output {
		if err := loadDir(dir, v); err != nil {
			t.Fatal(err)
		}
	}
}

func loadDir(dir string, output interface{}) error {
	if output == nil {
		return errors.New("output cannot be nil")
	}

	if k := reflect.TypeOf(output).Kind(); k != reflect.Ptr {
		return fmt.Errorf("output must be pointer value, instead got %s", k)
	}

	typ := reflect.TypeOf(output).Elem()
	val := reflect.ValueOf(output).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return fmt.Errorf("%s: failed to parse struct tag %s: %w", field.Name, tagName, err)
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			continue
		}

		if tag.Name == "" || tag.Name == "-" {
			continue
		}

		file := filepath.Join(dir, tag.Name)

		if isMap(field.Type) && strings.Contains(tag.Name, "*") {
			matches, err := filepath.Glob(file)
			if err != nil {
				return fmt.Errorf("%s: failed to list files %s: %w", field.Name, file, err)
			}

			m := reflect.MakeMap(field.Type)

			for _, match := range matches {
				rel, err := filepath.Rel(dir, match)
				if err != nil {
					return fmt.Errorf("%s: failed to resolve file %s: %w", field.Name, match, err)
				}

				key := reflect.ValueOf(rel)
				value := reflect.New(m.Type().Elem()).Elem()

				if err := loadFile(match, field, value, tag); err != nil {
					return err
				}

				m.SetMapIndex(key, value)
			}

			val.Field(i).Set(m)
			continue
		}

		if err := loadFile(file, field, val.Field(i), tag); err != nil {
			return err
		}
	}

	return nil
}

func loadFile(file string, field reflect.StructField, value reflect.Value, tag *structtag.Tag) error {
	f, err := openTagFile(file, tag)
	if err != nil {
		return fmt.Errorf("%s: failed to open file %s: %w", field.Name, file, err)
	} else if f == nil {
		return nil
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("%s: failed to read file %s: %w", field.Name, file, err)
	}

	// raw types
	if isBytes(value.Type()) {
		value.SetBytes(data)
		return nil
	} else if isString(value.Type()) {
		value.SetString(string(data))
		return nil
	}

	// structured types
	if filepath.Ext(file) == ".json" {
		v := reflect.New(value.Type()).Interface()
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("%s: failed to parse contents of %s as JSON: %w", field.Name, file, err)
		}
		value.Set(reflect.ValueOf(v).Elem())
		return nil
	}

	return fmt.Errorf("%s: file opened, but not decoded", field.Name)
}

// SaveTestData takes data from the input structs and saves it to disk, making
// this the reverse of LoadTestData.
//
// A common pattern for defining test cases is using "golden files", which are
// test fixtures that are automatically generated when code is known to be
// working in a specific way. Future tests are run and the output is compared
// against these test fixtures to detect unintended differences.
//
// By default, a file will always be written, even if that file turns out to be
// empty after encoding it. If you would prefer these empty files to not be
// present at all, include the "omitempty" option as well.
//
//   testdata:"error.txt,omitempty"
//
// Raw Types
//
// The types string and []byte are written to disk as-is.
//
// Struct Types
//
// Structs are encoded using json.MarshalIndent.
//
// Map Types
//
// Similar to LoadTestData, maps with string keys are given special treatment.
// If the path includes a "*" (a glob pattern), then the keys are treated as
// paths relative to dir and the data is encoded as described above.
//
// If a glob pattern is not used, then the value will be encoded like structs.
func SaveTestData(t *testing.T, dir string, input ...interface{}) {
	t.Helper()

	for _, v := range input {
		if err := saveDir(dir, v); err != nil {
			t.Fatal(err)
		}
	}
}

func saveDir(dir string, input interface{}) error {
	if input == nil {
		return errors.New("output cannot be nil")
	}

	if k := reflect.TypeOf(input).Kind(); k != reflect.Ptr {
		return fmt.Errorf("output must be pointer value, instead got %s", k)
	}

	typ := reflect.TypeOf(input).Elem()
	val := reflect.ValueOf(input).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return fmt.Errorf("%s: failed to parse struct tags: %w", field.Name, err)
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			return fmt.Errorf("%s: failed to parse struct tag: %w", field.Name, err)
		}

		if tag.Name == "" || tag.Name == "-" {
			continue
		}

		if isMap(field.Type) && strings.Contains(tag.Name, "*") {
			iter := val.Field(i).MapRange()

			for iter.Next() {
				k := iter.Key()
				v := iter.Value()

				file := filepath.Join(dir, k.String())

				if err := saveFile(file, field, v, tag); err != nil {
					return err
				}
			}

			continue
		}

		file := filepath.Join(dir, tag.Name)

		if err := saveFile(file, field, val.Field(i), tag); err != nil {
			return err
		}
	}

	return nil
}

func saveFile(file string, field reflect.StructField, val reflect.Value, tag *structtag.Tag) error {
	data, err := encode(file, field, val)
	if err != nil {
		return fmt.Errorf("%s: failed to encode file %s: %w", field.Name, file, err)
	}

	if len(data) > 0 {
		dir := filepath.Dir(file)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s: failed to create dir %s: %s", field.Name, dir, err)
		}

		if err := ioutil.WriteFile(file, data, 0644); err != nil {
			return fmt.Errorf("%s: failed to write file %s: %s", field.Name, file, err)
		}
	} else if tag.HasOption("omitempty") {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("%s: failed to delete file %s: %s", field.Name, file, err)
			}
		}
	}

	return nil
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

func isMap(targetType reflect.Type) bool {
	return targetType.Kind() == reflect.Map && isString(targetType.Key())
}

func encode(file string, field reflect.StructField, val reflect.Value) ([]byte, error) {
	if isBytes(val.Type()) {
		return val.Bytes(), nil
	} else if isString(val.Type()) {
		return []byte(val.String()), nil
	}

	switch filepath.Ext(file) {
	case ".json":
		return json.MarshalIndent(val.Interface(), "", "  ")
	}

	return nil, nil
}

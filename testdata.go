package got

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/dominicbarnes/got/codec"
	"github.com/fatih/structtag"
	"github.com/google/go-cmp/cmp"
)

var updateGolden bool

func init() {
	flag.BoolVar(&updateGolden, "update-golden", false, "instruct got.Assert to update golden files")
}

const tagName = "testdata"

// Load extracts the contents of dir into values which are structs annotated
// with the "testdata" struct tag.
//
// The main parameter of the struct tag will be a path to a file relative to the
// input directory.
//
// Fields with string or []byte as their types will be populated with the raw
// contents of the file.
//
// Struct values will be decoded using the file extension to map to a [Codec].
// For example, ".json" files can be processed using [JSONCodec] if it has been
// registered. Additional codecs (eg: YAML, TOML) can be registered if desired.
//
// Map values, by default, are decoded using the relevant [Codec].
//
// There is also a special mode that works files more dynamically, which is
// useful for highly variable outputs and is enabled with the "explode" option.
// When enabled, the struct tag name is treated as a glob pattern. The map is
// populated with a key corresponding to a relative filepath while the value can
// be any of the types described above.
func Load(t T, dir string, values ...any) {
	t.Helper()

	if err := loadDirs([]string{dir}, values...); err != nil {
		t.Fatal(err.Error())
	}
}

// LoadDirs is the same as Load but accepts multiple input directories, which
// can be used to set up test cases from a common/shared location while allowing
// an individual test-case to include it's own specific configuration.
func LoadDirs(t T, dirs []string, values ...any) {
	t.Helper()

	if err := loadDirs(dirs, values...); err != nil {
		t.Fatal(err.Error())
	}
}

// Assert ensures that all the fields within the struct values match what is on
// disk, using reflection to Load a fresh copy and then comparing the 2 structs
// using go-cmp to perform the equality check.
//
// When the "test.update-golden" flag is provided, the contents of each value
// struct will be persisted to disk instead. This allows any test to easily
// update their "golden files" and also do the assertion transparently.
func Assert(t T, dir string, values ...any) {
	t.Helper()

	if err := assert(dir, values...); err != nil {
		t.Fatal(err.Error())
	}
}

func assert(dir string, values ...any) error {
	if len(values) == 0 {
		return errors.New("at least 1 value required")
	}

	for _, actual := range values {
		if updateGolden {
			if err := saveDir(dir, actual); err != nil {
				return err
			}

			continue
		}

		expected := reflect.New(reflect.TypeOf(actual).Elem()).Interface()

		if err := loadDirs([]string{dir}, expected); err != nil {
			return fmt.Errorf("%T: %w", actual, err)
		}

		if !cmp.Equal(expected, actual) {
			return errors.New(cmp.Diff(expected, actual))
		}
	}

	return nil
}

func loadDirs(inputs []string, outputs ...any) error {
	if len(outputs) == 0 {
		return errors.New("at least 1 output required")
	}

	for _, output := range outputs {
		if err := loadDir(inputs, output); err != nil {
			return err
		}
	}

	return nil
}

func loadDir(inputs []string, output any) error {
	if output == nil {
		return errors.New("output cannot be nil")
	}

	if k := reflect.TypeOf(output).Kind(); k != reflect.Ptr {
		return fmt.Errorf("output must be a pointer, instead got %s", k)
	}

	typ := reflect.TypeOf(output).Elem()
	val := reflect.ValueOf(output).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return fmt.Errorf("%s: failed to parse struct tags: %w", field.Name, err)
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			continue
		} else if tag.Name == "" || tag.Name == "-" {
			continue
		}

		for _, input := range inputs {
			if err := loadDirInput(input, tag, field, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadDirInput(input string, tag *structtag.Tag, field reflect.StructField, value reflect.Value) error {
	file := filepath.Join(input, tag.Name)

	if isMap(field.Type) && tag.HasOption("explode") {
		matches, err := filepath.Glob(file)
		if err != nil {
			return fmt.Errorf("%s: failed to list files %s: %w", field.Name, file, err)
		}

		m := reflect.MakeMap(field.Type)

		for _, match := range matches {
			rel, err := filepath.Rel(input, match)
			if err != nil {
				return fmt.Errorf("%s: failed to resolve file %s: %w", field.Name, match, err)
			}

			key := reflect.ValueOf(rel)
			value := reflect.New(m.Type().Elem()).Elem()

			if err := loadFile(match, field, value); err != nil {
				return err
			}

			m.SetMapIndex(key, value)
		}

		if m.Len() > 0 {
			value.Set(m)
		}

		return nil
	}

	if err := loadFile(file, field, value); err != nil {
		return err
	}

	return nil
}

func loadFile(file string, field reflect.StructField, value reflect.Value) error {
	f, err := openTagFile(file)
	if err != nil {
		return fmt.Errorf("%s: %w", field.Name, err)
	} else if f == nil {
		return nil
	}

	data, err := io.ReadAll(f)
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

	ext := filepath.Ext(file)
	codec, err := codec.Get(ext)
	if err != nil {
		return fmt.Errorf("%s: failed to get codec for file extension %q", field.Name, ext)
	}

	p := reflect.New(value.Type())
	p.Elem().Set(value) // preserve any prior values
	if err := codec.Unmarshal(data, p.Interface()); err != nil {
		return fmt.Errorf("%s: failed to unmarshal %s: %w", field.Name, file, err)
	}
	value.Set(p.Elem()) // overwrite with the updated value
	return nil
}

func saveDir(dir string, input any) error {
	if input == nil {
		return errors.New("input cannot be nil")
	}

	if k := reflect.TypeOf(input).Kind(); k != reflect.Ptr {
		return fmt.Errorf("input must be a pointer, instead got %s", k)
	}

	typ := reflect.TypeOf(input).Elem()
	val := reflect.ValueOf(input).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		tags, err := structtag.Parse(string(field.Tag))
		if err != nil {
			return fmt.Errorf("%s: failed to parse struct tags: %w", field.Name, err)
		}

		tag, err := tags.Get(tagName)
		if err != nil {
			continue
		} else if tag.Name == "" || tag.Name == "-" {
			continue
		}

		if err := saveDirField(dir, tag, field, value); err != nil {
			return err
		}
	}

	return nil
}

func saveDirField(dir string, tag *structtag.Tag, field reflect.StructField, value reflect.Value) error {
	if isMap(field.Type) && tag.HasOption("explode") {
		iter := value.MapRange()

		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			file := filepath.Join(dir, k.String())
			if err := saveFile(file, field, v); err != nil {
				return err
			}
		}

		return nil
	}

	file := filepath.Join(dir, tag.Name)
	if err := saveFile(file, field, value); err != nil {
		return err
	}

	return nil
}

func saveFile(file string, field reflect.StructField, val reflect.Value) error {
	data, err := encode(file, field, val)
	if err != nil {
		return fmt.Errorf("%s: failed to encode file %s: %w", field.Name, file, err)
	}

	if len(data) == 0 {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("%s: failed to delete file %s: %w", field.Name, file, err)
			}
		}
	} else {
		dir := filepath.Dir(file)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("%s: failed to create dir %s: %w", field.Name, dir, err)
		}

		if err := os.WriteFile(file, data, 0644); err != nil {
			return fmt.Errorf("%s: failed to write file %s: %w", field.Name, file, err)
		}
	}

	return nil
}

func encode(file string, field reflect.StructField, val reflect.Value) ([]byte, error) {
	switch {
	case val.IsZero():
		return nil, nil
	case isBytes(val.Type()):
		return val.Bytes(), nil
	case isString(val.Type()):
		return []byte(val.String()), nil
	}

	ext := filepath.Ext(file)
	codec, err := codec.Get(ext)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get codec for file extension %q", field.Name, ext)
	}
	return codec.Marshal(val.Interface())
}

func openTagFile(file string) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		// suppress "not found" errors
		if os.IsNotExist(err) {
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

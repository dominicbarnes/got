package got

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDir(t *testing.T) {
	spec := []struct {
		name     string
		dir      string
		input    interface{}
		expected interface{}
		fail     bool
	}{
		// text
		{
			name:     "string",
			dir:      "text",
			input:    &testTextString{},
			expected: &testTextString{Input: "hello world"},
		},
		{
			name:     "bytes",
			dir:      "text",
			input:    &testTextBytes{},
			expected: &testTextBytes{Input: []byte("hello world")},
		},
		{
			name:     "json raw",
			dir:      "json",
			input:    &testJSONRaw{},
			expected: &testJSONRaw{Input: json.RawMessage("{\n  \"hello\": \"world\"\n}")},
		},
		// json
		{
			name:  "json struct",
			dir:   "json",
			input: &testJSONStruct{},
			expected: &testJSONStruct{Input: struct {
				Hello string `json:"hello"`
			}{"world"}},
		},
		{
			name:  "json map",
			dir:   "json",
			input: &testJSONMap{},
			expected: &testJSONMap{
				Input: map[string]interface{}{"hello": "world"},
			},
		},
		{
			name: "json should not clobber",
			dir:  "json",
			input: &testJSONMap{
				Input: map[string]interface{}{
					"hello": "dave", // should be overwritten
					"a":     "A",    // should not be deleted
				},
			},
			expected: &testJSONMap{
				Input: map[string]interface{}{
					"hello": "world",
					"a":     "A",
				},
			},
		},
		{
			name:  "json invalid",
			dir:   "json",
			input: &testJSONInvalid{},
			fail:  true,
		},
		{
			name:     "json optional",
			dir:      "json",
			input:    &testJSONOptional{},
			expected: &testJSONOptional{},
		},
		// yaml
		{
			name:  "yaml struct",
			dir:   "yaml",
			input: &testYAMLStruct{},
			expected: &testYAMLStruct{Input: struct {
				Hello string `yaml:"hello"`
			}{"world"}},
		},
		{
			name:  "yaml map",
			dir:   "yaml",
			input: &testYAMLMap{},
			expected: &testYAMLMap{
				Input: map[string]interface{}{"hello": "world"},
			},
		},
		{
			name: "yaml should not clobber",
			dir:  "yaml",
			input: &testYAMLMap{
				Input: map[string]interface{}{
					"hello": "dave", // should be overwritten
					"a":     "A",    // should not be deleted
				},
			},
			expected: &testYAMLMap{
				Input: map[string]interface{}{
					"hello": "world",
					"a":     "A",
				},
			},
		},
		{
			name:  "yaml invalid",
			dir:   "yaml",
			input: &testYAMLInvalid{},
			fail:  true,
		},
		{
			name:     "yaml optional",
			dir:      "yaml",
			input:    &testYAMLOptional{},
			expected: &testYAMLOptional{},
		},
		// multiple
		{
			name:  "multiple",
			dir:   "multiple",
			input: &testMultiple{},
			expected: &testMultiple{
				A: "A",
				B: []byte("B"),
			},
		},
		{
			name:  "multiple map",
			dir:   "multiple",
			input: &testMultipleMap{},
			expected: &testMultipleMap{
				Files: map[string]string{
					"a.txt": "A",
					"b.txt": "B",
				},
			},
		},
		// misc
		{
			name:  "not pointer",
			dir:   "text",
			input: struct{}{},
			fail:  true,
		},
		{
			name:  "missing file",
			dir:   "text",
			input: &testMissing{},
			fail:  true,
		},
		{
			name:     "missing optional file",
			dir:      "text",
			input:    &testMissingOptional{},
			expected: &testMissingOptional{},
		},
		{
			name:     "missing struct tag",
			dir:      "text",
			input:    &testMissingStructTag{},
			expected: &testMissingStructTag{},
		},
		{
			name:     "empty struct tag",
			dir:      "text",
			input:    &testEmptyStructTag{},
			expected: &testEmptyStructTag{},
		},
		{
			name:     "dash struct tag",
			dir:      "text",
			input:    &testDashStructTag{},
			expected: &testDashStructTag{},
		},
		{
			name: "nil",
			dir:  "text",
			fail: true,
		},
	}

	for _, test := range spec {
		t.Run(test.name, func(t *testing.T) {
			err := loadDir(filepath.Join("testdata", test.dir), test.input)
			if test.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, test.expected, test.input)
			}
		})
	}
}

func TestSaveTestData(t *testing.T) {
	spec := []struct {
		name     string
		expected interface{}
		fail     bool
	}{
		{
			name:     "string",
			expected: &testTextString{Input: "hello world"},
		},
		{
			name:     "bytes",
			expected: &testTextBytes{Input: []byte("hello world")},
		},
		{
			name:     "json raw",
			expected: &testJSONRaw{Input: json.RawMessage("{\n  \"hello\": \"world\"\n}")},
		},
		{
			name: "json struct",
			expected: &testJSONStruct{Input: struct {
				Hello string "json:\"hello\""
			}{"world"}},
		},
		{
			name: "json map",
			expected: &testJSONMap{
				Input: map[string]interface{}{"hello": "world"},
			},
		},
		{
			name: "multiple",
			expected: &testMultipleMap{
				Files: map[string]string{
					"a.txt": "A",
					"b.txt": "B",
				},
			},
		},
	}

	for _, test := range spec {
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", test.name)
			require.NoError(t, err)

			if test.fail {
				require.Error(t, saveDir(dir, test.expected))
			} else {
				require.NoError(t, saveDir(dir, test.expected))

				actual := reflect.New(reflect.TypeOf(test.expected).Elem()).Interface()
				require.NoError(t, loadDir(dir, actual))
				require.EqualValues(t, test.expected, actual)
			}
		})
	}

	t.Run("omitempty", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		type TestCase struct {
			Output string `testdata:"output.txt,optional,omitempty"`
		}

		expected := TestCase{Output: "hello world"}
		require.NoError(t, saveDir(dir, &expected))

		expected.Output = "" // trigger a delete
		require.NoError(t, saveDir(dir, &expected))

		var actual TestCase
		require.NoError(t, loadDir(dir, &actual))

		require.EqualValues(t, expected, actual)

		_, err = os.Open(filepath.Join(dir, "output.txt"))
		require.True(t, os.IsNotExist(err), "file should have been deleted")
	})
}

// text

type testTextString struct {
	Input string `testdata:"input.txt"`
}

type testTextBytes struct {
	Input []byte `testdata:"input.txt"`
}

// json

type testJSONRaw struct {
	Input json.RawMessage `testdata:"input.json"`
}

type testJSONMap struct {
	Input map[string]interface{} `testdata:"input.json"`
}

type testJSONStruct struct {
	Input struct {
		Hello string `json:"hello"`
	} `testdata:"input.json"`
}

type testJSONInvalid struct {
	Input struct{} `testdata:"invalid.json"`
}

type testJSONOptional struct {
	Input struct{} `testdata:"does-not-exist.json,optional"`
}

// yaml

type testYAMLMap struct {
	Input map[string]interface{} `testdata:"input.yaml"`
}

type testYAMLStruct struct {
	Input struct {
		Hello string `yaml:"hello"`
	} `testdata:"input.yaml"`
}

type testYAMLInvalid struct {
	Input struct{} `testdata:"invalid.yaml"`
}

type testYAMLOptional struct {
	Input struct{} `testdata:"does-not-exist.yaml,optional"`
}

// multiple

type testMultiple struct {
	A string `testdata:"a.txt"`
	B []byte `testdata:"b.txt"`
}

type testMultipleMap struct {
	Files map[string]string `testdata:"*.txt"`
}

// misc

type testMissing struct {
	Input string `testdata:"does-not-exist"`
}

type testMissingOptional struct {
	Input string `testdata:"does-not-exist,optional"`
}

type testMissingStructTag struct {
	Missing string
}

type testEmptyStructTag struct {
	Empty string `testdata:""`
}

type testDashStructTag struct {
	Empty string `testdata:"-"`
}

func ExampleLoadTestData() {
	t := new(testing.T) // not necessary in normal test code

	type TestCase struct {
		Input    string `testdata:"input.txt"`
		Expected string `testdata:"expected.txt"`
	}

	var test TestCase
	LoadTestData(t, "testdata", &test)

	actual := strings.ToUpper(test.Input)
	if actual != test.Expected {
		t.Fatalf("actual value '%s' did not match expected value '%s'", actual, test.Expected)
	}
}

func ExampleLoadTestData_jSON() {
	t := new(testing.T) // not necessary in normal test code

	type Input struct {
		Text string `json:"text"`
	}

	type Output struct {
		Text string `json:"text"`
	}

	type TestCase struct {
		Input    Input  `testdata:"input.json"`
		Expected Output `testdata:"expected.json"`
	}

	var test TestCase
	LoadTestData(t, "testdata", &test)

	actual := strings.ToUpper(test.Input.Text)
	if actual != test.Expected.Text {
		t.Fatalf("actual value '%s' did not match expected value '%s'", actual, test.Expected.Text)
	}
}

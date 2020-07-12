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
		expected interface{}
		fail     bool
	}{
		{
			name:     "string",
			dir:      "text",
			expected: testTextString{Input: "hello world"},
		},
		{
			name:     "bytes",
			dir:      "text",
			expected: testTextBytes{Input: []byte("hello world")},
		},
		{
			name:     "json raw",
			dir:      "json",
			expected: testJSONRaw{Input: json.RawMessage("{\n  \"hello\": \"world\"\n}")},
		},
		{
			name: "json struct",
			dir:  "json",
			expected: testJSONStruct{Input: struct {
				Hello string `json:"hello"`
			}{"world"}},
		},
		{
			name:     "json invalid",
			dir:      "json",
			expected: testJSONInvalid{},
			fail:     true,
		},
		{
			name:     "json optional",
			dir:      "json",
			expected: testJSONOptional{},
		},
		{
			name: "multiple",
			dir:  "multiple",
			expected: testMultiple{
				A: "A",
				B: []byte("B"),
			},
		},
		{
			name: "multiple map",
			dir:  "multiple",
			expected: testMultipleMap{
				Files: map[string]string{
					"a.txt": "A",
					"b.txt": "B",
				},
			},
		},
		{
			name:     "not pointer",
			dir:      "text",
			expected: struct{}{},
			fail:     true,
		},
		{
			name:     "missing file",
			dir:      "text",
			expected: testTextMissing{},
			fail:     true,
		},
		{
			name:     "missing optional file",
			dir:      "text",
			expected: testTextMissingOptional{},
		},
		{
			name:     "missing struct tag",
			dir:      "text",
			expected: testMissingStructTag{},
		},
		{
			name:     "empty struct tag",
			dir:      "text",
			expected: testEmptyStructTag{},
		},
		{
			name:     "dash struct tag",
			dir:      "text",
			expected: testDashStructTag{},
		},
		{
			name: "nil",
			dir:  "text",
			fail: true,
		},
	}

	for _, test := range spec {
		t.Run(test.name, func(t *testing.T) {
			var input interface{}
			if test.expected != nil {
				if s, ok := test.expected.(struct{}); ok {
					input = s
				} else {
					input = reflect.New(reflect.TypeOf(test.expected)).Interface()
				}
			}

			err := loadDir(filepath.Join("testdata", test.dir), input)
			if test.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				actual := reflect.ValueOf(input).Elem().Interface()
				require.EqualValues(t, test.expected, actual)
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
			Output string `testdata:"output.txt,optional,golden,omitempty"`
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

type testTextString struct {
	Input string `testdata:"input.txt"`
}

type testTextBytes struct {
	Input []byte `testdata:"input.txt"`
}

type testJSONRaw struct {
	Input json.RawMessage `testdata:"input.json"`
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

type testMultiple struct {
	A string `testdata:"a.txt"`
	B []byte `testdata:"b.txt"`
}

type testMultipleMap struct {
	Files map[string]string `testdata:"*.txt"`
}

type testTextMissing struct {
	Input string `testdata:"does-not-exist"`
}

type testTextMissingOptional struct {
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

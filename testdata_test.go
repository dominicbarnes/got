package got_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	reflect "reflect"
	"strings"
	"testing"

	. "github.com/dominicbarnes/got"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLoadTestData(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/input.txt")

		type TestCase struct {
			Input string `testdata:"input.txt"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		expected := TestCase{"hello world"}
		require.EqualValues(t, expected, actual)
	})

	t.Run("bytes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/input.txt")

		type TestCase struct {
			Input []byte `testdata:"input.txt"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		expected := TestCase{[]byte("hello world")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/input.txt")

		type TestCase struct {
			Input *os.File `testdata:"input.txt"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)

		require.EqualValues(t, filepath.Join("testdata/text/input.txt"), actual.Input.Name())
	})

	t.Run("json", func(t *testing.T) {
		t.Run("raw", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/input.json")

			type TestCase struct {
				Input json.RawMessage `testdata:"input.json"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
			expected := TestCase{json.RawMessage("{\n  \"hello\": \"world\"\n}")}
			require.EqualValues(t, expected, actual)
		})

		t.Run("map", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/input.json")

			type TestCase struct {
				Input map[string]interface{} `testdata:"input.json"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
			expected := TestCase{map[string]interface{}{"hello": "world"}}
			require.EqualValues(t, expected, actual)
		})

		t.Run("interface", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/input.json")

			type TestCase struct {
				Input interface{} `testdata:"input.json"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
			expected := TestCase{map[string]interface{}{"hello": "world"}}
			require.EqualValues(t, expected, actual)
		})

		t.Run("struct", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/input.json")

			type TestCase struct {
				Input struct {
					Hello string `json:"hello"`
				} `testdata:"input.json"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
			expected := TestCase{
				Input: struct {
					Hello string `json:"hello"`
				}{
					Hello: "world",
				},
			}
			require.EqualValues(t, expected, actual)
		})

		t.Run("invalid", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/invalid.json")
			mockt.EXPECT().Fatalf("%s: failed to parse as JSON: %s", "Input", "unexpected end of JSON input")

			type TestCase struct {
				Input map[string]interface{} `testdata:"invalid.json"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
		})

		t.Run("optional", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/json/input.json")

			type TestCase struct {
				Input struct {
					Hello string `json:"hello"`
				} `testdata:"input.json,optional"`
			}
			var actual TestCase
			LoadTestData(mockt, "testdata/json", &actual)
			expected := TestCase{
				Input: struct {
					Hello string `json:"hello"`
				}{
					Hello: "world",
				},
			}
			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("multiple", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "A", "testdata/multiple/a.txt")
		mockt.EXPECT().Logf("%s: reading file %s", "B", "testdata/multiple/b.txt")

		type TestCase struct {
			A string `testdata:"a.txt"`
			B []byte `testdata:"b.txt"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/multiple", &actual)
		expected := TestCase{"A", []byte("B")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Fatal("output cannot be nil")

		LoadTestData(mockt, "testdata/text", nil)
	})

	t.Run("non-pointer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Fatalf("output must be pointer value, instead got %s", reflect.Struct)

		LoadTestData(mockt, "testdata/text", struct{}{})
	})

	t.Run("missing file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/does-not-exist")
		mockt.EXPECT().Fatalf("%s: failed to open file: %s", "Input", "open testdata/text/does-not-exist: no such file or directory")

		type TestCase struct {
			Input string `testdata:"does-not-exist"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
	})

	t.Run("missing optional file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/does-not-exist")
		mockt.EXPECT().Logf("%s: failed to open optional file", "Input")

		type TestCase struct {
			Input string `testdata:"does-not-exist,optional"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
	})

	t.Run("missing struct tag", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()

		type TestCase struct {
			Missing string
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		require.Empty(t, actual.Missing)
	})

	t.Run("empty struct tag", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()

		type TestCase struct {
			Missing string `testdata:""`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		require.Empty(t, actual.Missing)
	})

	t.Run("dashed struct tag", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()

		type TestCase struct {
			Missing string `testdata:"-"`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		require.Empty(t, actual.Missing)
	})

	t.Run("invalid struct tag", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("failed to parse struct tags: %s", "bad syntax for struct tag value")

		type TestCase struct {
			Missing string `json:"missing-last-quote`
		}
		var actual TestCase
		LoadTestData(mockt, "testdata/text", &actual)
		require.Empty(t, actual.Missing)
	})
}

func TestSaveGoldenTestData(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.txt"))

		type TestCase struct {
			Output string `testdata:"output.txt,golden"`
		}

		expected := TestCase{Output: "hello world"}
		SaveGoldenTestData(mockt, &expected, dir)

		var actual TestCase
		LoadTestData(t, dir, &actual)

		require.EqualValues(t, expected, actual)
	})

	t.Run("bytes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.txt"))

		type TestCase struct {
			Output []byte `testdata:"output.txt,golden"`
		}

		expected := TestCase{Output: []byte("hello world")}
		SaveGoldenTestData(mockt, &expected, dir)

		var actual TestCase
		LoadTestData(t, dir, &actual)

		require.EqualValues(t, expected, actual)
	})

	t.Run("json", func(t *testing.T) {
		t.Run("raw", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dir, err := ioutil.TempDir("", "")
			require.NoError(t, err)

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.json"))

			type TestCase struct {
				Output json.RawMessage `testdata:"output.json,golden"`
			}

			expected := TestCase{Output: json.RawMessage(`{"hello":"world"}`)}
			SaveGoldenTestData(mockt, &expected, dir)

			var actual TestCase
			LoadTestData(t, dir, &actual)

			require.EqualValues(t, expected, actual)
		})

		t.Run("map", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dir, err := ioutil.TempDir("", "")
			require.NoError(t, err)

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.json"))

			type TestCase struct {
				Output map[string]interface{} `testdata:"output.json,golden"`
			}

			expected := TestCase{Output: map[string]interface{}{"hello": "world"}}
			SaveGoldenTestData(mockt, &expected, dir)

			var actual TestCase
			LoadTestData(t, dir, &actual)

			require.EqualValues(t, expected, actual)
		})

		t.Run("interface", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dir, err := ioutil.TempDir("", "")
			require.NoError(t, err)

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.json"))

			type TestCase struct {
				Output interface{} `testdata:"output.json,golden"`
			}

			expected := TestCase{Output: map[string]interface{}{"hello": "world"}}
			SaveGoldenTestData(mockt, &expected, dir)

			var actual TestCase
			LoadTestData(t, dir, &actual)

			require.EqualValues(t, expected, actual)
		})

		t.Run("struct", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dir, err := ioutil.TempDir("", "")
			require.NoError(t, err)

			mockt := NewMockTestingT(ctrl)
			mockt.EXPECT().Helper()
			mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.json"))

			type TestCase struct {
				Output struct {
					Hello string `json:"hello"`
				} `testdata:"output.json,golden"`
			}

			var expected TestCase
			expected.Output.Hello = "world"
			SaveGoldenTestData(mockt, &expected, dir)

			var actual TestCase
			LoadTestData(t, dir, &actual)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("omitempty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper().Times(2)
		mockt.EXPECT().Logf("%s: writing file %s", "Output", filepath.Join(dir, "output.txt")).Times(2)

		type TestCase struct {
			Output string `testdata:"output.txt,optional,golden,omitempty"`
		}

		expected := TestCase{Output: "hello world"}
		SaveGoldenTestData(mockt, &expected, dir)

		expected.Output = "" // trigger a delete
		SaveGoldenTestData(mockt, &expected, dir)

		var actual TestCase
		LoadTestData(t, dir, &actual)

		require.EqualValues(t, expected, actual)

		_, err = os.Open(filepath.Join(dir, "output.txt"))
		require.True(t, os.IsNotExist(err), "file should have been deleted")
	})
}

func ExampleLoadTestData(t *testing.T) {
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

func ExampleLoadTestData_jSON(t *testing.T) {
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

func ExampleLoadTestData_file(t *testing.T) {
	type TestCase struct {
		Input    *os.File `testdata:"input.txt"`
		Expected string   `testdata:"expected.txt"`
	}

	var test TestCase
	LoadTestData(t, "testdata", &test)

	input, err := ioutil.ReadAll(test.Input)
	if err != nil {
		t.Fatalf("failed to read input file: %s", err.Error())
	}

	actual := strings.ToUpper(string(input))
	if actual != test.Expected {
		t.Fatalf("actual value '%s' did not match expected value '%s'", actual, test.Expected)
	}
}

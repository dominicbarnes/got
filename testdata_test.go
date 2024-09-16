package got

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		testLoadError(t, "text", nil, "output cannot be nil")
	})

	t.Run("unsupported types", func(t *testing.T) {
		spec := []struct {
			output any
			typ    string
		}{
			{output: true, typ: "bool"},
			{output: 3.14, typ: "float64"},
			{output: 123, typ: "int"},
			{output: time.Minute, typ: "int64"},
			{output: struct{}{}, typ: "struct"},
			{output: time.Now(), typ: "struct"},
		}

		for _, test := range spec {
			t.Run(test.typ, func(t *testing.T) {
				testLoadError(t, "text", test.output, fmt.Sprintf("output must be a pointer, instead got %s", test.typ))
			})
		}
	})

	t.Run("struct tags", func(t *testing.T) {
		// FIXME: go vet will not allow code with invalid struct tags, even if
		// we are intentionally testing this behavior
		// t.Run("invalid", func(t
		// *testing.T) {
		//  type test struct {
		//      Invalid string `this is not valid`
		//  }

		// 	testLoadError(t, "text", new(test), "Invalid: failed to parse struct tags: bad syntax for struct tag pair")
		// })

		t.Run("missing", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string // intentionally missing testdata struct tag
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"})
		})

		t.Run("empty", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string `testdata:""`
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"})
		})

		t.Run("dashed", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string `testdata:"-"`
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"})
		})
	})

	t.Run("empty struct", func(t *testing.T) {
		type test struct{}

		testLoadOne(t, "text", new(test), new(test))
	})

	t.Run("string", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		testLoadOne(t, "text", new(test), &test{Input: "hello world"})
	})

	t.Run("bytes", func(t *testing.T) {
		type test struct {
			Input []byte `testdata:"input.txt"`
		}

		testLoadOne(t, "text", new(test), &test{Input: []byte("hello world")})
	})

	t.Run("raw json", func(t *testing.T) {
		type test struct {
			Input json.RawMessage `testdata:"input.json"`
		}

		testLoadOne(t, "json", new(test), &test{Input: json.RawMessage("{\n  \"hello\": \"world\"\n}")})
	})

	t.Run("multiple", func(t *testing.T) {
		type test struct {
			A string `testdata:"a.txt"`
			B string `testdata:"b.txt"`
		}

		testLoadOne(t, "multiple", new(test), &test{A: "A", B: "B"})
	})

	t.Run("maps", func(t *testing.T) {
		t.Run("expand glob", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"*.txt"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string]string{
					"a.txt": "A",
					"b.txt": "B",
				},
			})
		})

		t.Run("single file", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"a.txt"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string]string{
					"a.txt": "A",
				},
			})
		})

		t.Run("bytes", func(t *testing.T) {
			type test struct {
				Multiple map[string][]byte `testdata:"a.txt"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string][]byte{
					"a.txt": []byte("A"),
				},
			})
		})

		t.Run("glob without matches", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"*.log"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string]string{},
			})
		})
	})

	t.Run("json codec", func(t *testing.T) {
		type JSONInput struct {
			Hello string `json:"hello"`
		}

		type JSONComplex struct {
			String string         `json:"example_string"`
			Number float64        `json:"example_number"`
			Bool   bool           `json:"example_boolean"`
			Null   any            `json:"example_null"`
			Array  []string       `json:"example_array"`
			Object map[string]int `json:"example_object"`
		}

		t.Run("simple", func(t *testing.T) {
			type test struct {
				Input JSONInput `testdata:"input.json"`
			}

			testLoadOne(t, "json", new(test), &test{
				Input: JSONInput{Hello: "world"},
			})
		})

		t.Run("complex", func(t *testing.T) {
			type test struct {
				Complex JSONComplex `testdata:"complex.json"`
			}

			testLoadOne(t, "json", new(test), &test{
				Complex: JSONComplex{
					String: "hello world",
					Number: 3.14,
					Bool:   true,
					Null:   nil,
					Array:  []string{"a", "b", "c", "d"},
					Object: map[string]int{"abc": 123, "def": 456},
				},
			})
		})

		t.Run("unmarshal error", func(t *testing.T) {
			type test struct {
				Input struct {
					Hello int `json:"hello"` // string "hello world" is not a valid int
				} `testdata:"input.json"`
			}

			testLoadError(t, "json", new(test), "Input: failed to unmarshal testdata/json/input.json: json: cannot unmarshal string into Go struct field .hello of type int")
		})
	})

	t.Run("unknown codec", func(t *testing.T) {
		type test struct {
			Input struct{ Hello string } `testdata:"input.unknown"`
		}

		testLoadError(t, "unknown", new(test), `Input: failed to get codec for file extension ".unknown"`)
	})

	t.Run("no outputs", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()
		mockT.EXPECT().Fatal("at least 1 output required")

		Load(mockT, filepath.Join("testdata", "text"))
	})

	t.Run("multiple outputs", func(t *testing.T) {
		type test1 struct {
			A string `testdata:"a.txt"`
		}

		type test2 struct {
			B string `testdata:"b.txt"`
		}

		testLoadMany(t, "multiple",
			[]any{new(test1), new(test2)},
			[]any{&test1{A: "A"}, &test2{B: "B"}},
		)
	})
}

func TestLoadDirs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type test struct {
			A string `testdata:"a.txt"`
			B string `testdata:"b.txt"`
		}

		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()

		var actual test
		LoadDirs(mockT, []string{"testdata/multiple-dirs/dir1", "testdata/multiple-dirs/dir2", "testdata/unknown"}, &actual)

		require.EqualValues(t, test{A: "A", B: "B"}, actual)
	})

	t.Run("missing arguments", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()
		mockT.EXPECT().Fatal("at least 1 output required")

		LoadDirs(mockT, []string{"testdata/multiple-dirs/dir1", "testdata/multiple-dirs/dir2"})
	})
}

func TestAssert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()

		Assert(mockT, "testdata/text", &test{Input: "hello world"})
	})

	t.Run("fail", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()
		mockT.EXPECT().Fatal(gomock.Any())

		Assert(mockT, "testdata/text", &test{Input: "foo bar"})
	})

	t.Run("missing arguments", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockT := NewMockT(ctrl)
		mockT.EXPECT().Helper()
		mockT.EXPECT().Fatal("at least 1 value required")

		Assert(mockT, "testdata/text")
	})

	t.Run("update", func(t *testing.T) {
		spec := []struct {
			name     string
			expected any
			fail     bool
		}{
			{
				name: "string",
				expected: &struct {
					Input string `testdata:"input.txt"`
				}{
					Input: "hello world",
				},
			},
			{
				name: "bytes",
				expected: &struct {
					Input []byte `testdata:"input.txt"`
				}{
					Input: []byte("hello world"),
				},
			},
			{
				name: "json raw",
				expected: &struct {
					Input json.RawMessage `testdata:"input.json"`
				}{
					Input: json.RawMessage(`{}`),
				},
			},
			{
				name: "json struct",
				expected: &struct {
					Input struct {
						Hello string `json:"hello"`
					} `testdata:"input.json"`
				}{
					Input: struct {
						Hello string `json:"hello"`
					}{
						Hello: "world",
					},
				},
			},
			{
				name: "map",
				expected: &struct {
					Files map[string]string `testdata:"*.txt"`
				}{
					Files: map[string]string{"a.txt": "A", "b.txt": "B"},
				},
			},
			{
				name: "unknown codec",
				expected: &struct {
					Unknown struct {
						Input int
					} `testdata:"expected.unknown"`
				}{
					Unknown: struct {
						Input int
					}{
						Input: 42,
					},
				},
				fail: true,
			},
			{
				name: "empty",
				expected: &struct {
					Output string `testdata:"output.txt"`
					Empty  string `testdata:"-"`
				}{},
			},
			{
				name: "struct tag empty",
				expected: &struct {
					Output string `testdata:"output.txt"`
					Empty  string `testdata:""`
				}{
					Output: "hello world",
				},
			},
			{
				name: "struct tag dash",
				expected: &struct {
					Output string `testdata:"output.txt"`
					Empty  string `testdata:"-"`
				}{
					Output: "hello world",
				},
			},
			// FIXME: go vet will not allow code with invalid struct tags, even
			// if we are intentionally testing this behavior
			// {
			//  name: "struct tag invalid",
			//  expected: &struct {
			//      Output string `testdata:"invalid...`
			//  }{},
			//  fail: true,
			// },
			{
				name: "struct tag missing",
				expected: &struct {
					Output string
					Empty  string
				}{},
			},
		}

		for _, test := range spec {
			t.Run(test.name, func(t *testing.T) {
				updateGolden = true
				t.Cleanup(func() { updateGolden = false })

				dir, err := os.MkdirTemp("", test.name)
				require.NoError(t, err)

				t.Cleanup(func() { os.RemoveAll(dir) })

				ctrl := gomock.NewController(t)

				mockT := NewMockT(ctrl)
				mockT.EXPECT().Helper()

				if test.fail {
					mockT.EXPECT().Fatal(gomock.Any())

					Assert(mockT, dir, test.expected)
				} else {
					Assert(mockT, dir, test.expected)

					actual := reflect.New(reflect.TypeOf(test.expected).Elem()).Interface()
					require.NoError(t, loadDir([]string{dir}, actual))
					require.EqualValues(t, test.expected, actual)
				}
			})
		}
	})
}

func testLoadOne(t *testing.T, input string, output, expected any) {
	ctrl := gomock.NewController(t)

	mockT := NewMockT(ctrl)
	mockT.EXPECT().Helper()

	Load(mockT, filepath.Join("testdata", input), output)

	require.EqualValues(t, expected, output)
}

func testLoadMany(t *testing.T, input string, output, expected []any) {
	ctrl := gomock.NewController(t)

	mockT := NewMockT(ctrl)
	mockT.EXPECT().Helper()

	Load(mockT, filepath.Join("testdata", input), output...)

	require.EqualValues(t, expected, output)
}

func testLoadError(t *testing.T, input string, output any, expectedErr string) {
	ctrl := gomock.NewController(t)

	mockT := NewMockT(ctrl)
	mockT.EXPECT().Helper()
	mockT.EXPECT().Fatal(expectedErr)

	Load(mockT, filepath.Join("testdata", input), output)
}

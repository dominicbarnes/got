package got

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		testLoadError(t, "text", nil, "[GoT] Load: output cannot be nil")
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
				testLoadError(t, "text", test.output, "[GoT] Load: output must be a pointer, but got "+test.typ)
			})
		}
	})

	t.Run("struct tags", func(t *testing.T) {
		t.Run("invalid", func(t *testing.T) {
			type test struct {
				Invalid string `this is not valid`
			}

			testLoadError(t, "text", new(test), "[GoT] Load: *got.test.Invalid: failed to parse struct tags: bad syntax for struct tag pair")
		})

		t.Run("missing", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string // intentionally missing testdata struct tag
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`,
			})
		})

		t.Run("empty", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string `testdata:""`
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`,
			})
		})

		t.Run("dashed", func(t *testing.T) {
			type test struct {
				Input   string `testdata:"input.txt"`
				Missing string `testdata:"-"`
			}

			testLoadOne(t, "text", new(test), &test{Input: "hello world"}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`,
			})
		})
	})

	t.Run("empty struct", func(t *testing.T) {
		type test struct{}

		testLoadOne(t, "text", new(test), new(test), nil)
	})

	t.Run("string", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		testLoadOne(t, "text", new(test), &test{Input: "hello world"}, []string{
			`[GoT] Load: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`,
		})
	})

	t.Run("bytes", func(t *testing.T) {
		type test struct {
			Input []byte `testdata:"input.txt"`
		}

		testLoadOne(t, "text", new(test), &test{Input: []byte("hello world")}, []string{
			`[GoT] Load: *got.test.Input: loaded file "testdata/text/input.txt" as bytes (size 11)`,
		})
	})

	t.Run("raw json", func(t *testing.T) {
		type test struct {
			Input json.RawMessage `testdata:"input.json"`
		}

		testLoadOne(t, "json", new(test), &test{Input: json.RawMessage("{\n  \"hello\": \"world\"\n}")}, []string{
			`[GoT] Load: *got.test.Input: loaded file "testdata/json/input.json" as bytes (size 22)`,
		})
	})

	t.Run("multiple", func(t *testing.T) {
		type test struct {
			A string `testdata:"a.txt"`
			B string `testdata:"b.txt"`
		}

		testLoadOne(t, "multiple", new(test), &test{A: "A", B: "B"}, []string{
			`[GoT] Load: *got.test.A: loaded file "testdata/multiple/a.txt" as string (size 1)`,
			`[GoT] Load: *got.test.B: loaded file "testdata/multiple/b.txt" as string (size 1)`,
		})
	})

	t.Run("maps", func(t *testing.T) {
		t.Run("raw json", func(t *testing.T) {
			type test struct {
				Input map[string]any `testdata:"input.json"`
			}

			testLoadOne(t, "json", new(test), &test{
				Input: map[string]any{"hello": "world"},
			}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/json/input.json" as JSON (size 22)`,
			})
		})

		t.Run("expand glob", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"*.txt,explode"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string]string{
					"a.txt": "A",
					"b.txt": "B",
				},
			}, []string{
				`[GoT] Load: *got.test.Multiple["a.txt"]: loaded file "testdata/multiple/a.txt" as string (size 1)`,
				`[GoT] Load: *got.test.Multiple["b.txt"]: loaded file "testdata/multiple/b.txt" as string (size 1)`,
			})
		})

		t.Run("single file", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"a.txt,explode"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string]string{
					"a.txt": "A",
				},
			}, []string{
				`[GoT] Load: *got.test.Multiple["a.txt"]: loaded file "testdata/multiple/a.txt" as string (size 1)`,
			})
		})

		t.Run("bytes", func(t *testing.T) {
			type test struct {
				Multiple map[string][]byte `testdata:"a.txt,explode"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: map[string][]byte{
					"a.txt": []byte("A"),
				},
			}, []string{
				`[GoT] Load: *got.test.Multiple["a.txt"]: loaded file "testdata/multiple/a.txt" as bytes (size 1)`,
			})
		})

		t.Run("glob without matches", func(t *testing.T) {
			type test struct {
				Multiple map[string]string `testdata:"*.log,explode"`
			}

			testLoadOne(t, "multiple", new(test), &test{
				Multiple: nil,
			}, []string{
				`[GoT] Load: *got.test.Multiple: no matches found`,
			})
		})

		t.Run("glob nested", func(t *testing.T) {
			type test struct {
				Input    []string          `testdata:"input.json"`
				Multiple map[string]string `testdata:"expected/*.txt,explode"`
			}

			testLoadOne(t, "multiple-nested", new(test), &test{
				Input: []string{"a", "b"},
				Multiple: map[string]string{
					"expected/a.txt": "A",
					"expected/b.txt": "B",
				},
			}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/multiple-nested/input.json" as JSON (size 10)`,
				`[GoT] Load: *got.test.Multiple["expected/a.txt"]: loaded file "testdata/multiple-nested/expected/a.txt" as string (size 1)`,
				`[GoT] Load: *got.test.Multiple["expected/b.txt"]: loaded file "testdata/multiple-nested/expected/b.txt" as string (size 1)`,
			})
		})
	})

	t.Run("json codec", func(t *testing.T) {
		type JSONInput struct {
			Hello string `json:"hello"`
		}

		type JSONComplex struct {
			String string         `json:"exampleString"`
			Number float64        `json:"exampleNumber"`
			Bool   bool           `json:"exampleBoolean"`
			Null   any            `json:"exampleNull"`
			Array  []string       `json:"exampleArray"`
			Object map[string]int `json:"exampleObject"`
		}

		t.Run("simple", func(t *testing.T) {
			type test struct {
				Input JSONInput `testdata:"input.json"`
			}

			testLoadOne(t, "json", new(test), &test{
				Input: JSONInput{Hello: "world"},
			}, []string{
				`[GoT] Load: *got.test.Input: loaded file "testdata/json/input.json" as JSON (size 22)`,
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
			}, []string{
				`[GoT] Load: *got.test.Complex: loaded file "testdata/json/complex.json" as JSON (size 227)`,
			})
		})

		t.Run("unmarshal error", func(t *testing.T) {
			type test struct {
				Input struct {
					Hello int `json:"hello"` // string "hello world" is not a valid int
				} `testdata:"input.json"`
			}

			testLoadError(t, "json", new(test), `[GoT] Load: *got.test.Input: file "testdata/json/input.json" decode error: json: cannot unmarshal string into Go struct field .hello of type int`)
		})
	})

	t.Run("unknown codec", func(t *testing.T) {
		type test struct {
			Input struct{ Hello string } `testdata:"input.unknown"`
		}

		testLoadError(t, "unknown", new(test), `[GoT] Load: *got.test.Input: failed to get codec for file extension ".unknown"`)
	})

	t.Run("no outputs", func(t *testing.T) {
		var mt mockT
		Load(&mt, filepath.Join("testdata", "text"))

		require.EqualValues(t, mockT{
			helper: true,
			failed: true,
			logs: []string{
				"[GoT] Load: at least 1 output required",
			},
		}, mt)
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
			[]string{
				`[GoT] Load: *got.test1.A: loaded file "testdata/multiple/a.txt" as string (size 1)`,
				`[GoT] Load: *got.test2.B: loaded file "testdata/multiple/b.txt" as string (size 1)`,
			},
		)
	})
}

func TestLoadDirs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type test struct {
			A string `testdata:"a.txt"`
			B string `testdata:"b.txt"`
		}

		var mt mockT
		var actual test
		LoadDirs(&mt, []string{"testdata/multiple-dirs/dir1", "testdata/multiple-dirs/dir2", "testdata/unknown"}, &actual)

		require.EqualValues(t, test{A: "A", B: "B"}, actual)

		require.EqualValues(t, mockT{
			helper: true,
			logs: []string{
				`[GoT] Load: *got.test.A: loaded file "testdata/multiple-dirs/dir1/a.txt" as string (size 1)`,
				`[GoT] Load: *got.test.A: skipped: file "testdata/multiple-dirs/dir2/a.txt" not found`,
				`[GoT] Load: *got.test.A: skipped: file "testdata/unknown/a.txt" not found`,
				`[GoT] Load: *got.test.B: skipped: file "testdata/multiple-dirs/dir1/b.txt" not found`,
				`[GoT] Load: *got.test.B: loaded file "testdata/multiple-dirs/dir2/b.txt" as string (size 1)`,
				`[GoT] Load: *got.test.B: skipped: file "testdata/unknown/b.txt" not found`,
			},
		}, mt)
	})

	t.Run("missing arguments", func(t *testing.T) {
		var mt mockT
		LoadDirs(&mt, []string{"testdata/multiple-dirs/dir1", "testdata/multiple-dirs/dir2"})

		require.EqualValues(t, mockT{
			helper: true,
			failed: true,
			logs: []string{
				"[GoT] LoadDirs: at least 1 output required",
			},
		}, mt)
	})
}

func TestAssert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		var mt mockT
		Assert(&mt, "testdata/text", &test{Input: "hello world"})

		require.EqualValues(t, mockT{
			helper: true,
			logs: []string{
				`[GoT] Assert: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`,
			},
		}, mt)
	})

	t.Run("fail", func(t *testing.T) {
		type test struct {
			Input string `testdata:"input.txt"`
		}

		var mt mockT
		Assert(&mt, "testdata/text", &test{Input: "foo bar"})

		require.True(t, mt.helper)
		require.True(t, mt.failed)
		require.Len(t, mt.logs, 2)
		require.Equal(t, `[GoT] Assert: *got.test.Input: loaded file "testdata/text/input.txt" as string (size 11)`, mt.logs[0])
		require.True(t, strings.HasPrefix(mt.logs[1], "[GoT] Assert: test of *got.test failed:"))
	})

	t.Run("missing arguments", func(t *testing.T) {
		var mt mockT
		Assert(&mt, "testdata/text")

		require.EqualValues(t, mockT{
			helper: true,
			failed: true,
			logs: []string{
				"[GoT] Assert: at least 1 value required",
			},
		}, mt)
	})

	t.Run("update", func(t *testing.T) {
		spec := []struct {
			name     string
			expected any
			fail     bool
			logs     []string
		}{
			{
				name: "string",
				expected: &struct {
					Input string `testdata:"input.txt"`
				}{
					Input: "hello world",
				},
				logs: []string{
					`[GoT] Assert: .Input: saved file "<tmp>/input.txt" (size 11)`,
				},
			},
			{
				name: "bytes",
				expected: &struct {
					Input []byte `testdata:"input.txt"`
				}{
					Input: []byte("hello world"),
				},
				logs: []string{
					`[GoT] Assert: .Input: saved file "<tmp>/input.txt" (size 11)`,
				},
			},
			{
				name: "json raw",
				expected: &struct {
					Input json.RawMessage `testdata:"input.json"`
				}{
					Input: json.RawMessage(`{}`),
				},
				logs: []string{
					`[GoT] Assert: .Input: saved file "<tmp>/input.json" (size 2)`,
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
				logs: []string{
					`[GoT] Assert: .Input: saved file "<tmp>/input.json" (size 22)`,
				},
			},
			{
				name: "map json",
				expected: &struct {
					Input map[string]string `testdata:"input.json"`
				}{
					Input: map[string]string{"hello": "world"},
				},
				logs: []string{
					`[GoT] Assert: .Input: saved file "<tmp>/input.json" (size 22)`,
				},
			},
			{
				name: "map explode",
				expected: &struct {
					Files map[string]string `testdata:"*.txt,explode"`
				}{
					Files: map[string]string{"a.txt": "A", "b.txt": "B"},
				},
				logs: []string{
					`[GoT] Assert: .Files: saved file "<tmp>/a.txt" (size 1)`,
					`[GoT] Assert: .Files: saved file "<tmp>/b.txt" (size 1)`,
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
				logs: []string{
					`[GoT] Assert: .Output: removed file "<tmp>/output.txt": empty`,
				},
			},
			{
				name: "struct tag empty",
				expected: &struct {
					Output string `testdata:"output.txt"`
					Empty  string `testdata:""`
				}{
					Output: "hello world",
				},
				logs: []string{
					`[GoT] Assert: .Output: saved file "<tmp>/output.txt" (size 11)`,
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
				logs: []string{
					`[GoT] Assert: .Output: saved file "<tmp>/output.txt" (size 11)`,
				},
			},
			{
				name: "struct tag invalid",
				expected: &struct {
					Output string `testdata:"invalid...`
				}{},
				fail: true,
			},
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

				var mt mockT

				if test.fail {
					Assert(&mt, dir, test.expected)

					require.True(t, mt.failed)
					require.Len(t, mt.logs, 1)
					require.True(t, strings.HasPrefix(mt.logs[0], "[GoT] Assert:"))
				} else {
					Assert(&mt, dir, test.expected)

					actual := reflect.New(reflect.TypeOf(test.expected).Elem()).Interface()
					Load(t, dir, actual)
					require.EqualValues(t, test.expected, actual)

					// strip the temp directory name from logs, as it makes the
					// assertion non-deterministic
					for i := range mt.logs {
						mt.logs[i] = strings.ReplaceAll(mt.logs[i], dir, "<tmp>")
					}

					require.False(t, mt.failed)
					require.EqualValues(t, test.logs, mt.logs)
				}

				require.True(t, mt.helper)
			})
		}
	})
}

func testLoadOne(t *testing.T, input string, output, expected any, logs []string) {
	t.Helper()

	dir := filepath.Join("testdata", input)

	var mt mockT
	Load(&mt, dir, output)

	require.EqualValues(t, expected, output)

	require.EqualValues(t, mockT{
		helper: true,
		logs:   logs,
	}, mt)
}

func testLoadMany(t *testing.T, input string, output, expected []any, logs []string) {
	t.Helper()

	var mt mockT
	Load(&mt, filepath.Join("testdata", input), output...)

	require.EqualValues(t, expected, output)

	require.EqualValues(t, mockT{
		helper: true,
		logs:   logs,
	}, mt)
}

func testLoadError(t *testing.T, input string, output any, expectedErr string) {
	t.Helper()

	var mt mockT
	Load(&mt, filepath.Join("testdata", input), output)

	require.EqualValues(t, mockT{
		helper: true,
		failed: true,
		logs:   []string{expectedErr},
	}, mt)
}

package got_test

import (
	"encoding/json"
	reflect "reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	. "github.com/dominicbarnes/got"
)

func TestTestData(t *testing.T) {
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
		TestData(mockt, "testdata/text", &actual)
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
		TestData(mockt, "testdata/text", &actual)
		expected := TestCase{[]byte("hello world")}
		require.EqualValues(t, expected, actual)
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
			TestData(mockt, "testdata/json", &actual)
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
			TestData(mockt, "testdata/json", &actual)
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
			TestData(mockt, "testdata/json", &actual)
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
			TestData(mockt, "testdata/json", &actual)
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
			mockt.EXPECT().Fatalf("%s: failed to parse %s as JSON", "Input", "testdata/json/invalid.json")

			type TestCase struct {
				Input map[string]interface{} `testdata:"invalid.json"`
			}
			var actual TestCase
			TestData(mockt, "testdata/json", &actual)
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
		TestData(mockt, "testdata/multiple", &actual)
		expected := TestCase{"A", []byte("B")}
		require.EqualValues(t, expected, actual)
	})

	t.Run("nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Fatal("output cannot be nil")

		TestData(mockt, "testdata/text", nil)
	})

	t.Run("non-pointer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Fatalf("output must be pointer value, instead got %s", reflect.Struct)

		TestData(mockt, "testdata/text", struct{}{})
	})

	t.Run("missing file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockt := NewMockTestingT(ctrl)
		mockt.EXPECT().Helper()
		mockt.EXPECT().Logf("%s: reading file %s", "Input", "testdata/text/does-not-exist")
		mockt.EXPECT().Fatalf("%s: failed to read file: %s", "Input", "open testdata/text/does-not-exist: no such file or directory")

		type TestCase struct {
			Input string `testdata:"does-not-exist"`
		}
		var actual TestCase
		TestData(mockt, "testdata/text", &actual)
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
		TestData(mockt, "testdata/text", &actual)
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
		TestData(mockt, "testdata/text", &actual)
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
		TestData(mockt, "testdata/text", &actual)
		require.Empty(t, actual.Missing)
	})
}

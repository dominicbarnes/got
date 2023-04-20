package got

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListSubDirs(t *testing.T) {
	spec := []struct {
		name     string
		input    string
		expected []string
		fail     bool
	}{
		{
			name:     "success",
			input:    "testdata",
			expected: []string{"json", "multiple", "multiple-dirs", "text", "yaml"},
		},
		{
			name:  "fail",
			input: "does-not-exist",
			fail:  true,
		},
	}

	for _, test := range spec {
		t.Run(test.name, func(t *testing.T) {
			actual, err := listSubDirs(test.input)
			if test.fail {
				require.Nil(t, actual)
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, test.expected, actual)
			}
		})
	}
}

// func ExampleListSubDirs() {
// 	t := new(testing.T) // not necessary in normal test code

// 	for _, testName := range ListSubDirs(t, "testdata") {
// 		t.Run(testName, func(t *testing.T) {
// 			testDir := filepath.Join("testdata", testName)

// 			type TestCase struct {
// 				Input    string `testdata:"input.txt"`
// 				Expected string `testdata:"expected.txt"`
// 			}

// 			var test TestCase
// 			LoadTestData(t, testDir, &test)

// 			actual := strings.ToUpper(test.Input)
// 			if actual != test.Expected {
// 				t.Fatalf("actual value '%s' did not match expected value '%s'", actual, test.Expected)
// 			}
// 		})
// 	}
// }

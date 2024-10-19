package got

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RunTestSuite is a helper for running a common test suite. The Input type
// parameter determines what will be passed to Load, while the Output type
// parameter determines what will be passed to Assert. The passed func accepts
// the loaded Input and returns the Output directly.
//
// For more advanced cases like using TestSuite.SharedDir or situations where
// multiple types are passed to Load, the TestSuite should be used directly.
func RunTestSuite[Input any, Output any](t tester, dir string, fn func(t *testing.T, tc TestCase, test Input) Output) {
	t.Helper()

	suite := TestSuite{
		Dir: dir,
		TestFunc: func(t *testing.T, tc TestCase) {
			t.Helper()

			var input Input
			tc.Load(t, &input)

			output := fn(t, tc, input)

			Assert(t, tc.Dir, &output)
		},
	}

	suite.Run(t)
}

// TestCase is used to wrap up test metadata.
type TestCase struct {
	// Name is the base name for this test case (excluding any parent names).
	Name string

	// Skip indicates that the test should be skipped. This is indicated to the
	// TestSuite by having a directory name with a ".skip" suffix.
	Skip bool

	// Only indicates that every other test should be skipped. This is indicated
	// to the TestSuite by having a directory name with a ".only" suffix.
	Only bool

	// Dir is the base directory for this test case.
	Dir string

	// SharedDir is an alternate location for test case configuration, if the
	// suite has been configured to search for this.
	SharedDir string
}

// Load is a helper for loading testdata for this test case, factoring in a
// SharedDir automatically if applicable.
func (c TestCase) Load(t tester, values ...any) {
	if c.SharedDir != "" {
		LoadDirs(t, []string{c.SharedDir, c.Dir}, values...)
	} else {
		Load(t, c.Dir, values...)
	}
}

// Assert is a helper for checking and/or saving testdata for this test case.
func (c TestCase) Assert(t tester, values ...any) {
	Assert(t, c.Dir, values...)
}

// TestSuite defines a collection of tests backed by directories/files on disk.
type TestSuite struct {
	// Dir is the location of your test suite.
	Dir string

	// SharedDir adds an additional directory to search for test cases.
	//
	// When set, this directory is scanned first and is treated as the primary
	// test suite. For each sub-directory, a corresponding sub-directory must
	// also be found in Dir, or that sub-test will fail. Any sub-directories
	// found in Dir will be added to the test suite.
	//
	// This allows a test suite to be defined for a common interface, which can
	// then be run for all implementations of that interface, while allowing
	// each implementation to inculde their own additional test cases and
	// configuration.
	SharedDir string

	// TestFunc is the hook for running test code, it will be called for each
	// found test case in the suite.
	TestFunc func(*testing.T, TestCase)
}

// Run loads and executes the test suite.
func (s *TestSuite) Run(t tester) {
	t.Helper()

	hasOnly := false
	testCases := make(map[string]TestCase)

	for _, testDir := range listSubDirs(t, s.Dir) {
		name, skip, only := parseTestDir(testDir)
		if only {
			hasOnly = true
		}

		testCase := TestCase{
			Name: name,
			Skip: skip,
			Only: only,
			Dir:  filepath.Join(s.Dir, testDir),
		}

		testCases[name] = testCase
	}

	for _, testDir := range listSubDirs(t, s.SharedDir) {
		name, skip, only := parseTestDir(testDir)
		if only {
			hasOnly = true
		}

		sharedDir := filepath.Join(s.SharedDir, testDir)

		if tc, ok := testCases[name]; !ok {
			testCases[name] = TestCase{
				Name:      name,
				Skip:      skip,
				Only:      only,
				Dir:       filepath.Join(s.Dir, testDir),
				SharedDir: sharedDir,
			}
		} else {
			tc.SharedDir = sharedDir

			testCases[name] = tc
		}
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Helper()

			if hasOnly && !testCase.Only {
				t.Skip("skipping test because it is excluded by only")
			} else if testCase.Skip {
				t.Skip("skipping test because it is has been marked")
			}

			s.TestFunc(t, testCase)
		})
	}
}

func listSubDirs(t tester, dir string) []string {
	t.Helper()

	if dir == "" {
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir %s: %s", dir, err)
	}

	var list []string
	for _, file := range files {
		if file.IsDir() {
			list = append(list, file.Name())
		}
	}

	return list
}

// returns name, skip, only.
func parseTestDir(input string) (string, bool, bool) {
	switch {
	case strings.HasSuffix(input, ".skip"):
		return strings.TrimSuffix(input, ".skip"), true, false
	case strings.HasSuffix(input, ".only"):
		return strings.TrimSuffix(input, ".only"), false, true
	default:
		return input, false, false
	}
}

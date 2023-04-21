package got

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// TestCase is used to wrap up test metadata.
type TestCase struct {
	// Name is the base name for this test case (excluding any parent names).
	Name string

	// Dir is the base directory for this test case.
	Dir string

	// SharedDir is an alternate location for test case configuration, if the
	// suite has been configured to search for this.
	SharedDir string
}

// Load is a helper for loading testdata for this test case, factoring in a
// SharedDir automatically if applicable.
func (c TestCase) Load(t *testing.T, values ...any) {
	ctx := context.Background()

	if c.SharedDir != "" {
		if err := LoadDirs(ctx, []string{c.Dir, c.SharedDir}, values...); err != nil {
			t.Fatalf("failed to load testdata: %s", err)
		}
	} else {
		if err := Load(ctx, c.Dir, values...); err != nil {
			t.Fatalf("failed to load testdata: %s", err)
		}
	}
}

// TestSuite defines a collection of tests backed by directories/files on disk.
type TestSuite struct {
	// Dir is the location of your test suite.
	Dir string

	// SharedDir adds an additional directory to search for test cases.
	//
	// When set, this directory is scanned first and is treated as the primary
	// test suite. For each sub-directory, a corresponding sub-directory must also
	// be found in Dir, or that sub-test will fail. Any sub-directories found in
	// Dir will be added to the test suite.
	//
	// This allows a test suite to be defined for a common interface, which can
	// then be run for all implementations of that interface, while allowing each
	// implementation to inculde their own additional test cases and
	// configuration.
	SharedDir string

	// TestFunc is the hook for running test code, it will be called for each
	// found test case in the suite.
	TestFunc func(*testing.T, TestCase)
}

// Run loads and executes the test suite.
func (s *TestSuite) Run(t *testing.T) {
	t.Helper()

	testCases := make(map[string]struct{})

	for _, testName := range listSubDirs(t, s.Dir) {
		testCases[testName] = struct{}{}
	}

	for _, testName := range listSubDirs(t, s.SharedDir) {
		testCases[testName] = struct{}{}
	}

	for testName := range testCases {
		testCase := TestCase{
			Name: testName,
			Dir:  filepath.Join(s.Dir, testName),
		}

		if s.SharedDir != "" {
			testCase.SharedDir = filepath.Join(s.SharedDir, testName)
		}

		t.Run(testCase.Name, func(t *testing.T) {
			s.TestFunc(t, testCase)
		})
	}
}

func listSubDirs(t *testing.T, dir string) []string {
	if dir == "" {
		return nil
	}

	files, err := ioutil.ReadDir(dir)
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

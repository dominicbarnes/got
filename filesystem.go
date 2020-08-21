package got

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// RunSubTests is a helper for executing a sub-test for each directory within a
// base directory. This is useful in cases where each sub-folder is a separate
// test case.
func RunSubTests(t *testing.T, dir string, sub func(*testing.T, string, string)) {
	t.Helper()
	for _, testName := range ListSubDirs(t, dir) {
		testName := testName
		testDir := filepath.Join(dir, testName)
		t.Run(testName, func(t *testing.T) {
			sub(t, testName, testDir)
		})
	}
}

// ListSubDirs finds the only the nested directories within the input dir. This
// is useful for quickly grabbing a list of subtests represented by
// subdirectories in testdata or similar.
func ListSubDirs(t *testing.T, dir string) []string {
	t.Helper()
	dirs, err := listSubDirs(dir)
	if err != nil {
		t.Fatalf("failed to read testdata dir: %s", err.Error())
	}
	return dirs
}

func listSubDirs(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var list []string
	for _, file := range files {
		if file.IsDir() {
			list = append(list, file.Name())
		}
	}
	return list, nil
}

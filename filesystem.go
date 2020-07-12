package got

import (
	"io/ioutil"
	"testing"
)

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

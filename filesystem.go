package got

import "io/ioutil"

// GetDirs finds the only the nested directories within the input dir. This is
// useful for quickly grabbing a list of subtests represented by subdirectories
// in testdata/ or similar.
func GetDirs(t TestingT, dir string) []string {
	t.Helper()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal("failed to read testdata dir")
	}

	var dirs []string
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

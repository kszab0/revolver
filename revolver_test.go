package revolver

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Cannot create temp dir: %v", err)
	}
	clean := func() {
		os.RemoveAll(dir)
	}
	return dir, clean
}

func createTempNestedDirs(t *testing.T, dir string) string {
	dirs := filepath.Join(dir, "a", "b", "c", "d")
	if err := os.MkdirAll(dirs, 0700); err != nil {
		t.Fatalf("Cannot create nested dirs: %v", err)
	}
	return dirs
}

func createTempFile(t *testing.T, dir, name string) string {
	file, err := ioutil.TempFile(dir, name)
	if err != nil {
		t.Fatalf("Cannot create temp file: %v", err)
	}
	rel, err := filepath.Rel(dir, file.Name())
	if err != nil {
		t.Fatalf("Cannot get relative path: %v", err)
	}
	return rel
}

func writeFile(t *testing.T, name string) {
	f, err := os.OpenFile(name, os.O_WRONLY, 0755)
	if err != nil {
		t.Fatalf("Cannot open file: %v", err)
	}
	defer f.Close()

	time.Sleep(5 * time.Millisecond)

	if _, err := f.WriteString("change content"); err != nil {
		t.Fatalf("Cannot write to file: %v", err)
	}
}

func contains(arr []string, el string) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}

func equals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, el := range a {
		if !contains(b, el) {
			return false
		}
	}
	return true
}

func relative(t *testing.T, dir, name string) string {
	rel, err := filepath.Rel(dir, name)
	if err != nil {
		t.Errorf("Cannot get relative path: %v", err)
	}
	return rel
}

func TestDetect(t *testing.T) {

	type testCase func(t *testing.T, dir string) (expected []string, detect DetectFunc)

	for name, tc := range map[string]testCase{
		"empty dir": func(t *testing.T, dir string) ([]string, DetectFunc) {
			detect := Detect(dir, nil)
			detect()

			expected := []string{}
			return expected, detect
		},
		"flat dir no change": func(t *testing.T, dir string) ([]string, DetectFunc) {
			createTempFile(t, dir, "")

			detect := Detect(dir, nil)
			detect()

			expected := []string{}
			return expected, detect
		},
		"flat dir add file ": func(t *testing.T, dir string) ([]string, DetectFunc) {
			createTempFile(t, dir, "")

			detect := Detect(dir, nil)
			detect()

			file := createTempFile(t, dir, "")

			expected := []string{file}
			return expected, detect

		},
		"flat dir change file": func(t *testing.T, dir string) ([]string, DetectFunc) {
			file := createTempFile(t, dir, "")

			detect := Detect(dir, nil)
			detect()

			writeFile(t, filepath.Join(dir, file))

			expected := []string{file}
			return expected, detect
		},
		"flat dir delete file": func(t *testing.T, dir string) ([]string, DetectFunc) {
			file := createTempFile(t, dir, "")

			detect := Detect(dir, nil)
			detect()

			os.Remove(filepath.Join(dir, file))

			expected := []string{file}
			return expected, detect
		},
		"nested dir no change": func(t *testing.T, dir string) ([]string, DetectFunc) {
			createTempNestedDirs(t, dir)

			detect := Detect(dir, nil)
			detect()

			expected := []string{}
			return expected, detect
		},
		"nested dir new file": func(t *testing.T, dir string) ([]string, DetectFunc) {
			dirs := createTempNestedDirs(t, dir)

			detect := Detect(dir, nil)
			detect()

			file := createTempFile(t, dirs, "")

			expected := []string{relative(t, dir, filepath.Join(dirs, file))}
			return expected, detect
		},
		"nested dir change file": func(t *testing.T, dir string) ([]string, DetectFunc) {
			dirs := createTempNestedDirs(t, dir)

			detect := Detect(dir, nil)
			detect()

			file := createTempFile(t, dirs, "")

			writeFile(t, filepath.Join(dirs, file))

			expected := []string{relative(t, dir, filepath.Join(dirs, file))}
			return expected, detect
		},
		"nested dir delete file": func(t *testing.T, dir string) ([]string, DetectFunc) {
			dirs := createTempNestedDirs(t, dir)
			file := createTempFile(t, dirs, "")

			detect := Detect(dir, nil)
			detect()

            df := filepath.Join(dirs, file)
            os.Remove(df)

			expected := []string{relative(t, dir, df)}
			return expected, detect
		},
		"skip dir": func(t *testing.T, dir string) ([]string, DetectFunc) {
			nested := filepath.Join("a", "b", "c", "d")
			dirs := filepath.Join(dir, nested)
			if err := os.MkdirAll(dirs, 0700); err != nil {
				t.Fatalf("Cannot create nested dirs: %v", err)
			}
			createTempFile(t, dirs, "")

			excludeDirs := []string{nested}

			detect := Detect(dir, excludeDirs)

			expected := []string{}
			return expected, detect
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir, teardown := createTempDir(t)
			defer teardown()

			expected, detect := tc(t, dir)

			time.Sleep(5 * time.Millisecond)

			changed := detect()

			if !equals(expected, changed) {
				t.Errorf("Changed dirs should be: %v; got: %v", expected, changed)
			}
		})
	}
}

func TestRun(t *testing.T) {
    buildCmd := func(command string, args ...string) func(t *testing.T) []BuildFunc {
        return func(t *testing.T) []BuildFunc {
            return []BuildFunc{BuildCommand(command, args...)}
        }
    }
    buildErr := func(t *testing.T) []BuildFunc {
        return []BuildFunc{
            BuildCommand("exit", "1"),
            func() error {
                t.Errorf("BuildFunc should not execute")
                return nil
            },
        }
    }

    runCmd :=func(command string, args ...string) func(t *testing.T) RunFunc {
        return func(t *testing.T) RunFunc {
            return RunCommand(command, args...)
        }
    }
    runErr := func(t *testing.T) RunFunc {
        return func() (func(), error) {
            t.Errorf("RunFunc should not execute")
            return func(){}, nil
        }
    }

    type testCase struct {
        build func(*testing.T) []BuildFunc
        run func(*testing.T) RunFunc
        err bool
    }
    for name, tc := range map[string]testCase {
        "ok": {
            build: buildCmd("echo", "ok"),
            run: runCmd("tail", ""),
            err: false,
        },
        "build error": {
            build: buildCmd("exit", "1"),
            run: runErr,
            err: true,
        },
        "build chain error": {
            build: buildErr,
            run: runErr,
            err: true,
        },
        "empty run": {
            build: buildCmd("echo", "empty run"),
            err: false,
        },
        "run build": {
            run: runCmd("exit", "1"),
            err: true,
        },
    } {
        t.Run(name, func(t *testing.T) {

            var build []BuildFunc
            var run RunFunc

            if tc.build != nil {
                build = tc.build(t)
            }
            if tc.run != nil {
                run = tc.run(t)
            }

            stop, err := Run(build, run)
			if err != nil {
				if !tc.err {
					t.Errorf("Run() err = %v; wanted no errors", err)
				}
				return
			}
            if tc.err {
                t.Errorf("Run() err should not be nil")
            }
            stop()
        })
    }
}

func TestFilter(t *testing.T) {
    type testCase struct {
        files, includes, excludes []string
        changed bool
    }
    for name, tc := range map[string]testCase {
        "empty": {
            files: []string{},
            includes: []string{},
            excludes: []string{},
            changed: false,
        },
        "include all, no excludes": {
            files: []string{"file.go", "file_test.go"},
            includes: []string{"*"},
            excludes: []string{},
            changed: true,
        },
        "exclude included": {
            files: []string{"file.go", "file_test.go"},
            includes: []string{"*.go"},
            excludes: []string{"*.go"},
            changed: false,
        },
        "exclude _test.go files": {
            files: []string{"file.go", "file_test.go"},
            includes: []string{"*"},
            excludes: []string{"*_test.go"},
            changed: true,
        },
    } {
        t.Run(name, func(t *testing.T) {
            changed := Filter(tc.files, tc.includes, tc.excludes)()
            if changed != tc.changed {
                t.Errorf("Filter() should return %v; got: %v", tc.changed, changed)
            }
        })
    }
}

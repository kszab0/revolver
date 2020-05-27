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

	runCmd := func(command string, args ...string) func(t *testing.T) RunFunc {
		return func(t *testing.T) RunFunc {
			return RunCommand(command, args...)
		}
	}
	runErr := func(t *testing.T) RunFunc {
		return func() (func(), error) {
			t.Errorf("RunFunc should not execute")
			return func() {}, nil
		}
	}

	type testCase struct {
		build func(*testing.T) []BuildFunc
		run   func(*testing.T) RunFunc
		err   bool
	}
	for name, tc := range map[string]testCase{
		"ok": {
			build: buildCmd("echo", "ok"),
			run:   runCmd("tail", ""),
			err:   false,
		},
		"build error": {
			build: buildCmd("exit", "1"),
			run:   runErr,
			err:   true,
		},
		"build chain error": {
			build: buildErr,
			run:   runErr,
			err:   true,
		},
		"empty run": {
			build: buildCmd("echo", "empty run"),
			err:   false,
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

			if stop != nil {
				stop()
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type testCase struct {
		files, includes, excludes []string
		changed                   bool
	}
	for name, tc := range map[string]testCase{
		"empty": {
			files:    []string{},
			includes: []string{},
			excludes: []string{},
			changed:  false,
		},
		"include all, no excludes": {
			files:    []string{"file.go", "file_test.go"},
			includes: []string{"*"},
			excludes: []string{},
			changed:  true,
		},
		"exclude included": {
			files:    []string{"file.go", "file_test.go"},
			includes: []string{"*.go"},
			excludes: []string{"*.go"},
			changed:  false,
		},
		"exclude _test.go files": {
			files:    []string{"file.go", "file_test.go"},
			includes: []string{"*"},
			excludes: []string{"*_test.go"},
			changed:  true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := Filter(tc.includes, tc.excludes)(tc.files)
			if changed != tc.changed {
				t.Errorf("Filter() should return %v; got: %v", tc.changed, changed)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	equals := func(a, b Config) bool {
		if a.Dir != b.Dir ||
			len(a.ExcludeDirs) != len(b.ExcludeDirs) ||
			a.Interval != b.Interval ||
			len(a.Actions) != len(b.Actions) {
			return false
		}
		for i := 0; i < len(a.Actions); i++ {
			actionA := a.Actions[i]
			actionB := b.Actions[i]

			if actionA.Name != actionB.Name ||
				len(actionA.Patterns) != len(actionB.Patterns) ||
				len(actionA.ExcludePatterns) != len(actionB.ExcludePatterns) ||
				len(actionA.BuildCommands) != len(actionB.BuildCommands) ||
				actionA.RunCommand != actionB.RunCommand {
				return false
			}
			for i := 0; i < len(actionA.Patterns); i++ {
				if actionA.Patterns[i] != actionB.Patterns[i] {
					return false
				}
			}
			for i := 0; i < len(actionB.ExcludePatterns); i++ {
				if actionA.ExcludePatterns[i] != actionB.ExcludePatterns[i] {
					return false
				}
			}
			for i := 0; i < len(actionA.BuildCommands); i++ {
				if actionA.BuildCommands[i] != actionB.BuildCommands[i] {
					return false
				}
			}
		}
		for i := 0; i < len(a.ExcludeDirs); i++ {
			if a.ExcludeDirs[i] != b.ExcludeDirs[i] {
				return false
			}
		}

		return true
	}

	type testCase struct {
		content string
		config  Config
		err     bool
	}
	for name, tc := range map[string]testCase{
		"empty": {
			content: ``,
			err:     true,
		},
		"config: no action": {
			content: `action:`,
			err:     true,
		},
		"config: maleformed action": {
			content: `action: "maleformed"`,
			err:     true,
		},
		"config: no command": {
			content: `action:
  - name: "action"`,
			err: true,
		},
		"config: minimal": {
			content: `action:
  - build: ["echo ok"]`,
			config: Config{
				Dir:      ".",
				Interval: 500 * time.Millisecond,
				Actions: []Action{
					{
						Patterns:      []string{"**/*"},
						BuildCommands: []string{"echo ok"},
					},
				},
			},
			err: false,
		},
		"config: full": {
			content: `dir: "dir"
excludeDir: ["exclude"]
interval: 1s
action:
  - name: "action"
    pattern: ["**/*.go"]
    exclude: ["**/*_test.go"]
    build: ["echo build"]
    run: "echo run"`,
			config: Config{
				Dir:         "dir",
				ExcludeDirs: []string{"exclude"},
				Interval:    1 * time.Second,
				Actions: []Action{
					{
						Name:            "action",
						Patterns:        []string{"**/*.go"},
						ExcludePatterns: []string{"**/*_test.go"},
						BuildCommands:   []string{"echo build"},
						RunCommand:      "echo run",
					},
				},
			},
			err: false,
		},
		"config: without arrays": {
			content: `excludeDir: "exclude"
action:
  - pattern: "**/*.go"
    exclude: "**/*_test.go"
    build: "echo build"`,
			config: Config{
				Dir:         ".",
				ExcludeDirs: []string{"exclude"},
				Interval:    500 * time.Millisecond,
				Actions: []Action{
					{
						Patterns:        []string{"**/*.go"},
						ExcludePatterns: []string{"**/*_test.go"},
						BuildCommands:   []string{"echo build"},
					},
				},
			},
			err: false,
		},
		"simple: minimal": {
			content: `build: ["echo ok"]`,
			config: Config{
				Dir:      ".",
				Interval: 500 * time.Millisecond,
				Actions: []Action{
					{
						Patterns:      []string{"**/*"},
						BuildCommands: []string{"echo ok"},
					},
				},
			},
			err: false,
		},
		"simple: full": {
			content: `dir: "dir"
excludeDir: ["exclude"]
interval: 1s
pattern: ["**/*.go"]
exclude: ["**/*_test.go"]
build: ["echo build"]
run: "echo run"`,
			config: Config{
				Dir:         "dir",
				ExcludeDirs: []string{"exclude"},
				Interval:    1 * time.Second,
				Actions: []Action{
					{
						Patterns:        []string{"**/*.go"},
						ExcludePatterns: []string{"**/*_test.go"},
						BuildCommands:   []string{"echo build"},
						RunCommand:      "echo run",
					},
				},
			},
			err: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			config, err := ParseConfig([]byte(tc.content))
			if err != nil {
				if !tc.err {
					t.Errorf("ParseConfig() err should be nil; got: %v", err)
				}
				return
			}
			if tc.err {
				t.Errorf("ParseConfig() err should be %v; got: nil", err)
				return
			}

			if !equals(*config, tc.config) {
				t.Errorf("ParseConfig() should be %v; got: %v", tc.config, config)
			}
		})
	}
}

func TestParseActions(t *testing.T) {
	type testAction struct {
		id         string
		name       string
		buildFuncs int
		runFunc    bool
	}
	equals := func(a action, b testAction) bool {
		if a.ID != b.id ||
			a.Name != b.name ||
			len(a.BuildFuncs) != b.buildFuncs {
			return false
		}
		if b.runFunc {
			if a.RunFunc == nil {
				return false
			}
		} else {
			if a.RunFunc != nil {
				return false
			}
		}
		return true
	}
	type testCase struct {
		actions  []Action
		expected []testAction
	}
	for name, tc := range map[string]testCase{
		"without name": {
			actions: []Action{
				{},
			},
			expected: []testAction{
				{id: "1"},
			},
		},
		"multiple without name": {
			actions: []Action{
				{}, {}, {},
			},
			expected: []testAction{
				{id: "1"}, {id: "2"}, {id: "3"},
			},
		},
		"with name": {
			actions: []Action{
				{Name: "name"},
			},
			expected: []testAction{
				{id: "name", name: "name"},
			},
		},
		"duplicate with name": {
			actions: []Action{
				{Name: "name"},
				{Name: "name"},
			},
			expected: []testAction{
				{id: "name", name: "name"},
				{id: "name-2", name: "name"},
			},
		},
		"mixed with and without name": {
			actions: []Action{
				{Name: "name"},
				{},
				{Name: "name"},
				{Name: "asdf"},
				{},
			},
			expected: []testAction{
				{id: "name", name: "name"},
				{id: "2"},
				{id: "name-3", name: "name"},
				{id: "asdf", name: "asdf"},
				{id: "5"},
			},
		},
		"build funcs": {
			actions: []Action{
				{BuildCommands: []string{"echo asdf", "echo asdf"}},
			},
			expected: []testAction{
				{id: "1", buildFuncs: 2},
			},
		},
		"run func": {
			actions: []Action{
				{RunCommand: "echo asdf"},
			},
			expected: []testAction{
				{id: "1", runFunc: true},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actions := parseActions(tc.actions)
			if len(actions) != len(tc.expected) {
				t.Errorf("Actions length should be: %v; got: %v", len(tc.expected), len(actions))
				return
			}
			for i := 0; i < len(actions); i++ {
				if !equals(actions[i], tc.expected[i]) {
					t.Errorf("Action should be: %v; got: %v", actions[i], tc.expected[i])
				}
			}
		})
	}
}

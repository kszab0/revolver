package revolver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
)

func matchPatterns(patterns []string, name string) bool {
	for _, pattern := range patterns {
		if ok, _ := doublestar.PathMatch(pattern, name); ok {
			return true
		}
	}
	return false
}

// DetectFunc detects changes in a filesystem and returns the changed files.
type DetectFunc func() []string

// Detect returns a DetectFunc that will walk the filesystem from the given dir
// recursively, skipping the excludeDirs and return the changed files.
func Detect(dir string, excludeDirs []string) DetectFunc {
	prev := make(map[string]os.FileInfo)

	return func() []string {
		changed := []string{}
		curr := make(map[string]os.FileInfo)

		filepath.Walk(dir, func(path string, file os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			name, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			if file.IsDir() {
				if matchPatterns(excludeDirs, name) {
					return filepath.SkipDir
				}
				return nil
			}

			curr[name] = file

			prevFile, ok := prev[name]
			if !ok {
				changed = append(changed, name)
				return nil
			}
			if prevFile.ModTime() != file.ModTime() {
				changed = append(changed, name)
				return nil
			}

			return nil
		})

		for name := range prev {
			if _, ok := curr[name]; !ok {
				changed = append(changed, name)
			}
		}

		prev = curr
		return changed
	}
}

// BuildFunc is a function that is executed before a RunFunc
type BuildFunc func() error

// BuildCommand returns a BuildFunc that can execute a command with arguments.
func BuildCommand(command string, args ...string) BuildFunc {
	return func() error {
		cmd := exec.Command(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("Error running: \"%s %s\": %w", command, strings.Join(args, ""), err)
		}
		return nil
	}
}

// RunFunc is a function that runs like a daemon and can be stopped with the
// returned stop function.
type RunFunc func() (stop func(), err error)

// RunCommand returns a RunFunc that can start a command line app with arguments.
// It returns a function that can kill the started process.
func RunCommand(command string, args ...string) RunFunc {
	return func() (func(), error) {
		cmd := exec.Command(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("Error running: \"%s %s\": %w", command, strings.Join(args, " "), err)
		}
		stop := func() {
			cmd.Process.Kill()
		}
		return stop, nil
	}
}

// Run executes the build and run functions. All build functions are executed
// before the run function. It returns an error and stops the executions if an
// error happens. Otherwise it returns a function to stop the run function's execution.
func Run(builds []BuildFunc, run RunFunc) (func(), error) {
	for _, build := range builds {
		if err := build(); err != nil {
			return nil, err
		}
	}

	if run == nil {
		return func() {}, nil
	}

	return run()
}

// FilterFunc can filter files.
type FilterFunc func() bool

// Filter returns a FilterFunc that can filter files based on include and 
// exclude patterns.
func Filter(files, includePatterns, excludePatterns []string) FilterFunc {
    return func() bool {
        for _, file := range files {
            if matchPatterns(excludePatterns, file) {
                continue
            }
            if matchPatterns(includePatterns, file) {
                return true
            }
        }
        return false
    }
}

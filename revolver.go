package revolver

import (
	"os"
	"path/filepath"

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

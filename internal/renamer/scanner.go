package renamer

import (
	"os"
	"path/filepath"
	"strings"
)

// FileResult holds the result for a single file after processing.
type FileResult struct {
	Path         string
	Replacements int
	NewContent   string // only populated when changes exist
	Err          error
}

// ScanOptions controls which files are considered.
type ScanOptions struct {
	// ProjectRoot scans all .kt files recursively under this directory.
	ProjectRoot string
	// SingleFile restricts processing to one specific file.
	SingleFile string
}

// CollectKotlinFiles returns all .kt file paths according to opts.
func CollectKotlinFiles(opts ScanOptions) ([]string, error) {
	if opts.SingleFile != "" {
		abs, err := filepath.Abs(opts.SingleFile)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(abs, ".kt") {
			return nil, nil // silently skip non-kt files
		}
		return []string{abs}, nil
	}

	if opts.ProjectRoot == "" {
		return nil, nil
	}

	root, err := filepath.Abs(opts.ProjectRoot)
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden dirs and common non-source dirs
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "build" || name == "out" || name == ".gradle" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".kt") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ApplyToFiles runs renameFn over each file path, collecting results.
// If dryRun is false, modified files are written back.
func ApplyToFiles(paths []string, dryRun bool, renameFn func(content string) (string, int)) ([]FileResult, error) {
	results := make([]FileResult, 0, len(paths))

	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			results = append(results, FileResult{Path: path, Err: err})
			continue
		}

		original := string(raw)
		modified, count := renameFn(original)

		if count == 0 {
			continue // nothing changed in this file
		}

		r := FileResult{
			Path:         path,
			Replacements: count,
			NewContent:   modified,
		}

		if !dryRun {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				r.Err = err
			}
		}

		results = append(results, r)
	}

	return results, nil
}

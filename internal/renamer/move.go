package renamer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MoveOptions controls the kr move command.
type MoveOptions struct {
	// FilePath is the absolute (or relative) path to the .kt file to move.
	FilePath string
	// NewPackage is the target package, e.g. "com.example.newpackage".
	NewPackage string
	// ProjectRoot is used to scan all .kt files for import rewriting.
	ProjectRoot string
	// DryRun previews changes without writing.
	DryRun bool
}

// MoveResult contains the outcome of a move operation.
type MoveResult struct {
	// MovedFrom / MovedTo are file system paths.
	MovedFrom string
	MovedTo   string
	// ImportResults are files whose imports were updated.
	ImportResults []FileResult
}

// PackageMove performs the full package move:
//  1. Rewrites the package declaration in the source file.
//  2. Scans all .kt files in the project and rewrites imports.
//  3. Moves the file to the correct directory (unless DryRun).
func PackageMove(opts MoveOptions) (*MoveResult, error) {
	absFile, err := filepath.Abs(opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("resolving file path: %w", err)
	}

	if !strings.HasSuffix(absFile, ".kt") {
		return nil, fmt.Errorf("%s is not a Kotlin file", absFile)
	}

	// ── 1. Read source file ────────────────────────────────────────────────
	raw, err := os.ReadFile(absFile)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", absFile, err)
	}
	srcContent := string(raw)

	// ── 2. Detect current package ──────────────────────────────────────────
	oldPackage := extractPackage(srcContent)
	className := strings.TrimSuffix(filepath.Base(absFile), ".kt")

	// ── 3. Rewrite package declaration in source file ──────────────────────
	newSrcContent := rewritePackageDeclaration(srcContent, opts.NewPackage)

	// ── 4. Compute new file path ───────────────────────────────────────────
	newFilePath, err := computeNewPath(opts.ProjectRoot, absFile, opts.NewPackage)
	if err != nil {
		return nil, fmt.Errorf("computing new path: %w", err)
	}

	// ── 5. Rewrite imports in all project .kt files ────────────────────────
	projectFiles, err := CollectKotlinFiles(ScanOptions{ProjectRoot: opts.ProjectRoot})
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}

	// Build old and new fully-qualified names
	var oldFQN, newFQN string
	if oldPackage != "" {
		oldFQN = oldPackage + "." + className
	} else {
		oldFQN = className
	}
	newFQN = opts.NewPackage + "." + className

	// Filter out the source file itself (we've already updated it above)
	var otherFiles []string
	for _, f := range projectFiles {
		if f != absFile {
			otherFiles = append(otherFiles, f)
		}
	}

	importResults, err := ApplyToFiles(otherFiles, opts.DryRun, func(content string) (string, int) {
		return rewriteImport(content, oldFQN, newFQN)
	})
	if err != nil {
		return nil, err
	}

	result := &MoveResult{
		MovedFrom:     absFile,
		MovedTo:       newFilePath,
		ImportResults: importResults,
	}

	if opts.DryRun {
		return result, nil
	}

	// ── 6. Write updated source file to new location ───────────────────────
	if err := os.MkdirAll(filepath.Dir(newFilePath), 0755); err != nil {
		return nil, fmt.Errorf("creating directory for %s: %w", newFilePath, err)
	}

	if err := os.WriteFile(newFilePath, []byte(newSrcContent), 0644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", newFilePath, err)
	}

	// ── 7. Remove old file (only if it moved to a different path) ──────────
	if absFile != newFilePath {
		if err := os.Remove(absFile); err != nil {
			return nil, fmt.Errorf("removing old file %s: %w", absFile, err)
		}
	}

	return result, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

var packageDeclPat = regexp.MustCompile(`(?m)^package\s+[\w.]+`)

func extractPackage(src string) string {
	m := packageDeclPat.FindString(src)
	if m == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(m, "package"))
}

func rewritePackageDeclaration(src, newPackage string) string {
	if packageDeclPat.MatchString(src) {
		return packageDeclPat.ReplaceAllString(src, "package "+newPackage)
	}
	// No package declaration — prepend one
	return "package " + newPackage + "\n\n" + src
}

// rewriteImport replaces `import oldFQN` with `import newFQN` in a file.
// Returns modified content and count of replacements.
func rewriteImport(content, oldFQN, newFQN string) (string, int) {
	// Match: import <oldFQN>  optionally followed by end-of-line or whitespace
	pat := regexp.MustCompile(`(?m)^(import\s+)` + regexp.QuoteMeta(oldFQN) + `(\s*(?:as\s+\w+)?\s*)$`)
	count := 0
	result := pat.ReplaceAllStringFunc(content, func(match string) string {
		count++
		return pat.ReplaceAllString(match, "${1}"+newFQN+"${2}")
	})
	return result, count
}

// computeNewPath figures out where the file should live after the move.
// It looks for the standard "src/main/kotlin" or "src/test/kotlin" prefix in
// the current path and replaces the package path below it.
// Falls back to projectRoot/packagePath/FileName.kt.
func computeNewPath(projectRoot, currentFile, newPackage string) (string, error) {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}

	packageDir := strings.ReplaceAll(newPackage, ".", string(filepath.Separator))
	fileName := filepath.Base(currentFile)

	// Walk up directories from the file's parent to find the source root.
	// A source root is a directory named "kotlin" (or "java") that is an
	// ancestor of the file and a descendant of projectRoot.
	dir := filepath.Dir(currentFile)
	for dir != absRoot && len(dir) >= len(absRoot) {
		base := filepath.Base(dir)
		if base == "kotlin" || base == "java" {
			return filepath.Join(dir, packageDir, fileName), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: put it directly under projectRoot/packagePath
	return filepath.Join(absRoot, packageDir, fileName), nil
}

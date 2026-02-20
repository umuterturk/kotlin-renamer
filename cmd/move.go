package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/umut/kr/internal/renamer"
)

var (
	moveProject string
	moveDryRun  bool
)

var moveCmd = &cobra.Command{
	Use:   "move <file> <new.package.path>",
	Short: "Move a Kotlin file to a new package",
	Long: `Move a .kt file to a new package, updating:
  - The package declaration inside the file
  - All import statements across the project
  - The file's location on disk (to match standard src/main/kotlin layout)

Examples:
  kr move UserService.kt com.example.newpackage --project ./src
  kr move src/main/kotlin/com/example/UserService.kt com.example.util --project ./src --dry-run`,
	Args: cobra.ExactArgs(2),
	RunE: runMove,
}

func init() {
	moveCmd.Flags().StringVar(&moveProject, "project", "",
		"Project root — used to scan all .kt files for import rewriting")
	moveCmd.Flags().BoolVar(&moveDryRun, "dry-run", false,
		"Preview changes without writing files or moving the file")

	_ = moveCmd.MarkFlagRequired("project")
}

func runMove(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	newPackage := args[1]

	// Basic package name validation
	if !isValidPackageName(newPackage) {
		return fmt.Errorf("invalid package name: %q (expected e.g. com.example.mypackage)", newPackage)
	}

	opts := renamer.MoveOptions{
		FilePath:    filePath,
		NewPackage:  newPackage,
		ProjectRoot: moveProject,
		DryRun:      moveDryRun,
	}

	result, err := renamer.PackageMove(opts)
	if err != nil {
		return err
	}

	renamer.PrintMoveResult(os.Stdout, result, moveDryRun)
	return nil
}

func isValidPackageName(pkg string) bool {
	if pkg == "" {
		return false
	}
	// Each segment must be a valid identifier
	parts := splitDot(pkg)
	for _, p := range parts {
		if p == "" {
			return false
		}
		for i, c := range p {
			if i == 0 {
				if !isLetter(c) && c != '_' {
					return false
				}
			} else {
				if !isLetter(c) && !isDigit(c) && c != '_' {
					return false
				}
			}
		}
	}
	return true
}

func splitDot(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

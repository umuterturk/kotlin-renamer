package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/umut/kr/internal/renamer"
)

var (
	renameType    string
	renameFile    string
	renameProject string
	renameClass   string
	renameDryRun  bool
)

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a Kotlin symbol across the project",
	Long: `Rename a Kotlin symbol with syntax-aware, word-boundary matching.

Supported symbol types (--type flag):
  class       class, interface, object declarations + all usages
  interface   same as class
  object      same as class
  method      fun declarations, call sites, and method references
  property    val/var declarations and member access
  parameter   parameter names within function signatures and bodies

Examples:
  kr rename --type class User UserAccount --project ./src
  kr rename --type method calculateTotal computeTotal --project ./src
  kr rename --type method calculateTotal computeTotal --file CartService.kt
  kr rename --type property userId accountId --file UserService.kt --class UserService
  kr rename --type parameter userId accountId --file UserService.kt`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func init() {
	renameCmd.Flags().StringVar(&renameType, "type", "class",
		"Symbol type: class, interface, object, method, property, parameter")
	renameCmd.Flags().StringVar(&renameFile, "file", "",
		"Restrict to a single file")
	renameCmd.Flags().StringVar(&renameProject, "project", "",
		"Project root — scans all .kt files recursively")
	renameCmd.Flags().StringVar(&renameClass, "class", "",
		"(method/property) Scope rename to a specific class name")
	renameCmd.Flags().BoolVar(&renameDryRun, "dry-run", false,
		"Preview changes without writing files")
}

func runRename(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]

	// ── validation ────────────────────────────────────────────────────────────
	if err := renamer.ValidateIdentifier(oldName); err != nil {
		return err
	}
	if err := renamer.ValidateIdentifier(newName); err != nil {
		return err
	}

	symType := strings.ToLower(renameType)
	switch symType {
	case "class", "interface", "object", "method", "property", "parameter":
	default:
		return fmt.Errorf("unknown --type %q; use: class, interface, object, method, property, parameter", renameType)
	}

	if renameFile == "" && renameProject == "" {
		return fmt.Errorf("provide at least one of --file or --project")
	}

	// ── collect files ─────────────────────────────────────────────────────────
	opts := renamer.ScanOptions{
		ProjectRoot: renameProject,
		SingleFile:  renameFile,
	}
	files, err := renamer.CollectKotlinFiles(opts)
	if err != nil {
		return fmt.Errorf("scanning files: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .kt files found")
	}

	// ── build rename function ─────────────────────────────────────────────────
	renameFn := buildRenameFn(symType, oldName, newName)

	// ── apply ─────────────────────────────────────────────────────────────────
	results, err := renamer.ApplyToFiles(files, renameDryRun, renameFn)
	if err != nil {
		return err
	}

	renamer.PrintResults(os.Stdout, results, renameDryRun)
	return nil
}

// buildRenameFn returns a function that renames oldName→newName according to
// the symbol type.
func buildRenameFn(symType, oldName, newName string) func(string) (string, int) {
	switch symType {
	case "class", "interface", "object":
		r := &renamer.ClassRenamer{}
		return func(content string) (string, int) {
			return r.Rename(content, oldName, newName)
		}

	case "method":
		r := &renamer.MethodRenamer{ClassName: renameClass}
		return func(content string) (string, int) {
			return r.Rename(content, oldName, newName)
		}

	case "property":
		r := &renamer.PropertyRenamer{ClassName: renameClass}
		return func(content string) (string, int) {
			return r.Rename(content, oldName, newName)
		}

	case "parameter":
		r := &renamer.ParameterRenamer{}
		return func(content string) (string, int) {
			return r.Rename(content, oldName, newName)
		}
	}

	// unreachable after validation
	return func(content string) (string, int) { return content, 0 }
}

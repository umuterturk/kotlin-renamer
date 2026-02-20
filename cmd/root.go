package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kr",
	Short: "kr — Kotlin Renamer CLI",
	Long: `kr is a fast, syntax-aware rename tool for Kotlin projects.

It uses word-boundary matching and Kotlin-specific context rules to
rename symbols without touching substrings (e.g. renaming User does
NOT affect UserService).

Commands:
  rename    Rename a class, interface, object, method, property, or parameter
  move      Move a .kt file to a new package, updating all imports`,
	SilenceUsage: true,
}

// Execute is the entry point called from main.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion injects the build-time version string into the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}

func init() {
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(moveCmd)
}

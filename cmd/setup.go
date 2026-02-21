package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	claudeCodeOnly bool
	cursorOnly     bool

	// Embedded skill content, injected from main.go via SetSkills.
	claudeCodeSkill []byte
	cursorRule      []byte
)

// SetSkills injects the embedded skill files from the main package.
func SetSkills(claude, cursor []byte) {
	claudeCodeSkill = claude
	cursorRule = cursor
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install AI editor integrations (Claude Code skill, Cursor rule)",
	Long: `Install AI editor integrations so that Claude Code and Cursor
automatically use kr for Kotlin rename and move operations.

By default, both integrations are installed:
  - Claude Code skill → ~/.claude/skills/kr/SKILL.md  (global)
  - Cursor rule       → .cursor/rules/kotlin-renamer.mdc (current project)

Use --claude-code or --cursor to install only one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		installBoth := !claudeCodeOnly && !cursorOnly

		if installBoth || claudeCodeOnly {
			if err := installClaudeCodeSkill(); err != nil {
				return err
			}
		}

		if installBoth || cursorOnly {
			if err := installCursorRule(); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	setupCmd.Flags().BoolVar(&claudeCodeOnly, "claude-code", false, "Install only the Claude Code skill")
	setupCmd.Flags().BoolVar(&cursorOnly, "cursor", false, "Install only the Cursor rule")
}

func installClaudeCodeSkill() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	dir := filepath.Join(home, ".claude", "skills", "kr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", dir, err)
	}

	dest := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(dest, claudeCodeSkill, 0o644); err != nil {
		return fmt.Errorf("could not write %s: %w", dest, err)
	}

	fmt.Printf("✅ Claude Code skill installed → %s\n", dest)
	return nil
}

func installCursorRule() error {
	dir := filepath.Join(".cursor", "rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", dir, err)
	}

	dest := filepath.Join(dir, "kotlin-renamer.mdc")
	if err := os.WriteFile(dest, cursorRule, 0o644); err != nil {
		return fmt.Errorf("could not write %s: %w", dest, err)
	}

	fmt.Printf("✅ Cursor rule installed → %s\n", dest)
	return nil
}

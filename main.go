package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/umut/kr/cmd"
)

// version is set at build time via -ldflags="-X main.version=v1.0.0"
var version = "dev"

//go:embed skills/claude-code/SKILL.md
var claudeCodeSkill []byte

//go:embed skills/cursor/kotlin-renamer.mdc
var cursorRule []byte

func main() {
	cmd.SetVersion(version)
	cmd.SetSkills(claudeCodeSkill, cursorRule)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

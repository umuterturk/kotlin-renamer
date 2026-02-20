package main

import (
	"fmt"
	"os"

	"github.com/umut/kr/cmd"
)

// version is set at build time via -ldflags="-X main.version=v1.0.0"
var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

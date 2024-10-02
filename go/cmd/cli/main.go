package main

import (
	"os"

	"github.com/go-goal/tagger/internal/cli"
)

func main() {
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}

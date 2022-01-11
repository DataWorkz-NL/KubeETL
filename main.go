package main

import (
	"os"

	"github.com/dataworkz/kubeetl/internal/commands"
)

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

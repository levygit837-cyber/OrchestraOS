package main

import (
	"os"

	"github.com/levygit837-cyber/OrchestraOS/cmd/orchestraos/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

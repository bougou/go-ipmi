package main

import (
	"fmt"
	"os"

	"github.com/bougou/go-ipmi/goipmi/commands"
)

func main() {
	rootCmd := commands.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/xstp/go-ipmi/goipmi/commands"
)

func main() {
	rootCmd := commands.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

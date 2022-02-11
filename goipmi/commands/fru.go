package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdFRU() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fru",
		Short: "fru",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdFRUList())

	return cmd
}

func NewCmdFRUList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			frus, err := client.GetFRUs()
			if err != nil {
				CheckErr(fmt.Errorf("GetFRUs failed, err: %s", err))
			}

			for _, fru := range frus {
				fmt.Println(fru.String())
			}
		},
	}
	return cmd
}

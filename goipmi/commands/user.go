package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "user",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdUserList())

	return cmd
}

func NewCmdUserList() *cobra.Command {
	usage := `
list [<channel number>]
list [<channel number>]
`
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %s", err))
			}
			channelNumber := uint8(id)

			users, err := client.ListUser(channelNumber)
			if err != nil {
				CheckErr(fmt.Errorf("ListUser failed, err: %s", err))
			}

			client.Debug("users", users)
		},
	}
	return cmd
}

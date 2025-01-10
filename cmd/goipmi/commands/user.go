package commands

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi"
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
	cmd.AddCommand(NewCmdUserSummary())

	return cmd
}

func NewCmdUserList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [<channel number>]",
		Short: "list [<channel number>]",
		Run: func(cmd *cobra.Command, args []string) {
			var channelNumber uint8

			if len(args) == 0 {
				channelNumber = ipmi.ChannelNumberSelf
			}

			if len(args) > 1 {
				id, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
				}
				channelNumber = uint8(id)
			}

			ctx := context.Background()
			users, err := client.ListUser(ctx, channelNumber)
			if err != nil {
				CheckErr(fmt.Errorf("ListUser failed, err: %w", err))
			}

			fmt.Println(ipmi.FormatUsers(users))
		},
	}
	return cmd
}

func NewCmdUserSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary [<channel number>]",
		Short: "summary [<channel number>]",
		Run: func(cmd *cobra.Command, args []string) {
			var channelNumber uint8
			if len(args) == 0 {
				channelNumber = ipmi.ChannelNumberSelf
			}

			if len(args) > 1 {
				id, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
				}
				channelNumber = uint8(id)
			}

			ctx := context.Background()

			res, err := client.GetUserAccess(ctx, channelNumber, 0x01)
			if err != nil {
				CheckErr(fmt.Errorf("GetUserAccess failed, err: %w", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

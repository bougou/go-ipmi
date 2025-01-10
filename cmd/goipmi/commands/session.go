package commands

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "session",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdSessionInfo())

	return cmd
}

func NewCmdSessionInfo() *cobra.Command {
	// cSpell: disable
	usage := `Session Commands: info <active | all | id 0xnnnnnnnn | handle 0xnn>`
	// cSpell: enable

	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println(usage)
				return
			}

			ctx := context.Background()

			switch args[0] {
			case "active":
				request := &ipmi.GetSessionInfoRequest{
					SessionIndex: 0, // current active
				}
				res, err := client.GetSessionInfo(ctx, request)
				if err != nil {
					CheckErr(fmt.Errorf("GetSessionInfo failed, err: %w", err))
				}
				fmt.Println(res.Format())
			}
		},
	}
	return cmd
}

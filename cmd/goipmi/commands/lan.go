package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdLan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lan",
		Short: "lan",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdLanStats())
	cmd.AddCommand(NewCmdLanPrint())

	return cmd
}

func NewCmdLanStats() *cobra.Command {
	usage := `
stats get [<channel number>]
stats clear [<channel number>]
`
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "stats",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]
			id, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()
			switch action {
			case "get":
				res, err := client.GetIPStatistics(ctx, channelNumber, false)
				if err != nil {
					CheckErr(fmt.Errorf("GetIPStatistics failed, err: %w", err))
				}
				fmt.Println(res.Format())
			case "clear":
				res, err := client.GetIPStatistics(ctx, channelNumber, true)
				if err != nil {
					CheckErr(fmt.Errorf("GetIPStatistics failed, err: %w", err))
				}
				fmt.Println(res.Format())
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdLanPrint() *cobra.Command {
	usage := `
print [<channel number>]
`
	cmd := &cobra.Command{
		Use:   "print",
		Short: "print",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()
			lanConfig, err := client.GetLanConfigParams(ctx, channelNumber)
			if err != nil {
				CheckErr(fmt.Errorf("GetLanConfigParams failed, err: %w", err))
			}

			client.Debug("Lan Config", lanConfig)

			fmt.Println(lanConfig.Format())
		},
	}
	return cmd
}

package commands

import (
	"fmt"
	"github.com/xstp/go-ipmi"

	"github.com/spf13/cobra"
)

func NewCmdChannel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "channel",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdChannelInfo())

	return cmd
}

func NewCmdChannelInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			var channelNumber uint8
			if len(args) == 0 {
				channelNumber = 0x0e
			}
			if len(args) >= 1 {
				i, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid channel number, err: %s", err))
				}
				channelNumber = uint8(i)
			}
			res, err := client.GetChannelInfo(channelNumber)
			if err != nil {
				if err != nil {
					CheckErr(fmt.Errorf("GetChannelInfo failed, err: %s", err))
				}
			}
			fmt.Println(res.Format())

			res2, err := client.GetChannelAccess(channelNumber, ipmi.ChannelAccessOption_Volatile)
			if err != nil {
				if err != nil {
					CheckErr(fmt.Errorf("GetChannelAccess failed, err: %s", err))
				}
			}
			fmt.Println("  Volatile(active) Settings")
			fmt.Println(res2.Format())

			res3, err := client.GetChannelAccess(channelNumber, ipmi.ChannelAccessOption_NonVolatile)
			if err != nil {
				if err != nil {
					CheckErr(fmt.Errorf("GetChannelAccess failed, err: %s", err))
				}
			}
			fmt.Println("  Non-Volatile Settings")
			fmt.Println(res3.Format())
		},
	}
	return cmd
}

package commands

import (
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdSDR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdr",
		Short: "sdr",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sdr ...")
		},
	}
	cmd.AddCommand(NewCmdSDRInfo())
	cmd.AddCommand(NewCmdSDRGet())

	return cmd
}

func NewCmdSDRInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			sdrRepoInfo, err := client.GetSDRRepoInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRRepoInfo failed, err: %s", err))
			}
			fmt.Println(sdrRepoInfo.Format())
		},
	}
	return cmd
}

func NewCmdSDRGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetSDR(0x00)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRRepoInfo failed, err: %s", err))
			}

			client.DebugBytes("SDR Record Data", res.RecordData, 16)

			sdr, err := ipmi.ParseSDR(res.RecordData)
			if err != nil {
				CheckErr(fmt.Errorf("ParseSDR failed, err: %s", err))

			}
			client.Debug("SDR", sdr)
		},
	}
	return cmd
}

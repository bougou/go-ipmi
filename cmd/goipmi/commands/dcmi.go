package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdDCMI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dcmi",
		Short: "dcmi",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdDCMIPower())
	cmd.AddCommand(NewCmdDCMIAssetTag())

	return cmd
}

func NewCmdDCMIPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "power",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.AddCommand(NewCmdDCMIRead())
	return cmd
}

func NewCmdDCMIRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reading",
		Short: "reading",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetDCMIPowerReading()
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIPowerReading failed, err: %s", err))
			}
			fmt.Println(resp.Format())
		},
	}
	return cmd
}

func NewCmdDCMIAssetTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset_tag",
		Short: "asset_tag",
		Run: func(cmd *cobra.Command, args []string) {
			var assetTag string
			var offset uint8
			for {
				resp, err := client.GetDCMIAssetTag(offset)
				if err != nil {
					CheckErr(fmt.Errorf("GetDCMIAssetTag failed, err: %s", err))
				}
				assetTag += string(resp.AssetTag)
				if resp.TotalLength <= offset+uint8(len(resp.AssetTag)) {
					break
				}
				offset += uint8(len(resp.AssetTag))
			}
			fmt.Printf("Asset tag: %s\n", assetTag)
		},
	}
	return cmd
}

package commands

import (
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdSEL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sel",
		Short: "sel",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdSELInfo())
	cmd.AddCommand(NewCmdSELList())
	cmd.AddCommand(NewCmdSELElist())

	return cmd
}

func NewCmdSELInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			selInfo, err := client.GetSELInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}
			fmt.Println(selInfo.Format())

			selAllocInfo, err := client.GetSELAllocInfo()
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}
			fmt.Println(selAllocInfo.Format())
		},
	}
	return cmd
}

func NewCmdSELList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			selEntries, err := client.GetSELEntries(0)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}

			fmt.Println(ipmi.FormatSELs(selEntries, nil))
		},
	}
	return cmd
}

func NewCmdSELElist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "elist",
		Short: "elist",
		Run: func(cmd *cobra.Command, args []string) {
			sdrMap, err := client.GetSDRsMap(0)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRsMap failed, err: %s", err))
			}

			selEntries, err := client.GetSELEntries(0)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %s", err))
			}

			fmt.Println(ipmi.FormatSELs(selEntries, sdrMap))
		},
	}
	return cmd
}

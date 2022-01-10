package commands

import (
	"errors"
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
	cmd.AddCommand(NewCmdSELGet())
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

func NewCmdSELGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(errors.New("no Record ID supplied"))
			}
			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid Record ID passed, err: %s", err))
			}
			recordID := uint16(id)

			selEntryRes, err := client.GetSELEntry(0x0, recordID)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELEntry failed, err: %s", err))
			}

			sel, err := ipmi.ParseSEL(selEntryRes.Data)
			if err != nil {
				CheckErr(fmt.Errorf("ParseSEL failed, err: %s", err))
			}
			fmt.Println(ipmi.FormatSELs([]*ipmi.SEL{sel}, nil))
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

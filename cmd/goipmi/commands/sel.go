package commands

import (
	"context"
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
			ctx := context.Background()
			selInfo, err := client.GetSELInfo(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %w", err))
			}
			fmt.Println(selInfo.Format())

			selAllocInfo, err := client.GetSELAllocInfo(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %w", err))
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
				CheckErr(fmt.Errorf("invalid Record ID passed, err: %w", err))
			}
			recordID := uint16(id)

			ctx := context.Background()
			selEntryRes, err := client.GetSELEntry(ctx, 0x0, recordID)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELEntry failed, err: %w", err))
			}

			sel, err := ipmi.ParseSEL(selEntryRes.Data)
			if err != nil {
				CheckErr(fmt.Errorf("ParseSEL failed, err: %w", err))
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
			ctx := context.Background()
			selEntries, err := client.GetSELEntries(ctx, 0)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %w", err))
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
			ctx := context.Background()
			sdrsMap, err := client.GetSDRsMap(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRsMap failed, err: %w", err))
			}

			selEntries, err := client.GetSELEntries(ctx, 0)
			if err != nil {
				CheckErr(fmt.Errorf("GetSELInfo failed, err: %w", err))
			}

			fmt.Println(ipmi.FormatSELs(selEntries, sdrsMap))
		},
	}
	return cmd
}

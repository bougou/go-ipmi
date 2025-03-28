package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdMC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mc",
		Short: "mc",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdMCInfo())
	cmd.AddCommand(NewCmdMCReset())
	cmd.AddCommand(NewCmdMC_ACPI())
	cmd.AddCommand(NewCmdMC_GUID())
	cmd.AddCommand(NewCmdMC_Watchdog())

	return cmd
}

func NewCmdMCInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetDeviceID(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDeviceID failed, err: %w", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdMCReset() *cobra.Command {
	usage := "reset <cold|warm>"

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "reset",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
			ctx := context.Background()
			switch args[0] {
			case "warm":
				if err := client.WarmReset(ctx); err != nil {
					CheckErr(fmt.Errorf("WarmReset failed, err: %w", err))
				}

			case "cold":
				if err := client.ColdReset(ctx); err != nil {
					CheckErr(fmt.Errorf("ColdReset failed, err: %w", err))
				}
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdMC_ACPI() *cobra.Command {
	usage := "acpi <get|set>"

	cmd := &cobra.Command{
		Use:   "acpi",
		Short: "acpi",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
			ctx := context.Background()
			switch args[0] {
			case "get":
				res, err := client.GetACPIPowerState(ctx)
				if err != nil {
					CheckErr(fmt.Errorf("GetACPIPowerState failed, err: %w", err))
				}
				fmt.Println(res.Format())
			case "set":
				//
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdMC_GUID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guid",
		Short: "guid",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetSystemGUID(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetSystemGUID failed, err: %w", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdMC_Watchdog() *cobra.Command {
	usage := `watchdog <get|reset|off>
  get    :  Get Current Watchdog settings
  reset  :  Restart Watchdog timer based on most recent settings
  off    :  Shut off a running Watchdog timer
`

	cmd := &cobra.Command{
		Use:   "watchdog",
		Short: "watchdog",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
			ctx := context.Background()
			switch args[0] {
			case "get":
				res, err := client.GetWatchdogTimer(ctx)
				if err != nil {
					CheckErr(fmt.Errorf("GetWatchdogTimer failed, err: %w", err))
				}
				fmt.Println(res.Format())
			case "reset":
				if _, err := client.ResetWatchdogTimer(ctx); err != nil {
					CheckErr(fmt.Errorf("ResetWatchdogTimer failed, err: %w", err))
				}
			case "off":
				//
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

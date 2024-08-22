package commands

import (
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
			res, err := client.GetDeviceID()
			if err != nil {
				CheckErr(fmt.Errorf("GetDeviceID failed, err: %s", err))
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
			switch args[0] {
			case "warm":
				if err := client.WarmReset(); err != nil {
					CheckErr(fmt.Errorf("WarmReset failed, err: %s", err))
				}

			case "cold":
				if err := client.ColdReset(); err != nil {
					CheckErr(fmt.Errorf("ColdReset failed, err: %s", err))
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
			switch args[0] {
			case "get":
				res, err := client.GetACPIPowerState()
				if err != nil {
					CheckErr(fmt.Errorf("GetACPIPowerState failed, err: %s", err))
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
			res, err := client.GetSystemGUID()
			if err != nil {
				CheckErr(fmt.Errorf("GetSystemGUID failed, err: %s", err))
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
			switch args[0] {
			case "get":
				res, err := client.GetWatchdogTimer()
				if err != nil {
					CheckErr(fmt.Errorf("GetWatchdogTimer failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "reset":
				if _, err := client.ResetWatchdogTimer(); err != nil {
					CheckErr(fmt.Errorf("ResetWatchdogTimer failed, err: %s", err))
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

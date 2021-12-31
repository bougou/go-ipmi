package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdChassis() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chassis",
		Short: "chassis",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.AddCommand(NewCmdChassisStatus())
	cmd.AddCommand(NewCmdChassisPolicy())
	cmd.AddCommand(NewCmdChassisPower())
	cmd.AddCommand(NewCmdChassisCapabilities())
	cmd.AddCommand(NewCmdChassisRestartCause())
	cmd.AddCommand(NewCmdChassisBootParam())

	return cmd
}

func NewCmdChassisStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "status",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := client.GetChassisStatus()
			if err != nil {
				CheckErr(fmt.Errorf("GetChassisStatus failed, err: %s", err))
			}
			fmt.Println(status.Format())
		},
	}
	return cmd
}

func NewCmdChassisPolicy() *cobra.Command {
	usage := `chassis policy <state>
list        : return supported policies
always-on   : turn on when power is restored
previous    : return to previous state when power is restored
always-off  : stay off after power is restored`

	cmd := &cobra.Command{
		Use:   "policy",
		Short: "policy",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) >= 1 {
				switch args[0] {
				case "list":
					fmt.Printf("Supported chassis power policy: %s\n", strings.Join(ipmi.SupportedPowerRestorePolicies, " "))
					return
				case "always-on":
					_, err := client.SetPowerRestorePolicy(ipmi.PowerRestorePolicyAlwaysOn)
					if err != nil {
						CheckErr(fmt.Errorf("SetPowerRestorePolicy failed, err: %s", err))
					}
				case "previous":
					_, err := client.SetPowerRestorePolicy(ipmi.PowerRestorePolicyAlwaysPrevious)
					if err != nil {
						CheckErr(fmt.Errorf("SetPowerRestorePolicy failed, err: %s", err))
					}
				case "always-off":
					_, err := client.SetPowerRestorePolicy(ipmi.PowerRestorePolicyAlwaysOff)
					if err != nil {
						CheckErr(fmt.Errorf("SetPowerRestorePolicy failed, err: %s", err))
					}
				default:
					fmt.Println(usage)
				}
			}
		},
	}

	return cmd
}

func NewCmdChassisPower() *cobra.Command {
	usage := "chassis power Commands: status, on, off, cycle, reset, diag, soft"

	cmd := &cobra.Command{
		Use:   "power",
		Short: "power",
		Run: func(cmd *cobra.Command, args []string) {
			var c ipmi.ChassisControl
			if len(args) == 0 {
				fmt.Println(usage)
			}

			if len(args) >= 1 {
				switch args[0] {
				case "status":
					status, err := client.GetChassisStatus()
					if err != nil {
						CheckErr(fmt.Errorf("GetChassisStatus failed, err: %s", err))
					}
					powerStatus := "off"
					if status.PowerIsOn {
						powerStatus = "on"
					}
					fmt.Printf("Chassis Power is %s\n", powerStatus)
					return
				case "on":
					c = ipmi.ChassisControlPowerUp
				case "off":
					c = ipmi.ChassisControlPowerDown
				case "cycle":
					c = ipmi.ChassisControlPowerCycle
				case "reset":
					c = ipmi.ChassisControlHardwareRest
				case "diag":
					c = ipmi.ChassisControlDiagnosticInterrupt
				case "soft":
					c = ipmi.ChassisControlSoftShutdown
				default:
					CheckErr(errors.New(usage))
					return
				}

				if _, err := client.ChassisControl(c); err != nil {
					CheckErr(fmt.Errorf("ChassisControl failed, err: %s", err))
					return
				}
			}
		},
	}

	return cmd
}

func NewCmdChassisCapabilities() *cobra.Command {
	usage := "chassis cap Commands: get or set"

	cmd := &cobra.Command{
		Use:   "Get or set chassis capabilities",
		Short: "cap",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println(usage)
				return
			}
			if len(args) >= 1 {
				switch args[0] {
				case "get":
					cap, err := client.GetChassisCapabilities()
					if err != nil {
						CheckErr(fmt.Errorf("GetChassisCapabilities failed, err: %s", err))
						return
					}
					fmt.Println(cap.Format())
					return
				case "set":
				}
			}
		},
	}
	return cmd
}

func NewCmdChassisRestartCause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart_cause",
		Short: "restart_cause",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetSystemRestartCause()
			if err != nil {
				CheckErr(fmt.Errorf("GetSystemRestartCause failed, err: %s", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdChassisBootParam() *cobra.Command {
	usage := `bootparam get <param #>
bootparam set bootflag <device> [options=...]
 Legal devices are:
  none        : No override
  force_pxe   : Force PXE boot
  force_disk  : Force boot from default Hard-drive
  force_safe  : Force boot from default Hard-drive, request Safe Mode
  force_diag  : Force boot from Diagnostic Partition
  force_cdrom : Force boot from CD/DVD
  force_bios  : Force boot into BIOS Setup
 Legal options are:
  help    : print this message
  PEF     : Clear valid bit on reset/power cycle cause by PEF
  timeout : Automatically clear boot flag valid bit on timeout
  watchdog: Clear valid bit on reset/power cycle cause by watchdog
  reset   : Clear valid bit on push button reset/soft reset
  power   : Clear valid bit on power up via power push button or wake event
 Any Option may be prepended with no- to invert sense of operation
`

	cmd := &cobra.Command{
		Use:   "bootparam",
		Short: "bootparam",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println(usage)
				return
			}

			switch args[0] {
			case "get":
				parameterSelector := args[1]
				i, err := strconv.Atoi(parameterSelector)
				if err != nil {
					CheckErr(fmt.Errorf("param # must be a valid interger in range (0-127), err: %s", err))
				}

				res, err := client.GetSystemBootOptions(ipmi.BootOptionParameterSelector(i))
				if err != nil {
					CheckErr(fmt.Errorf("GetSystemBootOptions failed, err: %s", err))
				}
				fmt.Println(res.Format())
			}

		},
	}
	return cmd
}

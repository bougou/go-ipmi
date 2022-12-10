package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xstp/go-ipmi"

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
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdChassisStatus())
	cmd.AddCommand(NewCmdChassisPolicy())
	cmd.AddCommand(NewCmdChassisPower())
	cmd.AddCommand(NewCmdChassisCapabilities())
	cmd.AddCommand(NewCmdChassisRestartCause())
	cmd.AddCommand(NewCmdChassisBootParam())
	cmd.AddCommand(NewCmdChassisBootdev())
	cmd.AddCommand(NewCmdChassisPoh())

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
			if len(args) == 0 {
				fmt.Println(usage)
			}

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
					_, err := client.SetPowerRestorePolicy(ipmi.PowerRestorePolicyPrevious)
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
		Use:   "cap",
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
available param #
  0 : Set In Progress (volatile)
	1 : service partition selector (semi-volatile)
	2 : service partition scan (non-volatile)
	3 : BMC boot flag valid bit clearing (semi-volatile)
	4 : boot info acknowledge (semi-volatile)
	5 : boot flags (semi-volatile)
	6 : boot initiator info (semi-volatile)
	7 : boot initiator mailbox (semi-volatile)
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
				i, err := parseStringToInt64(parameterSelector)
				if err != nil {
					CheckErr(fmt.Errorf("param %s must be a valid interger in range (0-127), err: %s", parameterSelector, err))
				}

				res, err := client.GetSystemBootOptions(ipmi.BootOptionParameterSelector(i))
				if err != nil {
					CheckErr(fmt.Errorf("GetSystemBootOptions failed, err: %s", err))
				}
				fmt.Println(res.Format())

			case "set":
				if len(args) < 3 {
					fmt.Println(usage)
					return
				}

				parameterSelector := args[1]
				// currently only support set bootflag
				if parameterSelector != "bootflag" {
					fmt.Println(usage)
					return
				}

				var f = func(bootDevice string) ipmi.BootDeviceSelector {
					m := map[string]ipmi.BootDeviceSelector{
						"none":        ipmi.BootDeviceSelectorNoOverride,
						"force_pxe":   ipmi.BootDeviceSelectorForcePXE,
						"force_disk":  ipmi.BootDeviceSelectorForceHardDrive,
						"force_safe":  ipmi.BootDeviceSelectorForceHardDriveSafe,
						"force_diag":  ipmi.BootDeviceSelectorForceDiagnosticPartition,
						"force_cdrom": ipmi.BootDeviceSelectorForceCDROM,
						"force_bios":  ipmi.BootDeviceSelectorForceBIOSSetup,
					}
					if s, ok := m[bootDevice]; ok {
						return s
					}
					return ipmi.BootDeviceSelectorNoOverride
				}
				bootDeviceSelector := f(args[2])

				request := &ipmi.SetSystemBootOptionsRequest{
					MarkParameterInvalid: false,
					ParameterSelector:    ipmi.BOPS_BootFlags,
					BootOptionParameter: ipmi.BootOptionParameter{
						BootFlags: &ipmi.BOP_BootFlags{
							BootFlagsValid:     true,
							Persist:            false,
							BIOSBootType:       ipmi.BIOSBootTypeLegacy,
							BootDeviceSelector: bootDeviceSelector,
						},
					},
				}
				res, err := client.SetSystemBootOptions(request)
				if err != nil {
					CheckErr(fmt.Errorf("SetSystemBootOptions failed, err: %s", err))
				}
				fmt.Println(res.Format())
			}
		},
	}
	return cmd
}

func NewCmdChassisPoh() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poh",
		Short: "poh",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetPOHCounter()
			if err != nil {
				CheckErr(fmt.Errorf("GetSystemRestartCause failed, err: %s", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdChassisBootdev() *cobra.Command {
	usage := `bootdev <device> [clear-cmos=yes|no]
bootdev <device> [options=help,...]
  none  : Do not change boot device order
  pxe   : Force PXE boot
  disk  : Force boot from default Hard-drive
  safe  : Force boot from default Hard-drive, request Safe Mode
  diag  : Force boot from Diagnostic Partition
  cdrom : Force boot from CD/DVD
  bios  : Force boot into BIOS Setup
  floppy: Force boot from Floppy/primary removable media`

	cmd := &cobra.Command{
		Use:   "bootdev",
		Short: "bootdev",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println(usage)
				return
			}

			var dev ipmi.BootDeviceSelector
			switch args[0] {
			case "none":
				dev = ipmi.BootDeviceSelectorNoOverride
			case "pxe":
				dev = ipmi.BootDeviceSelectorForcePXE
			case "disk":
				dev = ipmi.BootDeviceSelectorForceHardDrive
			case "safe":
				dev = ipmi.BootDeviceSelectorForceHardDriveSafe
			case "diag":
				dev = ipmi.BootDeviceSelectorForceDiagnosticPartition
			case "cdrom":
				dev = ipmi.BootDeviceSelectorForceCDROM
			case "floppy":
				dev = ipmi.BootDeviceSelectorForceFloppy
			case "bios":
				dev = ipmi.BootDeviceSelectorForceBIOSSetup
			default:
				return
			}
			bootFlags := &ipmi.BOP_BootFlags{
				BootDeviceSelector: dev,
			}

			if len(args) > 1 {
				var optionsStr string

				if args[1] == "clear-cmos=yes" {
					optionsStr = "clear-cmos"
				} else if strings.HasPrefix(args[1], "options=") {
					optionsStr = strings.TrimPrefix(args[1], "options=")
				}

				options := strings.Split(optionsStr, ",")
				for _, option := range options {
					if option == "help" {
						fmt.Println(bootFlags.OptionsHelp())
						return
					}
				}
				if err := bootFlags.ParseFromOptions(options); err != nil {
					CheckErr(fmt.Errorf("ParseFromOptions failed, err: %s", err))
					return
				}
			}

			if err := client.SetBootParamBootFlags(bootFlags); err != nil {
				CheckErr(fmt.Errorf("SetBootParamBootFlags failed, err: %s", err))
			}

			fmt.Printf("Set Boot Device to %s\n", args[0])
		},
	}

	return cmd
}

package commands

import (
	"context"
	"fmt"
	"strconv"

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
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdSDRInfo())
	cmd.AddCommand(NewCmdSDRGet())
	cmd.AddCommand(NewCmdSDRList())
	cmd.AddCommand(NewCmdSDRType())

	return cmd
}

func NewCmdSDRInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			sdrRepoInfo, err := client.GetSDRRepoInfo(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRRepoInfo failed, err: %w", err))
			}
			fmt.Println(sdrRepoInfo.Format())
		},
	}
	return cmd
}

func parseStringToInt64(s string) (int64, error) {
	if len(s) > 2 {
		if s[0] == '0' {
			return strconv.ParseInt(s, 0, 64)
		}
	}
	return strconv.ParseInt(s, 10, 64)
}

func NewCmdSDRGet() *cobra.Command {
	usage := `sdr get <sensorNumber> or <sensorName>, sensorName should be quoted if contains space`

	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("no Sensor ID or Sensor Name supplied, usage: %s", usage))
			}
			ctx := context.Background()

			var sdr *ipmi.SDR
			var err error

			id, err := parseStringToInt64(args[0])
			if err != nil {
				// suppose args is sensor name
				sdr, err = client.GetSDRBySensorName(ctx, args[0])
				if err != nil {
					CheckErr(fmt.Errorf("GetSDRBySensorName failed, err: %w", err))
				}
			} else {
				sensorID := uint8(id)
				sdr, err = client.GetSDRBySensorID(ctx, sensorID)
				if err != nil {
					CheckErr(fmt.Errorf("GetSDRBySensorID failed, err: %w", err))
				}
			}

			client.Debug("SDR", sdr)
			fmt.Println(sdr)
		},
	}

	return cmd
}

func NewCmdSDRType() *cobra.Command {

	sensorTypesText := `
Sensor Types:
	Temperature               (0x01)   Voltage                   (0x02)
	Current                   (0x03)   Fan                       (0x04)
	Physical Security         (0x05)   Platform Security         (0x06)
	Processor                 (0x07)   Power Supply              (0x08)
	Power Unit                (0x09)   Cooling Device            (0x0a)
	Other                     (0x0b)   Memory                    (0x0c)
	Drive Slot / Bay          (0x0d)   POST Memory Resize        (0x0e)
	System Firmwares          (0x0f)   Event Logging Disabled    (0x10)
	Watchdog1                 (0x11)   System Event              (0x12)
	Critical Interrupt        (0x13)   Button                    (0x14)
	Module / Board            (0x15)   Microcontroller           (0x16)
	Add-in Card               (0x17)   Chassis                   (0x18)
	Chip Set                  (0x19)   Other FRU                 (0x1a)
	Cable / Interconnect      (0x1b)   Terminator                (0x1c)
	System Boot Initiated     (0x1d)   Boot Error                (0x1e)
	OS Boot                   (0x1f)   OS Critical Stop          (0x20)
	Slot / Connector          (0x21)   System ACPI Power State   (0x22)
	Watchdog2                 (0x23)   Platform Alert            (0x24)
	Entity Presence           (0x25)   Monitor ASIC              (0x26)
	LAN                       (0x27)   Management Subsys Health  (0x28)
	Battery                   (0x29)   Session Audit             (0x2a)
	Version Change            (0x2b)   FRU State                 (0x2c)
`
	// usage := `sdr type [all|<sensorTypeName>|<sensorTypeNumber>]`
	cmd := &cobra.Command{
		Use:   "type",
		Short: "type",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println(sensorTypesText)
				return
			}

			if len(args) >= 1 {
				ctx := context.Background()

				if args[0] == "all" {
					sensors, err := client.GetSensors(ctx)
					if err != nil {
						fmt.Printf("failed to get all sensors: %s", err)
						return
					}

					fmt.Println(ipmi.FormatSensors(true, sensors...))
					return
				}

				sensorType, err := ipmi.SensorTypeFromNameOrNumber(args[0])
				if err != nil {
					fmt.Printf("invalid sensor type: %s", args[0])
					return
				}

				sensors, err := client.GetSensors(ctx, ipmi.SensorFilterOptionIsSensorType(sensorType))
				if err != nil {
					fmt.Printf("failed to get (%s) sensors: %s", sensorType, err)
					return
				}

				fmt.Println(ipmi.FormatSensors(true, sensors...))
			}
		},
	}

	return cmd
}

func NewCmdSDRList() *cobra.Command {
	usage := `sdr list <all|full|compact|event|mcloc|fru|generic>`
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			recordTypes := []ipmi.SDRRecordType{}

			// default only get Full and Compact SDR
			if len(args) == 0 {
				recordTypes = append(recordTypes, ipmi.SDRRecordTypeFullSensor, ipmi.SDRRecordTypeCompactSensor)
			}

			ctx := context.Background()

			if len(args) >= 1 {
				switch args[0] {
				case "all":
					// no filter, recordTypes is empty.
				case "full":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeFullSensor)
				case "compact":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeCompactSensor)
				case "event":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeEventOnly)
				case "mcloc":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeManagementControllerDeviceLocator)
				case "fru":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeFRUDeviceLocator)
					sdrs, err := client.GetSDRs(ctx, recordTypes...)
					if err != nil {
						CheckErr(fmt.Errorf("GetSDRs failed, err: %w", err))
					}

					fmt.Println(ipmi.FormatSDRs_FRU(sdrs))
					return

				case "generic":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeGenericLocator)
				default:
					CheckErr(fmt.Errorf("unknown supported record type (%s), usage: %s", args[0], usage))
					return
				}
			}

			sdrs, err := client.GetSDRs(ctx, recordTypes...)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRs failed, err: %w", err))
			}

			fmt.Println(ipmi.FormatSDRs(sdrs))
		},
	}

	return cmd
}

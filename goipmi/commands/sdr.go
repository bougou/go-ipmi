package commands

import (
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

			var sdr *ipmi.SDR
			var err error

			id, err := parseStringToInt64(args[0])
			if err != nil {
				// suppose args is sensor name
				sdr, err = client.GetSDRBySensorName(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("GetSDRBySensorName failed, err: %s", err))
				}
			} else {
				sensorID := uint8(id)
				sdr, err = client.GetSDRBySensorID(sensorID)
				if err != nil {
					CheckErr(fmt.Errorf("GetSDRBySensorID failed, err: %s", err))
				}
			}

			client.Debug("SDR", sdr)
			fmt.Println(sdr)
		},
	}

	return cmd
}

func NewCmdSDRList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			recordTypes := []ipmi.SDRRecordType{}

			// default only get Full and Compacat SDR
			if len(args) == 0 {
				recordTypes = append(recordTypes, ipmi.SDRRecordTypeFullSensor, ipmi.SDRRecordTypeCompactSensor)
			}

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
				case "generic":
					recordTypes = append(recordTypes, ipmi.SDRRecordTypeGenericLocator)
				default:
					CheckErr(fmt.Errorf("unkown supported record type (%s)", args[0]))
					return
				}
			}

			sdrs, err := client.GetSDRs(recordTypes...)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRs failed, err: %s", err))
			}

			fmt.Println(ipmi.FormatSDRs(sdrs))
		},
	}

	return cmd
}

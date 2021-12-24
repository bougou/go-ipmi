package commands

import (
	"errors"
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

func NewCmdSDRGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(errors.New("no Sensor ID supplied"))
			}
			id, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				CheckErr(fmt.Errorf("invalid Sensor ID passed, err: %s", err))
			}
			sensorID := uint16(id)
			res, err := client.GetSDR(sensorID)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDR failed, err: %s", err))
			}

			client.DebugBytes("SDR Record Data", res.RecordData, 16)

			sdr, err := ipmi.ParseSDR(res.RecordData, res.NextRecordID)
			if err != nil {
				CheckErr(fmt.Errorf("ParseSDR failed, err: %s", err))
			}
			client.Debug("SDR", sdr)
		},
	}

	return cmd
}

func NewCmdSDRList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			var recordType ipmi.SDRRecordType = 0

			if len(args) >= 1 {
				switch args[0] {
				case "all":
				case "full":
					recordType = ipmi.SDRRecordTypeFullSensor
				case "compact":
					recordType = ipmi.SDRRecordTypeCompactSensor
				case "event":
					recordType = ipmi.SDRRecordTypeEventOnly
				case "mcloc":
					recordType = ipmi.SDRRecordTypeManagementControllerDeviceLocator
				case "fru":
					recordType = ipmi.SDRRecordTypeFRUDeviceLocator
				case "generic":
					recordType = ipmi.SDRRecordTypeGenericLocator
				default:
					CheckErr(fmt.Errorf("unkown supported record type (%s)", args[0]))
					return
				}
			}

			sdrs, err := client.GetSDRs(recordType)
			if err != nil {
				CheckErr(fmt.Errorf("GetSDRs failed, err: %s", err))
			}
			for _, sdr := range sdrs {
				fmt.Println(sdr)
			}
		},
	}

	return cmd
}

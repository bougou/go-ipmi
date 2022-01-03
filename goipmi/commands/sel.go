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
			for i, v := range selEntries {
				if i == 0 {
					fmt.Println(v.StringHeader())
				}
				fmt.Println(v.Format())
				if i == len(selEntries)-1 {
					fmt.Println(v.StringHeader())
				}
			}
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
			for i, v := range selEntries {
				var sensorName string
				if v.RecordType == ipmi.EventRecordTypeSystemEvent {
					gid := uint16(v.Default.GeneratorID)
					sn := uint8(v.Default.SensorNumber)
					sdr, ok := sdrMap[gid][sn]
					if !ok {
						sensorName = fmt.Sprintf("N/A %#04x, %#02x", gid, sn)
					} else {
						sensorName = sdr.SensorName()
					}
				}

				if i == 0 {
					fmt.Println(v.StringHeader())
				}
				fmt.Println(v.Format(), " | ", sensorName)
				if i == len(selEntries)-1 {
					fmt.Println(v.StringHeader())
				}
			}
		},
	}
	return cmd
}

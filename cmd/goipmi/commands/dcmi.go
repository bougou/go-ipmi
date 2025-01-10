package commands

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdDCMI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dcmi",
		Short: "dcmi",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}

			subcmd := args[0]
			if !contains([]string{
				"power",
				"asset_tag",
				"set_asset_tag",
				"discover",
				"get_conf_param",
				"get_mc_id_string",
				"set_mc_id_string",
				"thermalpolicy",
			}, subcmd) {
				fmt.Printf("unknown dcmi subcommand (%s)\n", subcmd)
				cmd.Help()
				return
			}
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdDCMIPower())
	cmd.AddCommand(NewCmdDCMIAssetTag())
	cmd.AddCommand(NewCmdDCMISetAssetTag())
	cmd.AddCommand(NewCmdDCMIDiscover())
	cmd.AddCommand(NewCmdDCMIGetConfigParam())
	cmd.AddCommand(NewCmdDCMISensors())
	cmd.AddCommand(NewCmdDCMIGetMCIDString())
	cmd.AddCommand(NewCmdDCMISetMCIDString())
	cmd.AddCommand(NewCmdDCMIThermalPolicy())
	cmd.AddCommand(NewCmdDCMIGetTempReading())

	return cmd
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func NewCmdDCMIDiscover() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "discover",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			dcmiCap, err := client.GetDCMICapabilities(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIPowerReading failed, err: %w", err))
			}
			fmt.Println(dcmiCap.Format())
		},
	}

	return cmd
}

func NewCmdDCMIGetConfigParam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_conf_param",
		Short: "get_conf_param",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			dcmiConfig, err := client.GetDCMIConfig(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIConfig failed, err: %w", err))
			}
			fmt.Println(dcmiConfig.Format())
		},
	}

	return cmd
}

func NewCmdDCMIPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "power",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(newCmdDCMIPowerRead())
	cmd.AddCommand(newCmdDCMIPowerGetLimit())
	cmd.AddCommand(newCmdDCMIPowerSetLimit())
	cmd.AddCommand(newCmdDCMIPowerActivate())
	cmd.AddCommand(newCmdDCMIPowerDeactivate())
	return cmd
}

func newCmdDCMIPowerRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reading",
		Short: "reading",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			resp, err := client.GetDCMIPowerReading(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIPowerReading failed, err: %w", err))
			}
			fmt.Println(resp.Format())
		},
	}
	return cmd
}

func newCmdDCMIPowerGetLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_limit",
		Short: "get_limit",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			resp, err := client.GetDCMIPowerLimit(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIPowerLimit failed, err: %w", err))
			}
			fmt.Println(resp.Format())
		},
	}
	return cmd
}

func newCmdDCMIPowerSetLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use: `set_limit <parameter> <value>

set_limit <parameter> <value>
    action        <no_action | sel_logging | power_off>
    limit         <number in Watts>
    correction    <number in milliseconds>
    sample        <number in seconds>
`,
		Short: "set_limit",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				cmd.Help()
				return
			}

			param := args[0]
			value := args[1]

			if !contains([]string{"action", "limit", "correction", "sample"}, param) {
				cmd.Help()
				return
			}

			ctx := context.Background()
			req := &ipmi.SetDCMIPowerLimitRequest{}

			{
				resp, err := client.GetDCMIPowerLimit(ctx)
				if err != nil {
					CheckErr(fmt.Errorf("GetDCMIPowerLimit failed, err: %w", err))
				}

				req.ExceptionAction = resp.ExceptionAction
				req.PowerLimitRequested = resp.PowerLimitRequested
				req.CorrectionTimeLimitMilliSec = resp.CorrectionTimeLimitMilliSec
				req.StatisticsSamplingPeriodSec = resp.StatisticsSamplingPeriodSec
			}

			switch param {
			case "action":
				if value == "no_action" {
					req.ExceptionAction = ipmi.DCMIExceptionAction_NoAction
				} else if value == "sel_logging" {
					req.ExceptionAction = ipmi.DCMIExceptionAction_LogSEL
				} else if value == "power_off" {
					req.ExceptionAction = ipmi.DCMIExceptionAction_PowerOffAndLogSEL
				} else {
					CheckErr(fmt.Errorf("invalid value for parameter action: %s", value))
				}

			case "limit":
				limit, err := parseStringToInt64(value)
				if err != nil {
					CheckErr(fmt.Errorf("invalid value for parameter limit: %s", value))
				}
				req.PowerLimitRequested = uint16(limit)

			case "correction":
				correction, err := parseStringToInt64(value)
				if err != nil {
					CheckErr(fmt.Errorf("invalid value for parameter correction: %s", value))
				}
				req.CorrectionTimeLimitMilliSec = uint32(correction)

			case "sample":
				sample, err := parseStringToInt64(value)
				if err != nil {
					CheckErr(fmt.Errorf("invalid value for parameter sample: %s", value))
				}
				req.StatisticsSamplingPeriodSec = uint16(sample)

			default:
				CheckErr(fmt.Errorf("invalid parameter: %s", param))
			}

			_, err := client.SetDCMIPowerLimit(ctx, req)
			if err != nil {
				CheckErr(fmt.Errorf("SetDCMIPowerLimit failed, err: %w", err))
			}

			resp, err := client.GetDCMIPowerLimit(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIPowerLimit failed, err: %w", err))
			}

			fmt.Println(resp.Format())
		},
	}
	return cmd
}

func newCmdDCMIPowerActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "activate",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			_, err := client.ActivateDCMIPowerLimit(ctx, true)
			if err != nil {
				CheckErr(fmt.Errorf("ActivateDCMIPowerLimit (activate) failed, err: %w", err))
			}
			fmt.Println("Power limit successfully activated")
		},
	}
	return cmd
}

func newCmdDCMIPowerDeactivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "deactivate",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			_, err := client.ActivateDCMIPowerLimit(ctx, false)
			if err != nil {
				CheckErr(fmt.Errorf("ActivateDCMIPowerLimit (deactivate) failed, err: %w", err))
			}
			fmt.Println("Power limit successfully deactivated")
		},
	}
	return cmd
}

func NewCmdDCMIAssetTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset_tag",
		Short: "asset_tag",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			assetTagRaw, typeLength, err := client.GetDCMIAssetTagFull(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIAssetTagFull failed, err: %w", err))
			}

			assetTag, err := typeLength.Chars(assetTagRaw)
			if err != nil {
				CheckErr(fmt.Errorf("convert raw to chars failed, err: %w", err))

			}
			fmt.Printf("Asset tag: %s\nTypeLength: %s\n", assetTag, typeLength)
		},
	}
	return cmd
}

func NewCmdDCMISetAssetTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_asset_tag [asset_tag]",
		Short: "set_asset_tag",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}

			var assetTag = []byte(args[0])
			ctx := context.Background()

			if err := client.SetDCMIAssetTagFull(ctx, assetTag); err != nil {
				CheckErr(fmt.Errorf("SetDCMIAssetTagFull failed, err: %w", err))
			}
		},
	}
	return cmd
}

func NewCmdDCMISensors() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensors",
		Short: "sensors",
		Run: func(cmd *cobra.Command, args []string) {

			const EntityID_DCMI_Inlet ipmi.EntityID = 0x40
			const EntityID_DCMI_CPU ipmi.EntityID = 0x41
			const EntityID_DCMI_Baseboard ipmi.EntityID = 0x42

			ctx := context.Background()

			{
				sdrs, err := client.GetDCMISensors(ctx, EntityID_DCMI_Inlet)
				if err != nil {
					CheckErr(fmt.Errorf("GetDCMISensors for entityID (%#02x) failed, err: %s", EntityID_DCMI_Inlet, err))
				}
				fmt.Printf("Inlet: %d temperature sensors found\n", len(sdrs))
				if len(sdrs) > 0 {
					fmt.Println(ipmi.FormatSDRs(sdrs))
				}
			}

			{
				sdrs, err := client.GetDCMISensors(ctx, EntityID_DCMI_CPU)
				if err != nil {
					CheckErr(fmt.Errorf("GetDCMISensors for entityID (%#02x) failed, err: %s", EntityID_DCMI_Inlet, err))
				}
				fmt.Printf("CPU: %d temperature sensors found\n", len(sdrs))
				if len(sdrs) > 0 {
					fmt.Println(ipmi.FormatSDRs(sdrs))
				}
			}

			{
				sdrs, err := client.GetDCMISensors(ctx, EntityID_DCMI_Baseboard)
				if err != nil {
					CheckErr(fmt.Errorf("GetDCMISensors for entityID (%#02x) failed, err: %s", EntityID_DCMI_Inlet, err))
				}
				fmt.Printf("Baseboard: %d temperature sensors found\n", len(sdrs))
				if len(sdrs) > 0 {
					fmt.Println(ipmi.FormatSDRs(sdrs))
				}
			}

		},
	}

	return cmd
}

func NewCmdDCMIGetMCIDString() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_mc_id_string",
		Short: "get_mc_id_string",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			id, err := client.GetDCMIMgmtControllerIdentifierFull(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIMgmtControllerIdentifierFull failed, err: %w", err))
			}

			fmt.Printf("Management Controller Identifier String: %s\n", id)
		},
	}
	return cmd
}

func NewCmdDCMISetMCIDString() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_mc_id_string [id_str]",
		Short: "set_mc_id_string",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}

			var idStr = []byte(args[0])
			ctx := context.Background()

			if err := client.SetDCMIMgmtControllerIdentifierFull(ctx, idStr); err != nil {
				CheckErr(fmt.Errorf("SetDCMIMgmtControllerIdentifierFull failed, err: %w", err))
			}
		},
	}
	return cmd
}

func NewCmdDCMIThermalPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thermalpolicy",
		Short: "thermalpolicy",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(newCmdDCMIThermalPolicyGet())
	cmd.AddCommand(newCmdDCMIThermalPolicySet())
	return cmd
}

func newCmdDCMIThermalPolicyGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <entityID> <instanceID>",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				cmd.Help()
				return
			}

			entityID, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("parse entityID (%s) failed, err: %s", args[0], err))
			}

			entityInstance, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("parse entityInstance (%s) failed, err: %s", args[1], err))
			}

			ctx := context.Background()
			resp, err := client.GetDCMIThermalLimit(ctx, ipmi.EntityID(entityID), ipmi.EntityInstance(entityInstance))
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMIThermalLimit failed, err: %w", err))
			}
			fmt.Println(resp.Format())
		},
	}
	return cmd
}

func newCmdDCMIThermalPolicySet() *cobra.Command {
	cmd := &cobra.Command{
		Use: `set <entityID> <instanceID> <volatile-param> <poweroff-param> <sel-param> <temperatureLimit> <exceptionTime>

thermalpolicy instance parameters:
		valid volatile parameters: <volatile/nonvolatile/disabled>
		valid poweroff parameters: <poweroff/nopoweroff/disabled>
		valid sel parameters:      <sel/nosel/disabled>
		                           <temperatureLimit>
		                           <exceptionTime>

    volatile       Current Power Cycle
    nonvolatile    Set across power cycles
    poweroff       Hard Power Off system
    nopoweroff     No 'Hard Power Off' action
    sel            Log event to SEL
    nosel          No 'Log event to SEL' action
    disabled       Disabled`,

		Short: "set",

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 7 {
				cmd.Help()
				return
			}

			entityID, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("parse entityID (%s) failed, err: %s", args[0], err))
			}

			entityInstance, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("parse entityInstance (%s) failed, err: %s", args[1], err))
			}

			var powerOff bool
			powerParam := args[3]
			switch powerParam {
			case "poweroff":
				powerOff = true
			case "nopoweroff", "disabled":
				powerOff = false
			case "default":
				CheckErr(fmt.Errorf("invalid poweroff parameter: %s", powerParam))
			}

			var sel bool
			selParam := args[4]
			switch selParam {
			case "sel":
				sel = true
			case "nosel", "disabled":
				sel = false
			case "default":
				CheckErr(fmt.Errorf("invalid sel parameter: %s", selParam))
			}

			temperatureLimit, err := parseStringToInt64(args[5])
			if err != nil {
				CheckErr(fmt.Errorf("parse temperatureLimit (%s) failed, err: %s", args[5], err))
			}
			exceptionTime, err := parseStringToInt64(args[6])
			if err != nil {
				CheckErr(fmt.Errorf("parse exceptionTime (%s) failed, err: %s", args[6], err))
			}

			req := &ipmi.SetDCMIThermalLimitRequest{
				EntityID:                          ipmi.EntityID(entityID),
				EntityInstance:                    ipmi.EntityInstance(entityInstance),
				ExceptionAction_PowerOffAndLogSEL: powerOff,
				ExceptionAction_LogSELOnly:        sel,
				TemperatureLimit:                  uint8(temperatureLimit),
				ExceptionTimeSec:                  uint16(exceptionTime),
			}

			ctx := context.Background()
			if _, err := client.SetDCMIThermalLimit(ctx, req); err != nil {
				CheckErr(fmt.Errorf("SetDCMIThermalLimit failed, err: %w", err))
			}

			fmt.Println("SetDCMIThermalLimit succeeded:")
		},
	}
	return cmd
}

func NewCmdDCMIGetTempReading() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_temp_reading",
		Short: "get_temp_reading",
		Run: func(cmd *cobra.Command, args []string) {

			const EntityID_DCMI_Inlet ipmi.EntityID = 0x40
			const EntityID_DCMI_CPU ipmi.EntityID = 0x41
			const EntityID_DCMI_Baseboard ipmi.EntityID = 0x42

			ctx := context.Background()

			readings, err := client.GetDCMITemperatureReadingsForEntities(ctx, EntityID_DCMI_Inlet, EntityID_DCMI_CPU, EntityID_DCMI_Baseboard)
			if err != nil {
				CheckErr(fmt.Errorf("GetDCMISensors for entityID (%#02x) failed, err: %s", EntityID_DCMI_Inlet, err))
			}

			fmt.Printf("Got: %d temperature readings found\n", len(readings))
			fmt.Println(ipmi.FormatDCMITemperatureReadings(readings))
		},
	}

	return cmd
}

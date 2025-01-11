package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bougou/go-ipmi"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

const timeFormat = time.RFC3339

// x Experimental commands.
func NewCmdX() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "x",
		Short: "x",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdXGetSDRs())
	cmd.AddCommand(NewCmdXGetSensors())
	cmd.AddCommand(NewCmdXGetDeviceSDRs())
	cmd.AddCommand(NewCmdXGetPayloadActivationStatus())
	cmd.AddCommand(NewCmdXGetDeviceGUID())
	cmd.AddCommand(NewCmdXGetSystemGUID())
	cmd.AddCommand(NewCmdXGetPEFConfigParams())
	cmd.AddCommand(NewCmdXGetLanConfigParamsFor())
	cmd.AddCommand(NewCmdXGetLanConfigParamsFull())
	cmd.AddCommand(NewCmdXGetLanConfig())
	cmd.AddCommand(NewCmdXGetDCMIConfigParams())
	cmd.AddCommand(NewCmdXGetBootOptions())
	cmd.AddCommand(NewCmdXGetSystemInfoParams())
	cmd.AddCommand(NewCmdXGetSystemInfoParamsFor())
	cmd.AddCommand(NewCmdXGetSystemInfo())

	return cmd
}

func NewCmdXGetSDRs() *cobra.Command {
	var show bool
	var loop bool

	cmd := &cobra.Command{
		Use:   "get-sdr",
		Short: "get-sdr",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			for {
				fmt.Printf("\n\nGet SDRs at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSDRs(ctx)
				if err != nil {
					fmt.Printf("GetSDRs failed, err: %w", err)
					if loop {
						goto WAIT
					} else {
						return
					}
				}
				fmt.Printf("GetSDRs succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSDRs(res))
				}

				if loop {
					goto WAIT
				} else {
					return
				}

			WAIT:
				fmt.Println("wait for next loop")
				time.Sleep(30 * time.Second)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")
	cmd.PersistentFlags().BoolVarP(&loop, "loop", "l", false, "loop")

	return cmd
}

func NewCmdXGetSensors() *cobra.Command {
	var show bool
	var loop bool

	cmd := &cobra.Command{
		Use:   "get-sensor",
		Short: "get-sensor",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			for {
				fmt.Printf("\n\nGet Sensors at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSensors(ctx)
				if err != nil {
					fmt.Printf("GetSensors failed, err: %w", err)
					if loop {
						goto WAIT
					} else {
						return
					}
				}
				fmt.Printf("GetSensors succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSensors(true, res...))
				}
				if loop {
					goto WAIT
				} else {
					return
				}

			WAIT:
				fmt.Println("wait for next loop")
				time.Sleep(30 * time.Second)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")
	cmd.PersistentFlags().BoolVarP(&loop, "loop", "l", false, "loop")

	return cmd
}

func NewCmdXGetDeviceSDRs() *cobra.Command {
	var show bool

	cmd := &cobra.Command{
		Use:   "get-device-sdr",
		Short: "get-device-sdr",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			fmt.Printf("\n\nGet Device SDR at %s\n", time.Now().Format(timeFormat))
			res, err := client.GetDeviceSDRs(ctx)
			if err != nil {
				fmt.Printf("GetDeviceSDRs failed, err: %w", err)
				return
			}

			fmt.Printf("GetDeviceSDRs succeeded, %d records\n", len(res))
			if show {
				fmt.Println(ipmi.FormatSDRs(res))
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")

	return cmd
}

func NewCmdXGetPayloadActivationStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-payload-activation-status",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("usage: get-payload-activation-status {payload-type}")
				return
			}

			payloadType, err := parseStringToInt64(args[0])
			if err != nil {
				fmt.Println(err)
			}

			ctx := context.Background()
			res, err := client.GetPayloadActivationStatus(ctx, ipmi.PayloadType(payloadType))
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(res.Format())
		},
	}

	return cmd
}

func NewCmdXGetSystemGUID() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-system-guid",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetSystemGUID(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(res.Format())

			fmt.Println("\nDetail of GUID\n==============")
			fmt.Println()
			fmt.Println(ipmi.FormatGUIDDetails(res.GUID))
		},
	}

	return cmd
}

func NewCmdXGetDeviceGUID() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-device-guid",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetDeviceGUID(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(res.Format())

			fmt.Println("\nDetail of GUID\n==============")
			fmt.Println()
			fmt.Println(ipmi.FormatGUIDDetails(res.GUID))
		},
	}

	return cmd
}

func NewCmdXGetPEFConfigParams() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-pef-config-params",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			pefConfigParams, err := client.GetPEFConfigParams(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(pefConfigParams.Format())
		},
	}

	return cmd
}

func NewCmdXGetLanConfigParamsFor() *cobra.Command {
	usage := `
	get-lan-config-params-for [<channel number>]
	`

	cmd := &cobra.Command{
		Use:   "get-lan-config-params-for",
		Short: "get-lan-config-params-for",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()

			lanConfigParams := ipmi.LanConfigParams{
				IP:               &ipmi.LanConfigParam_IP{},
				SubnetMask:       &ipmi.LanConfigParam_SubnetMask{},
				DefaultGatewayIP: &ipmi.LanConfigParam_DefaultGatewayIP{},
			}

			if err := client.GetLanConfigParamsFor(ctx, channelNumber, &lanConfigParams); err != nil {
				CheckErr(fmt.Errorf("GetLanConfigParamsFor failed, err: %w", err))
			}

			client.Debug("Lan Config", lanConfigParams)

			fmt.Println(lanConfigParams.Format())
		},
	}
	return cmd
}

func NewCmdXGetLanConfigParamsFull() *cobra.Command {
	usage := `
	get-lan-config-params-full [<channel number>]
	`

	cmd := &cobra.Command{
		Use:   "get-lan-config-params-full",
		Short: "get-lan-config-params-full",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()

			lanConfigParams, err := client.GetLanConfigParamsFull(ctx, channelNumber)
			if err != nil {
				CheckErr(fmt.Errorf("GetLanConfigParamsFull failed, err: %w", err))
			}

			client.Debug("Lan Config", lanConfigParams)
			fmt.Println(lanConfigParams.Format())
		},
	}
	return cmd

}

func NewCmdXGetLanConfig() *cobra.Command {
	usage := `
	get-lan-config [<channel number>]
	`

	cmd := &cobra.Command{
		Use: "get-lan-config",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %w", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()
			lanConfig, err := client.GetLanConfig(ctx, channelNumber)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(lanConfig.Format())
		},
	}

	return cmd
}

func NewCmdXGetDCMIConfigParams() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-dcmi-config-params",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			dcmiConfigParams, err := client.GetDCMIConfigParams(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(dcmiConfigParams.Format())
		},
	}

	return cmd
}

func NewCmdXGetBootOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-boot-options",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			bootOptionsParams, err := client.GetSystemBootOptionsParams(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(bootOptionsParams.Format())
		},
	}

	return cmd
}

func NewCmdXGetSystemInfoParams() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-system-info-params",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			systemInfoParams, err := client.GetSystemInfoParams(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(systemInfoParams.Format())
		},
	}

	return cmd
}

func NewCmdXGetSystemInfoParamsFor() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-system-info-params-for",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			systemInfoParams := &ipmi.SystemInfoParams{
				SetInProgress:          nil,
				SystemFirmwareVersions: nil,
				SystemNames:            make([]*ipmi.SystemInfoParam_SystemName, 0),
			}
			pretty.Println(systemInfoParams)
			if err := client.GetSystemInfoParamsFor(ctx, systemInfoParams); err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(systemInfoParams.Format())
		},
	}

	return cmd
}

func NewCmdXGetSystemInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-system-info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			systemInfo, err := client.GetSystemInfo(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(systemInfo.Format())
		},
	}

	return cmd
}

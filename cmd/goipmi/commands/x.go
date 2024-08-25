package commands

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

const timeFormat = time.RFC3339

func NewCmdX() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "x",
		Short: "x",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdXGetSDR())
	cmd.AddCommand(NewCmdXGetSensor())
	cmd.AddCommand(NewCmdXGetPayloadActivationStatus())
	cmd.AddCommand(NewCmdXGetDeviceGUID())
	cmd.AddCommand(NewCmdXGetSystemGUID())
	cmd.AddCommand(NewCmdXGetPEFConfigSystemUUID())

	return cmd
}

func NewCmdXGetSDR() *cobra.Command {
	var show bool

	cmd := &cobra.Command{
		Use:   "get-sdr",
		Short: "get-sdr",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				fmt.Printf("\n\nGet SDR at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSDRs()
				if err != nil {
					fmt.Printf("GetSDRs failed, err: %s", err)
					goto WAIT
				}
				fmt.Printf("GetSDRs succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSDRs(res))
				}
				goto WAIT

			WAIT:
				time.Sleep(30 * time.Second)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")

	return cmd
}

func NewCmdXGetSensor() *cobra.Command {
	var show bool

	cmd := &cobra.Command{
		Use:   "get-sensor",
		Short: "get-sensor",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				fmt.Printf("\n\nGet Sensors at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSensors()
				if err != nil {
					fmt.Printf("GetSensors failed, err: %s", err)
					goto WAIT
				}
				fmt.Printf("GetSensors succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSensors(true, res...))
				}
				goto WAIT

			WAIT:
				time.Sleep(30 * time.Second)
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

			res, err := client.GetPayloadActivationStatus(ipmi.PayloadType(payloadType))
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
			res, err := client.GetSystemGUID()
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
			res, err := client.GetDeviceGUID()
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

func NewCmdXGetPEFConfigSystemUUID() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-pef-config-system-uuid",
		Run: func(cmd *cobra.Command, args []string) {
			param, err := client.GetPEFConfigParameters_SystemUUID()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(param.Format())
		},
	}

	return cmd
}

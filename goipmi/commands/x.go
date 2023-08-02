package commands

import (
	"fmt"
	"time"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

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

	return cmd
}

func NewCmdXGetSDR() *cobra.Command {
	var show bool

	cmd := &cobra.Command{
		Use:   "getsdr",
		Short: "getsdr",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				fmt.Printf("\n\nGet SDR at %s\n", time.Now().Format(time.RFC3339))
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
		Use:   "getsensor",
		Short: "getsensor",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				fmt.Printf("\n\nGet Sensors at %s\n", time.Now().Format(time.RFC3339))
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

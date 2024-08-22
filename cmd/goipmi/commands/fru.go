package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdFRU() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fru",
		Short: "fru",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdFRUPrint())

	return cmd
}

func NewCmdFRUPrint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "print",
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) < 1 {
				frus, err := client.GetFRUs()
				if err != nil {
					CheckErr(fmt.Errorf("GetFRUs failed, err: %s", err))
				}

				for _, fru := range frus {
					fmt.Println(fru.String())
				}
			} else {
				id, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid FRU Device ID passed, err: %s", err))
				}
				fruID := uint8(id)

				fru, err := client.GetFRU(fruID, "")
				if err != nil {
					CheckErr(fmt.Errorf("GetFRU failed, err: %s", err))
				}
				fmt.Println(fru.String())
			}

		},
	}
	return cmd
}

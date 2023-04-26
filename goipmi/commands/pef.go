package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdPEF() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pef",
		Short: "pef",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdPEFCapabilities())
	cmd.AddCommand(NewCmdPEFStatus())
	cmd.AddCommand(NewCmdPEFInfo())

	return cmd
}

func NewCmdPEFCapabilities() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetPEFCapabilities()
			if err != nil {
				CheckErr(fmt.Errorf("GetPEFCapabilities failed, err: %s", err))
			}

			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdPEFStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "status",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetLastProcessedEventId()
			if err != nil {
				CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %s", err))
			}

			fmt.Println(res.Format())
		},
	}
	return cmd
}

// Reports PEF capabilities + System GUID
func NewCmdPEFInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			systemGUID, err := client.GetPEFConfigSystemUUID()
			if err != nil {
				CheckErr(fmt.Errorf("GetPEFConfigSystemUUID failed, err: %s", err))
			}

			fmt.Println(systemGUID)
		},
	}
	return cmd
}

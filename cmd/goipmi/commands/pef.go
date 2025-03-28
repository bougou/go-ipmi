package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bougou/go-ipmi"
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
			if len(args) < 2 {
				cmd.Help()
				return
			}

			if !contains([]string{
				"capabilities",
				"status",
				"filter",
				"info",
				"policy",
			}, args[1]) {
				cmd.Help()
				return
			}
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdPEFCapabilities())
	cmd.AddCommand(NewCmdPEFStatus())
	cmd.AddCommand(NewCmdPEFFilter())
	cmd.AddCommand(NewCmdPEFInfo())
	cmd.AddCommand(NewCmdPEFPolicy())

	return cmd
}

func NewCmdPEFCapabilities() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetPEFCapabilities(ctx)
			if err != nil {
				CheckErr(fmt.Errorf("GetPEFCapabilities failed, err: %w", err))
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
			ctx := context.Background()

			{
				res, err := client.GetLastProcessedEventId(ctx)
				if err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %w", err))
				}
				fmt.Println(res.Format())
			}

			{
				param := &ipmi.PEFConfigParam_Control{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %w", err))
				}
				fmt.Println(param.Format())
			}

			{
				param := &ipmi.PEFConfigParam_ActionGlobalControl{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %w", err))
				}
				fmt.Println(param.Format())
			}

		},
	}
	return cmd
}

func NewCmdPEFFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "filter",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(NewCmdPEFFilterList())
	return cmd
}

func NewCmdPEFFilterList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			var numberOfEventFilters uint8

			{
				param := &ipmi.PEFConfigParam_EventFiltersCount{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(fmt.Errorf("get number of event filters failed, err: %w", err))
				}
				numberOfEventFilters = param.Value
			}

			var eventFilters = make([]*ipmi.PEFEventFilter, numberOfEventFilters)
			for i := uint8(0); i < numberOfEventFilters; i++ {
				// 1-based
				filterNumber := i + 1

				param := &ipmi.PEFConfigParam_EventFilter{
					SetSelector: filterNumber,
				}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(fmt.Errorf("get event filter entry %d failed, err: %w", filterNumber, err))
				}

				eventFilters[i] = param.Filter
			}

			fmt.Println(ipmi.FormatEventFilters(eventFilters))
		},
	}
	return cmd
}

func NewCmdPEFInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			{
				param := &ipmi.PEFConfigParam_SystemGUID{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(err)
				}
				fmt.Println(param.Format())

				if !param.UseGUID {
					res, err := client.GetSystemGUID(ctx)
					if err != nil {
						CheckErr(err)
					}
					fmt.Println(ipmi.FormatGUIDDetails(res.GUID))
				}
			}

			{
				param := &ipmi.PEFConfigParam_AlertPoliciesCount{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(err)
				}
				fmt.Println(param.Format())
			}

			{
				res, err := client.GetPEFCapabilities(ctx)
				if err != nil {
					CheckErr(err)
				}
				fmt.Println(res.Format())
			}

		},
	}
	return cmd
}

func NewCmdPEFPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "policy",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(NewCmdPEFPolicyList())
	return cmd
}

func NewCmdPEFPolicyList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {

			ctx := context.Background()

			numberOfAlertPolicies := uint8(0)
			{
				param := &ipmi.PEFConfigParam_AlertPoliciesCount{}
				if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
					CheckErr(err)
				}
				numberOfAlertPolicies = param.Value
			}

			{

				headers := []string{
					"Entry",
					"PolicyNumber",
					"PolicyState",
					"PolicyAction",
					"Channel",
					"ChannelMedium",
					"Destination",
					"IsEventSpecific",
					"AlertStringKey",
				}
				rows := make([][]string, numberOfAlertPolicies)
				for i := uint8(0); i < numberOfAlertPolicies; i++ {
					entry := i + 1

					param := &ipmi.PEFConfigParam_AlertPolicy{
						SetSelector: entry,
					}
					if err := client.GetPEFConfigParamFor(ctx, param); err != nil {
						CheckErr(fmt.Errorf("get alert policy number (%d) failed, err: %w", entry, err))
					}
					alertPolicy := param.Policy

					channelNumber := param.Policy.ChannelNumber
					resp, err := client.GetChannelInfo(ctx, channelNumber)
					if err != nil {
						CheckErr(fmt.Errorf("get channel info (%d) failed, err: %w", channelNumber, err))
					}

					// Todo Get Lan Config Param
					// if resp.ChannelMedium == ipmi.ChannelMediumLAN {
					// Number of Destinations : 0x11 (17)
					// Destination Type : 0x12 (18)
					// Community String : 0x10 (16)
					// Destination Address: 0x13 (19)
					// }

					row := []string{
						strconv.Itoa(int(entry)),
						fmt.Sprintf("%d", alertPolicy.PolicyNumber),
						fmt.Sprintf("%v", alertPolicy.PolicyState),
						alertPolicy.PolicyAction.ShortString(),
						fmt.Sprintf("%d", alertPolicy.ChannelNumber),
						resp.ChannelMedium.String(),
						fmt.Sprintf("%d", alertPolicy.Destination),
						fmt.Sprintf("%v", alertPolicy.IsEventSpecific),
						fmt.Sprintf("%d", alertPolicy.AlertStringKey),
					}

					rows[i] = row
				}

				fmt.Println(formatTable(headers, rows))
			}
		},
	}
	return cmd
}

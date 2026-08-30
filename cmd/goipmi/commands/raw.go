package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bougou/go-ipmi/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdRaw() *cobra.Command {
	usage := `raw <netfn 0xnn> <cmd 0xnn> [data 0xnn ...]`

	cmd := &cobra.Command{
		Use:   "raw",
		Short: "raw",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			raw, err := parseRawBytes(args)
			if err != nil {
				CheckErr(err)
			}

			netFn := types.NetFn(raw[0])
			rawCmd := raw[1]
			data := raw[2:]

			ctx := context.Background()
			res, err := client.RawCommand(ctx, netFn, rawCmd, data, "Raw")
			if err != nil {
				CheckErr(fmt.Errorf("RawCommand failed, err: %w", err))
			}

			fmt.Println(formatRawResponse(res.Response))
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}

	return cmd
}

func parseRawBytes(args []string) ([]byte, error) {
	out := make([]byte, 0, len(args))

	for _, arg := range args {
		v, err := parseStringToInt64(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value (%s) passed, err: %w", arg, err)
		}
		if v < 0 || v > 0xff {
			return nil, fmt.Errorf("invalid byte value (%s) passed, out of range 0x00-0xff", arg)
		}
		out = append(out, byte(v))
	}

	return out, nil
}

func formatRawResponse(data []byte) string {
	hexStrings := make([]string, len(data))
	for i, b := range data {
		hexStrings[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(hexStrings, " ")
}

//go:build !goipmi_notablewriter

package types

import (
	"bytes"
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// RenderTable formats a table from a slice of rows.
// The Headers is a slice of strings, the order of the headers is the order of the columns in the table.
// Each row is represented as a map, the keys of the map are the headers of the table.
func RenderTable(headers []string, rows []map[string]string) string {
	var buf = new(bytes.Buffer)

	table := tablewriter.NewTable(buf)

	table.Options(
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignRight),
		tablewriter.WithRowAutoWrap(0), // Disable auto wrap
		tablewriter.WithFooterAlignmentConfig(tw.CellAlignment{
			Global: tw.AlignCenter, // Center align footer like header
		}),
		tablewriter.WithFooterAutoFormat(tw.On),
	)

	table.Header(headers)
	table.Footer(headers)

	// Add rows
	for _, _row := range rows {
		row := make([]any, len(headers))
		for i, header := range headers {
			row[i] = _row[header]
		}
		table.Append(row)
	}

	table.Render()
	return buf.String()
}

func RenderTableStream(headers []string, rowSeq iter.Seq[map[string]string]) error {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))

	table.Options(
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignRight),
		tablewriter.WithRowAutoWrap(0), // Disable auto wrap
		tablewriter.WithFooterAlignmentConfig(tw.CellAlignment{
			Global: tw.AlignCenter, // Center align footer like header
		}),
		tablewriter.WithFooterAutoFormat(tw.On),
	)

	// Start streaming
	if err := table.Start(); err != nil {
		return fmt.Errorf("table start failed: %w", err)
	}
	defer table.Close()

	table.Header(headers)

	for _row := range rowSeq {
		row := make([]any, len(headers))
		for i, header := range headers {
			canonicalHeader := strings.TrimSpace(header)
			canonicalHeader = strings.Trim(canonicalHeader, "-")
			canonicalHeader = strings.TrimSpace(canonicalHeader)
			row[i] = _row[canonicalHeader]
		}
		table.Append(row)
	}

	table.Footer(headers)

	return nil

}

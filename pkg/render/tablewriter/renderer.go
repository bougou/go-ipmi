// Package tablewriter provides the bordered table renderer used by goipmi.
package tablewriter

import (
	"bytes"
	"fmt"
	"io"
	"iter"

	"github.com/bougou/go-ipmi/pkg/render"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// Renderer formats tables with ASCII borders, aligned cells and a footer.
type Renderer struct{}

func configureTable(table *tablewriter.Table) {
	table.Options(
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		tablewriter.WithRowAlignment(tw.AlignRight),
		tablewriter.WithRowAutoWrap(0), // Disable auto wrap
		tablewriter.WithFooterAlignmentConfig(tw.CellAlignment{
			Global: tw.AlignCenter, // Center align footer like header
		}),
		tablewriter.WithFooterAutoFormat(tw.On),
	)
}

func (Renderer) Render(columns []render.Column, rows []map[string]string) string {
	names := columnNames(columns)
	var buf = new(bytes.Buffer)

	table := tablewriter.NewTable(buf)
	configureTable(table)

	table.Header(names)
	table.Footer(names)

	for _, _row := range rows {
		row := make([]any, len(columns))
		for i, c := range columns {
			row[i] = _row[c.Name]
		}
		table.Append(row)
	}

	table.Render()
	return buf.String()
}

func (Renderer) RenderStream(w io.Writer, columns []render.Column, rowSeq iter.Seq[map[string]string]) (err error) {
	names := columnNames(columns)
	table := tablewriter.NewTable(w, tablewriter.WithStreaming(tw.StreamConfig{Enable: true}))
	configureTable(table)
	table.Options(tablewriter.WithColumnWidths(columnWidths(columns)))

	if err := table.Start(); err != nil {
		return fmt.Errorf("table start failed: %w", err)
	}
	defer func() {
		closeErr := table.Close()
		if err == nil {
			err = closeErr
		}
	}()

	table.Header(names)

	for _row := range rowSeq {
		row := make([]any, len(columns))
		for i, c := range columns {
			row[i] = _row[c.Name]
		}
		if err := table.Append(row); err != nil {
			return err
		}
	}

	table.Footer(names)

	return nil
}

func columnNames(columns []render.Column) []string {
	names := make([]string, len(columns))
	for i, c := range columns {
		names[i] = c.Name
	}
	return names
}

func columnWidths(columns []render.Column) tw.Mapper[int, int] {
	widths := tw.Mapper[int, int]{}
	for i, c := range columns {
		// tablewriter's width includes one padding space per side; keep at
		// least its minimum.
		w := c.ContentWidth() + 2
		if w < tw.MinimumColumnWidth {
			w = tw.MinimumColumnWidth
		}
		widths[i] = w
	}
	return widths
}

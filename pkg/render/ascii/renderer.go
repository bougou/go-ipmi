// Package ascii renders left-aligned tables with ASCII borders.
package ascii

import (
	"io"
	"iter"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/bougou/go-ipmi/pkg/render"
)

// Renderer sizes batch columns to their contents. Streaming columns use the
// declared widths and wrap longer values by rune.
type Renderer struct{}

var whitespace = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ", "\v", " ", "\f", " ")

func clean(value string) string { return whitespace.Replace(strings.TrimSpace(value)) }

func (Renderer) Render(columns []render.Column, rows []map[string]string) string {
	names := columnNames(columns)
	widths := headerWidths(columns)
	for _, row := range rows {
		for i, c := range columns {
			widths[i] = max(widths[i], utf8.RuneCountInString(clean(row[c.Name])))
		}
	}
	var out strings.Builder
	writeTable(&out, names, widths, slices.Values(rows)) // strings.Builder writes always succeed.
	return out.String()
}

func (Renderer) RenderStream(w io.Writer, columns []render.Column, rows iter.Seq[map[string]string]) error {
	return writeTable(w, columnNames(columns), headerWidths(columns), rows)
}

func columnNames(columns []render.Column) []string {
	names := make([]string, len(columns))
	for i, c := range columns {
		names[i] = c.Name
	}
	return names
}

func headerWidths(columns []render.Column) []int {
	widths := make([]int, len(columns))
	for i, c := range columns {
		widths[i] = max(1, c.ContentWidth())
	}
	return widths
}

func writeTable(w io.Writer, keys []string, widths []int, rows iter.Seq[map[string]string]) error {
	if len(keys) == 0 {
		return nil
	}
	var border strings.Builder
	border.WriteByte('+')
	for _, width := range widths {
		border.WriteString(strings.Repeat("-", width+2))
		border.WriteByte('+')
	}
	border.WriteByte('\n')
	rule := border.String()
	if _, err := io.WriteString(w, rule); err != nil {
		return err
	}
	if err := writeRow(w, keys, widths); err != nil {
		return err
	}
	if _, err := io.WriteString(w, rule); err != nil {
		return err
	}
	cells := make([]string, len(keys))
	for row := range rows {
		for i, key := range keys {
			cells[i] = row[key]
		}
		if err := writeRow(w, cells, widths); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, rule)
	return err
}

func writeRow(w io.Writer, cells []string, widths []int) error {
	remaining := make([][]rune, len(cells))
	for i, cell := range cells {
		remaining[i] = []rune(clean(cell))
	}
	for {
		var line strings.Builder
		line.WriteByte('|')
		more := false
		for i, runes := range remaining {
			n := min(len(runes), widths[i])
			line.WriteByte(' ')
			line.WriteString(string(runes[:n]))
			line.WriteString(strings.Repeat(" ", widths[i]-n+1))
			line.WriteByte('|')
			remaining[i] = runes[n:]
			more = more || len(remaining[i]) > 0
		}
		line.WriteByte('\n')
		if _, err := io.WriteString(w, line.String()); err != nil {
			return err
		}
		if !more {
			return nil
		}
	}
}

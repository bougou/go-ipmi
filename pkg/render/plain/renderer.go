// Package plain renders borderless, whitespace-aligned columns with
// text/tabwriter from the standard library.
package plain

import (
	"fmt"
	"io"
	"iter"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/bougou/go-ipmi/pkg/render"
)

// Renderer renders plain, borderless tables.
type Renderer struct{}

var cellWhitespace = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")

func cell(value string) string {
	return cellWhitespace.Replace(strings.TrimSpace(value))
}

func (Renderer) Render(columns []render.Column, rows []map[string]string) string {
	var out strings.Builder
	w := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	cells := make([]string, len(columns))
	for i, c := range columns {
		cells[i] = cell(c.Name)
	}
	fmt.Fprintln(w, strings.Join(cells, "\t"))
	for _, row := range rows {
		for i, c := range columns {
			cells[i] = cell(row[c.Name])
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	w.Flush() // strings.Builder writes always succeed.
	return out.String()
}

func (Renderer) RenderStream(w io.Writer, columns []render.Column, rows iter.Seq[map[string]string]) error {
	// Use fixed widths so rows stream without buffering; wider cells shift
	// the rest of the row.
	names := make([]string, len(columns))
	cells := make([]string, len(columns))
	widths := make([]int, len(columns))
	for i, c := range columns {
		names[i] = c.Name
		cells[i] = c.Name
		widths[i] = c.ContentWidth()
	}
	writeRow := func() error {
		var line strings.Builder
		for i, value := range cells {
			v := cell(value)
			line.WriteString(v)
			if i+1 < len(cells) {
				line.WriteString(strings.Repeat(" ", 2+max(0, widths[i]-utf8.RuneCountInString(v))))
			}
		}
		_, err := fmt.Fprintln(w, strings.TrimRight(line.String(), " "))
		return err
	}
	if err := writeRow(); err != nil {
		return err
	}
	for row := range rows {
		for i, name := range names {
			cells[i] = row[name]
		}
		if err := writeRow(); err != nil {
			return err
		}
	}
	return nil
}

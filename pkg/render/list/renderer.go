// Package list renders each record as a "Name : Value" block, one field per
// line, with a blank line between records.
package list

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/bougou/go-ipmi/pkg/render"
)

// Renderer renders records as key/value blocks.
type Renderer struct{}

func (Renderer) Render(columns []render.Column, rows []map[string]string) string {
	var b strings.Builder
	for i, row := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		_ = writeRecord(&b, columns, row)
	}
	return b.String()
}

func (Renderer) RenderStream(w io.Writer, columns []render.Column, rows iter.Seq[map[string]string]) error {
	first := true
	for row := range rows {
		if !first {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		first = false
		if err := writeRecord(w, columns, row); err != nil {
			return err
		}
	}
	return nil
}

func writeRecord(w io.Writer, columns []render.Column, row map[string]string) error {
	for _, c := range columns {
		if _, err := fmt.Fprintf(w, "%-24s : %s\n", c.Name, row[c.Name]); err != nil {
			return err
		}
	}
	return nil
}

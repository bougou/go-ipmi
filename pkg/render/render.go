// Package render defines the pluggable renderer used by the types.Format*With
// functions. Built-in renderers live in subpackages; write your own by
// implementing the Renderer interface.
package render

import (
	"io"
	"iter"
	"slices"
	"strings"
	"unicode/utf8"
)

// Column describes one named field of a record.
type Column struct {
	// Name is the field name used to look up row values.
	Name string
	// Width is the column width in runes; 0 means use the width of Name.
	Width int
}

// ContentWidth returns the declared Width, or the width of Name when Width is 0.
func (c Column) ContentWidth() int {
	if c.Width > 0 {
		return c.Width
	}
	return utf8.RuneCountInString(c.Name)
}

// Columns builds a []Column from names, each with zero width.
func Columns(names ...string) []Column {
	cols := make([]Column, len(names))
	for i, n := range names {
		cols[i] = Column{Name: n}
	}
	return cols
}

// Renderer renders records to text. Each row maps a column name to its string
// value; implementations must not mutate their inputs.
type Renderer interface {
	// Render formats all rows at once (may size columns to their content).
	Render(columns []Column, rows []map[string]string) string
	// RenderStream writes rows to w as they are produced.
	RenderStream(w io.Writer, columns []Column, rows iter.Seq[map[string]string]) error
}

// Func adapts plain functions into a Renderer. Either field may be nil; the
// other is derived from it.
type Func struct {
	RenderFn       func(columns []Column, rows []map[string]string) string
	RenderStreamFn func(w io.Writer, columns []Column, rows iter.Seq[map[string]string]) error
}

func (f Func) Render(columns []Column, rows []map[string]string) string {
	if f.RenderFn != nil {
		return f.RenderFn(columns, rows)
	}
	var b strings.Builder
	_ = f.RenderStream(&b, columns, slices.Values(rows))
	return b.String()
}

func (f Func) RenderStream(w io.Writer, columns []Column, rows iter.Seq[map[string]string]) error {
	if f.RenderStreamFn != nil {
		return f.RenderStreamFn(w, columns, rows)
	}
	all := make([]map[string]string, 0)
	for row := range rows {
		all = append(all, row)
	}
	_, err := io.WriteString(w, f.Render(columns, all))
	return err
}

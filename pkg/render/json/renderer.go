// Package json renders records as a pretty-printed JSON array of objects,
// preserving column order.
package json

import (
	stdjson "encoding/json"
	"io"
	"iter"
	"slices"
	"strings"

	"github.com/bougou/go-ipmi/pkg/render"
)

// Renderer renders records as a pretty-printed JSON array.
type Renderer struct{}

func (Renderer) Render(columns []render.Column, rows []map[string]string) string {
	var b strings.Builder
	_ = Renderer{}.RenderStream(&b, columns, slices.Values(rows))
	return b.String()
}

func (Renderer) RenderStream(w io.Writer, columns []render.Column, rows iter.Seq[map[string]string]) error {
	if _, err := io.WriteString(w, "["); err != nil {
		return err
	}
	first := true
	for row := range rows {
		if !first {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		}
		first = false
		if err := writeObject(w, columns, row); err != nil {
			return err
		}
	}
	if first {
		_, err := io.WriteString(w, "]\n")
		return err
	}
	_, err := io.WriteString(w, "\n]\n")
	return err
}

// writeObject writes one indented JSON object.
func writeObject(w io.Writer, columns []render.Column, row map[string]string) error {
	if _, err := io.WriteString(w, "\n  {\n"); err != nil {
		return err
	}
	for i, c := range columns {
		if i > 0 {
			if _, err := io.WriteString(w, ",\n"); err != nil {
				return err
			}
		}
		key, _ := stdjson.Marshal(c.Name)
		val, _ := stdjson.Marshal(row[c.Name])
		if _, err := io.WriteString(w, "    "+string(key)+": "+string(val)); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n  }")
	return err
}

package render

import (
	"io"
	"iter"
	"slices"
	"strings"
	"testing"
)

func TestFuncDerivesFromStream(t *testing.T) {
	f := Func{RenderStreamFn: func(w io.Writer, columns []Column, rows iter.Seq[map[string]string]) error {
		for row := range rows {
			if _, err := io.WriteString(w, row["ID"]+"\n"); err != nil {
				return err
			}
		}
		return nil
	}}
	if got := f.Render(Columns("ID"), []map[string]string{{"ID": "1"}, {"ID": "2"}}); got != "1\n2\n" {
		t.Fatalf("got %q", got)
	}
}

func TestFuncDerivesFromRender(t *testing.T) {
	f := Func{RenderFn: func(columns []Column, rows []map[string]string) string {
		return strings.Join([]string{rows[0]["ID"], rows[1]["ID"]}, ",")
	}}
	var b strings.Builder
	err := f.RenderStream(&b, Columns("ID"), slices.Values([]map[string]string{{"ID": "1"}, {"ID": "2"}}))
	if err != nil {
		t.Fatal(err)
	}
	if b.String() != "1,2" {
		t.Fatalf("got %q", b.String())
	}
}

func TestColumnContentWidth(t *testing.T) {
	if got := (Column{Name: "ID"}).ContentWidth(); got != 2 {
		t.Fatalf("name width = %d, want 2", got)
	}
	if got := (Column{Name: "ID", Width: 40}).ContentWidth(); got != 40 {
		t.Fatalf("declared width = %d, want 40", got)
	}
}

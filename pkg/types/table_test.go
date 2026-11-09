package types

import (
	"errors"
	"io"
	"iter"
	"strings"
	"testing"

	"github.com/bougou/go-ipmi/pkg/render"
	"github.com/bougou/go-ipmi/pkg/render/plain"
)

func TestDefaultTableOutput(t *testing.T) {
	got := RenderTable([]string{"Name", "ID"}, []map[string]string{{"ID": " 1 ", "Name": "\tFan \n"}})
	want := "+------+----+\n" +
		"| Name | ID |\n" +
		"+------+----+\n" +
		"| Fan  | 1  |\n" +
		"+------+----+\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPlainStreamWritesBeforeNextRow(t *testing.T) {
	var out strings.Builder
	seq := func(yield func(map[string]string) bool) {
		yield(map[string]string{"ID": " 1 ", "Name": "\tFan \n"})
		if !strings.Contains(out.String(), "1       Fan\n") {
			t.Errorf("output after first row = %q", out.String())
		}
	}
	columns := []render.Column{{Name: "ID", Width: 6}, {Name: "Name"}}
	if err := (plain.Renderer{}).RenderStream(&out, columns, seq); err != nil {
		t.Fatal(err)
	}
}

type failingRenderer struct{}

var _ render.Renderer = failingRenderer{}

func (failingRenderer) Render([]render.Column, []map[string]string) string { return "" }

func (failingRenderer) RenderStream(io.Writer, []render.Column, iter.Seq[map[string]string]) error {
	return io.ErrClosedPipe
}

func TestFormatStreamPropagatesRendererError(t *testing.T) {
	err := formatStream[int](failingRenderer{}, nil, nil, nil)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("expected closed pipe, got %v", err)
	}
}

func TestRendererExamples(t *testing.T) {
	r := plain.Renderer{}
	for _, tc := range []struct {
		name    string
		headers []string
		rows    []map[string]string
		want    string
	}{
		{name: "basic", headers: []string{"SensorName", "Reading", "Unit"},
			rows: []map[string]string{
				{"SensorName": "CPU Temp", "Reading": "45", "Unit": "degrees C"},
				{"SensorName": "Fan1", "Reading": "12000", "Unit": "RPM"},
			},
			want: "SensorName  Reading  Unit\n" +
				"CPU Temp    45       degrees C\n" +
				"Fan1        12000    RPM\n",
		},
		{name: "missing_cells", headers: []string{"A", "B", "C"},
			rows: []map[string]string{
				{"A": "1"},
				{"B": "two\nlines", "C": "  tab\there  "},
			},
			want: "A  B          C\n" +
				"1             \n" +
				"   two lines  tab here\n",
		},
		{name: "no_rows", headers: []string{"A", "B"}, want: "A  B\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := r.Render(render.Columns(tc.headers...), tc.rows); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("stream", func(t *testing.T) {
		const firstRows = "ID  SensorName          Status\n" +
			"1   CPU Temp            ok\n"
		const want = firstRows + "333  Fan1                cr\n"
		var out strings.Builder
		columns := []render.Column{{Name: "ID"}, {Name: "SensorName", Width: 18}, {Name: "Status"}}
		rows := func(yield func(map[string]string) bool) {
			if !yield(map[string]string{"ID": "1", "SensorName": "CPU Temp", "Status": "ok"}) {
				return
			}
			if got := out.String(); got != firstRows {
				t.Errorf("before second yield: got %q, want %q", got, firstRows)
			}
			yield(map[string]string{"ID": "333", "SensorName": "Fan1", "Status": "cr"})
		}
		if err := r.RenderStream(&out, columns, rows); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); got != want {
			t.Errorf("output = %q, want %q", got, want)
		}
	})
}

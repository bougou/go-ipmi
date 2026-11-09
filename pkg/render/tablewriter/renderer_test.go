package tablewriter

import (
	"slices"
	"strings"
	"testing"

	"github.com/bougou/go-ipmi/pkg/render"
)

func TestLegacyOutput(t *testing.T) {
	columns := render.Columns("ID", "Name")
	rows := []map[string]string{{"ID": "1", "Name": "Fan"}}
	var stream strings.Builder
	r := Renderer{}
	if err := r.RenderStream(&stream, columns, slices.Values(rows)); err != nil {
		t.Fatal(err)
	}

	const wantBatch = "+----+------+\n" +
		"| ID | NAME |\n" +
		"+----+------+\n" +
		"|  1 |  Fan |\n" +
		"+----+------+\n" +
		"| ID | NAME |\n" +
		"+----+------+\n"
	const wantStream = "+--------+--------+\n" +
		"|   ID   |  NAME  |\n" +
		"+--------+--------+\n" +
		"|      1 |    Fan |\n" +
		"+--------+--------+\n" +
		"|   ID   |  NAME  |\n" +
		"+--------+--------+\n"
	if got := r.Render(columns, rows); got != wantBatch {
		t.Errorf("batch = %q, want %q", got, wantBatch)
	}
	if got := stream.String(); got != wantStream {
		t.Errorf("stream = %q, want %q", got, wantStream)
	}
}

func TestRendererExamples(t *testing.T) {
	r := Renderer{}
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
			want: "+-------------+---------+-----------+\n" +
				"| SENSOR NAME | READING |   UNIT    |\n" +
				"+-------------+---------+-----------+\n" +
				"|    CPU Temp |      45 | degrees C |\n" +
				"|        Fan1 |   12000 |       RPM |\n" +
				"+-------------+---------+-----------+\n" +
				"| SENSOR NAME | READING |   UNIT    |\n" +
				"+-------------+---------+-----------+\n",
		},
		{name: "missing_cells", headers: []string{"A", "B", "C"},
			rows: []map[string]string{
				{"A": "1"},
				{"B": "two\nlines", "C": "  tab\there  "},
			},
			want: "+---+-------+---------+\n" +
				"| A |   B   |    C    |\n" +
				"+---+-------+---------+\n" +
				"| 1 |       |         |\n" +
				"|   |   two | tab\there |\n" +
				"|   | lines |         |\n" +
				"+---+-------+---------+\n" +
				"| A |   B   |    C    |\n" +
				"+---+-------+---------+\n",
		},
		{name: "no_rows", headers: []string{"A", "B"}, want: "+---+---+\n" +
			"| A | B |\n" +
			"+---+---+\n" +
			"+---+---+\n" +
			"| A | B |\n" +
			"+---+---+\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := r.Render(render.Columns(tc.headers...), tc.rows); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("stream", func(t *testing.T) {
		const firstRows = "+--------+--------------------+--------+\n" +
			"|   ID   |    SENSOR NAME     | STATUS |\n" +
			"+--------+--------------------+--------+\n" +
			"|      1 |           CPU Temp |     ok |\n"
		const want = firstRows + "|    333 |               Fan1 |     cr |\n" +
			"+--------+--------------------+--------+\n" +
			"|   ID   |    SENSOR NAME     | STATUS |\n" +
			"+--------+--------------------+--------+\n"
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

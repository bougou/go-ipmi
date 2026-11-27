package ascii_test

import (
	"encoding/hex"
	"errors"
	"io"
	"iter"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/render"
	"github.com/bougou/go-ipmi/pkg/render/ascii"
	"github.com/bougou/go-ipmi/pkg/types"
)

var _ render.Renderer = ascii.Renderer{}

func TestBatchFitsContents(t *testing.T) {
	got := (ascii.Renderer{}).Render(render.Columns("ID", "Name"), []map[string]string{{"Name": " Fan\tA ", "ID": "1"}, {"Name": "B"}})
	want := "+----+-------+\n| ID | Name  |\n+----+-------+\n| 1  | Fan A |\n|    | B     |\n+----+-------+\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStreamPreservesWideValues(t *testing.T) {
	var out strings.Builder
	prefix := "+-----+\n| X   |\n+-----+\n"
	rows := func(yield func(map[string]string) bool) {
		if out.String() != prefix {
			t.Fatalf("header = %q", out.String())
		}
		yield(map[string]string{"X": "abcdef"})
		if out.String() != prefix+"| abc |\n| def |\n" {
			t.Fatalf("first row = %q", out.String())
		}
	}
	err := (ascii.Renderer{}).RenderStream(&out, []render.Column{{Name: "X", Width: 3}}, rows)
	if err != nil || out.String() != prefix+"| abc |\n| def |\n+-----+\n" {
		t.Fatalf("output %q, error %v", out.String(), err)
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestStreamStopsAtHeaderError(t *testing.T) {
	var rows iter.Seq[map[string]string] = func(func(map[string]string) bool) { t.Fatal("unexpected iteration") }
	err := (ascii.Renderer{}).RenderStream(brokenWriter{}, render.Columns("ID"), rows)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}

type capturedStream struct {
	ascii.Renderer
	out strings.Builder
}

func (r *capturedStream) RenderStream(_ io.Writer, columns []render.Column, rows iter.Seq[map[string]string]) error {
	return r.Renderer.RenderStream(&r.out, columns, rows)
}

func TestSELStreamKeepsSensorSpecificFieldsOnOneLine(t *testing.T) {
	renderer := new(capturedStream)
	timestamp := time.Date(2012, 1, 1, 8, 18, 24, 0, time.FixedZone("CST", 8*3600))
	record := &types.SEL{RecordID: 1, RecordType: 2, Standard: &types.SELStandard{Timestamp: timestamp, GeneratorID: 0x20, SensorType: 0x0c, EventReadingType: 0x6f}}
	rows := slices.Values([]*types.Result[types.SEL]{{Ok: record}})
	if err := types.FormatSELsStreamWith(renderer, rows, nil); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(renderer.out.String(), "\n") {
		if strings.Contains(line, "0x0001") && strings.Contains(line, timestamp.String()) && strings.Contains(line, "0x0020") &&
			strings.Contains(line, "0x6f (Sensor Specific)") && strings.Contains(line, "Correctable ECC / other correctable memory error") {
			return
		}
	}
	t.Fatalf("expected intact SEL fields on one line:\n%s", renderer.out.String())
}

func TestOEMSELListings(t *testing.T) {
	for _, tc := range []struct {
		name   string
		raw    string
		fields []string
	}{
		{"timestamped", "c200c10100000000000006980a000af8",
			[]string{"0x00c2", "OEM record c1", time.Unix(1, 0).String(), "Manufacturer ID 000000", "06980a000af8"}},
		{"nonzero manufacturer", "c300df01000000341200000102030405",
			[]string{"0x00c3", "OEM record df", "Manufacturer ID 001234", "000102030405"}},
		{"non-timestamped", "c400e0000102030405060708090a00ff",
			[]string{"0x00c4", "OEM record e0", "000102030405060708090a00ff"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := hex.DecodeString(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			record, err := types.ParseSEL(raw)
			if err != nil {
				t.Fatal(err)
			}
			batch := types.FormatSELs([]*types.SEL{record}, nil)
			stream := new(capturedStream)
			rows := slices.Values([]*types.Result[types.SEL]{{Ok: record}})
			if err := types.FormatSELsStreamWith(stream, rows, nil); err != nil {
				t.Fatal(err)
			}
			for mode, output := range map[string]string{"batch": batch, "stream": stream.out.String()} {
				lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
				if len(lines) != 5 {
					t.Errorf("%s: expected one data row, got:\n%s", mode, output)
					continue
				}
				cells := strings.Split(lines[3], "|")
				for i := range cells {
					cells[i] = strings.TrimSpace(cells[i])
				}
				for _, field := range tc.fields {
					if !slices.Contains(cells, field) {
						t.Errorf("%s: missing cell %q in %s", mode, field, lines[3])
					}
				}
			}
		})
	}
}

func TestSELListingColumnOrder(t *testing.T) {
	const columns = "ID RecordType EvmRev Timestamp GID SensorNumber SensorTypeCode SensorType EventReadingType EventDescription EventDirection EventSeverity EventData"
	for _, tc := range []struct {
		name string
		sdrs types.SDRMapBySensorNumber
		want string
	}{
		{"normal", nil, columns},
		{"extended", types.SDRMapBySensorNumber{}, columns + " SensorName"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := new(capturedStream)
			if err := types.FormatSELsStreamWith(stream, slices.Values([]*types.Result[types.SEL]{}), tc.sdrs); err != nil {
				t.Fatal(err)
			}
			for mode, output := range map[string]string{"batch": types.FormatSELs(nil, tc.sdrs), "stream": stream.out.String()} {
				lines := strings.Split(output, "\n")
				if len(lines) < 2 {
					t.Fatalf("%s: missing header in %q", mode, output)
				}
				got := strings.Fields(strings.ReplaceAll(lines[1], "|", " "))
				if !slices.Equal(got, strings.Fields(tc.want)) {
					t.Errorf("%s columns = %v, want %s", mode, got, tc.want)
				}
			}
		})
	}
}

func TestSensorStreamKeepsUnitsOnReadingLine(t *testing.T) {
	renderer := new(capturedStream)
	sensors := []*types.Result[types.Sensor]{
		{Ok: &types.Sensor{SDRRecordType: types.SDRRecordTypeFullSensor, Name: "CPU1 Temperature", SensorType: 1, EventReadingType: 1, SensorUnit: types.SensorUnit{BaseUnit: types.SensorUnitType_DegreesC}}},
		{Ok: &types.Sensor{SDRRecordType: types.SDRRecordTypeFullSensor, Name: "+VCORE1", SensorType: 2, EventReadingType: 1, SensorUnit: types.SensorUnit{BaseUnit: types.SensorUnitType_Volts}}},
	}
	if err := types.FormatSensorsStreamWith(renderer, false, slices.Values(sensors)); err != nil {
		t.Fatal(err)
	}
	for _, unit := range []string{"degrees C", "Volts"} {
		if !strings.Contains(renderer.out.String(), unit) {
			t.Errorf("%s split in table:\n%s", unit, renderer.out.String())
		}
	}
}

func TestSDRStreamKeepsUnitOnReadingLine(t *testing.T) {
	renderer := new(capturedStream)
	sdr := &types.SDR{RecordHeader: &types.SDRHeader{RecordType: types.SDRRecordTypeFullSensor}, Full: &types.SDRFull{SensorUnit: types.SensorUnit{BaseUnit: types.SensorUnitType_MilliSecond}}}
	if err := types.FormatSDRsStreamWith(renderer, slices.Values([]*types.Result[types.SDR]{{Ok: sdr}})); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderer.out.String(), "millisecond") {
		t.Errorf("unit split in table:\n%s", renderer.out.String())
	}
}

func TestRendererExamples(t *testing.T) {
	r := ascii.Renderer{}
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
			want: "+------------+---------+-----------+\n" +
				"| SensorName | Reading | Unit      |\n" +
				"+------------+---------+-----------+\n" +
				"| CPU Temp   | 45      | degrees C |\n" +
				"| Fan1       | 12000   | RPM       |\n" +
				"+------------+---------+-----------+\n",
		},
		{name: "missing_cells", headers: []string{"A", "B", "C"},
			rows: []map[string]string{
				{"A": "1"},
				{"B": "two\nlines", "C": "  tab\there  "},
			},
			want: "+---+-----------+----------+\n" +
				"| A | B         | C        |\n" +
				"+---+-----------+----------+\n" +
				"| 1 |           |          |\n" +
				"|   | two lines | tab here |\n" +
				"+---+-----------+----------+\n",
		},
		{name: "no_rows", headers: []string{"A", "B"}, want: "+---+---+\n" +
			"| A | B |\n" +
			"+---+---+\n" +
			"+---+---+\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := r.Render(render.Columns(tc.headers...), tc.rows); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("stream", func(t *testing.T) {
		const firstRows = "+----+--------------------+--------+\n" +
			"| ID | SensorName         | Status |\n" +
			"+----+--------------------+--------+\n" +
			"| 1  | CPU Temp           | ok     |\n"
		const want = firstRows + "| 33 | Fan1               | cr     |\n" +
			"| 3  |                    |        |\n" +
			"+----+--------------------+--------+\n"
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

//go:build goipmi_notablewriter

package types

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

func Test_RenderTable_NoTablewriter(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		rows     []map[string]string
		expected string
	}{
		{
			name:    "basic",
			headers: []string{"SensorName", "Reading", "Unit"},
			rows: []map[string]string{
				{"SensorName": "CPU Temp", "Reading": "45", "Unit": "degrees C"},
				{"SensorName": "Fan1", "Reading": "12000", "Unit": "RPM"},
			},
			expected: "" +
				"SensorName  Reading  Unit\n" +
				"CPU Temp    45       degrees C\n" +
				"Fan1        12000    RPM\n",
		},
		{
			name:    "missing cells and special characters",
			headers: []string{"A", "B", "C"},
			rows: []map[string]string{
				{"A": "1"},
				{"B": "two\nlines", "C": "  tab\there  "},
			},
			expected: "" +
				"A  B          C\n" +
				"1             \n" +
				"   two lines  tab here\n",
		},
		{
			name:     "no rows",
			headers:  []string{"A", "B"},
			expected: "A  B\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RenderTable(tt.headers, tt.rows); got != tt.expected {
				t.Errorf("RenderTable() =\n%q\nwant\n%q", got, tt.expected)
			}
		})
	}
}

func Test_RenderTableStream_NoTablewriter(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	defer r.Close()

	stdout := os.Stdout
	os.Stdout = w

	defer func() { os.Stdout = stdout }()

	// the column widths are fixed by the headers (2, 18, 6)
	headers := []string{"ID", "--- SensorName ---", "Status"}
	line := func(id, name, status string) string {
		return fmt.Sprintf("%-2s  %-18s  %s\n", id, name, status)
	}

	firstRows := line("ID", "SensorName", "Status") + line("1", "CPU Temp", "ok")

	rowSeq := func(yield func(map[string]string) bool) {
		if !yield(map[string]string{"ID": "1", "SensorName": "CPU Temp", "Status": "ok"}) {
			return
		}

		// the header and the first row should be already written before the next row is produced
		if err := r.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Logf("pipe deadlines are not supported: %s", err)
		}

		buf := make([]byte, len(firstRows))

		if _, err := io.ReadFull(r, buf); err != nil {
			t.Errorf("rows are not streamed: %s", err)
		} else if string(buf) != firstRows {
			t.Errorf("RenderTableStream() first rows =\n%q\nwant\n%q", string(buf), firstRows)
		}

		if err := r.SetReadDeadline(time.Time{}); err != nil {
			t.Logf("pipe deadlines are not supported: %s", err)
		}

		// cells which don't fit are not truncated
		yield(map[string]string{"ID": "333", "SensorName": "Fan1", "Status": "cr"})
	}

	err = RenderTableStream(headers, rowSeq)

	os.Stdout = stdout

	w.Close()

	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected := "333  Fan1                cr\n"

	if string(out) != expected {
		t.Errorf("RenderTableStream() last row =\n%q\nwant\n%q", string(out), expected)
	}
}

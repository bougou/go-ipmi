//go:build goipmi_notablewriter

package types

import (
	"bytes"
	"io"
	"iter"
	"os"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

// This file provides a minimal table rendering based on text/tabwriter,
// used when building with the goipmi_notablewriter build tag.
//
// The build tag allows to exclude github.com/olekukonko/tablewriter and its dependencies
// from programs which use go-ipmi as a library and don't need pretty tables.

// RenderTable formats a table from a slice of rows.
// The Headers is a slice of strings, the order of the headers is the order of the columns in the table.
// Each row is represented as a map, the keys of the map are the headers of the table.
func RenderTable(headers []string, rows []map[string]string) string {
	var buf bytes.Buffer

	tw := newTabWriter(&buf)

	writeTableRow(tw, headers)

	for _, row := range rows {
		writeTableRow(tw, rowCells(headers, row, false))
	}

	tw.Flush()

	return buf.String()
}

// RenderTableStream renders the rows from the sequence as a table to stdout, as they are received.
//
// As the rows are not known in advance, the column widths are fixed by the headers:
// headers can be decorated with "-" (e.g. "--- SensorName ---") to make a column wider.
// Cells which don't fit into the column are not truncated, they shift the rest of the row.
func RenderTableStream(headers []string, rowSeq iter.Seq[map[string]string]) error {
	canonicalHeaders := make([]string, len(headers))
	widths := make([]int, len(headers))

	for i, header := range headers {
		canonicalHeaders[i] = canonicalHeader(header)
		widths[i] = utf8.RuneCountInString(header)
	}

	if err := writeStreamRow(os.Stdout, canonicalHeaders, widths); err != nil {
		return err
	}

	for row := range rowSeq {
		if err := writeStreamRow(os.Stdout, rowCells(headers, row, true), widths); err != nil {
			return err
		}
	}

	return nil
}

// writeStreamRow writes a single row with fixed column widths, so that it can be written immediately.
func writeStreamRow(w io.Writer, cells []string, widths []int) error {
	var line strings.Builder

	for i, cell := range cells {
		if i > 0 {
			line.WriteString("  ")
		}

		cell = cellReplacer.Replace(strings.TrimSpace(cell))
		line.WriteString(cell)

		if pad := widths[i] - utf8.RuneCountInString(cell); pad > 0 {
			line.WriteString(strings.Repeat(" ", pad))
		}
	}

	// avoid trailing whitespace, e.g. for empty cells at the end of the row
	_, err := io.WriteString(w, strings.TrimRight(line.String(), " ")+"\n")

	return err
}

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

// canonicalHeader strips the decorations used to widen the columns in stream output.
func canonicalHeader(header string) string {
	header = strings.TrimSpace(header)
	header = strings.Trim(header, "-")

	return strings.TrimSpace(header)
}

func rowCells(headers []string, row map[string]string, canonical bool) []string {
	cells := make([]string, len(headers))

	for i, header := range headers {
		if canonical {
			header = canonicalHeader(header)
		}

		cells[i] = row[header]
	}

	return cells
}

// cellReplacer replaces the characters which break the column alignment.
var cellReplacer = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")

func writeTableRow(w io.Writer, cells []string) {
	for i, cell := range cells {
		if i > 0 {
			io.WriteString(w, "\t")
		}

		io.WriteString(w, cellReplacer.Replace(strings.TrimSpace(cell)))
	}

	io.WriteString(w, "\n")
}

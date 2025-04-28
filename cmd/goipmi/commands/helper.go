package commands

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const (
	DefaultErrorExitCode = 1
)

// ErrExit may be passed to CheckErr to instruct it to output nothing but exit with status code 1.
var ErrExit = fmt.Errorf("exit")

func CheckErr(err error) {
	checkErr(err, fatalErrHandler)
}

// checkErr formats a given error as a string and calls the passed handleErr
// func with that string and an exit code.
func checkErr(err error, handleErr func(string, int)) {

	if err == nil {
		return
	}

	switch {
	case err == ErrExit:
		handleErr("", DefaultErrorExitCode)
	default:
		switch err := err.(type) {
		// other err type, can return different code

		default:
			msg := err.Error()
			if !strings.HasPrefix(msg, "error: ") {
				msg = fmt.Sprintf("error: %s", msg)
			}
			handleErr(msg, DefaultErrorExitCode)
		}
	}
}

var fatalErrHandler = fatal

// fatal prints the message (if provided) and then exits.
func fatal(msg string, code int) {
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(code)
}

// formatTable formats a table from a slice of rows.
// The row is represented as a map, the keys of the map are the headers of the table.
func formatTable(headers []string, rows []map[string]string) string {
	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader(headers)
	table.SetFooter(headers)

	row := make([]string, len(headers))
	for _, _row := range rows {
		for i, header := range headers {
			row[i] = _row[header]
		}
		table.Append(row)
	}

	table.Render()
	return buf.String()
}

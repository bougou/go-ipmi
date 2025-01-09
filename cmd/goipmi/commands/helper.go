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

// ErrExit may be passed to CheckError to instruct it to output nothing but exit with status code 1.
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

func formatTable(headers []string, rows [][]string) string {
	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader(headers)
	table.SetFooter(headers)

	for _, row := range rows {
		table.Append(row)
	}

	table.Render()
	return buf.String()
}

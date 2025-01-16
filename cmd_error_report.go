package ipmi

import (
	"context"
	"fmt"
)

type ErrorReportRequest struct {
}

type ErrorReportResponse struct {
	OriginalCommand uint8
	ErrorCode       uint8
}

func (req *ErrorReportRequest) Command() Command {
	return CommandErrorReport
}

func (req *ErrorReportRequest) Pack() []byte {
	return []byte{}
}

func (res *ErrorReportResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.OriginalCommand = msg[0]
	res.ErrorCode = msg[1]
	return nil
}

func (res *ErrorReportResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *ErrorReportResponse) Format() string {
	return fmt.Sprintf(`
Original Command : %#02x
Error Code       : %#02x
`,
		res.OriginalCommand, res.ErrorCode)
}

func (c *Client) ErrorReport(ctx context.Context) (response *ErrorReportResponse, err error) {
	request := &ErrorReportRequest{}
	response = &ErrorReportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

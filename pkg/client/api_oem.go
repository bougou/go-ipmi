package client

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/cmd/oem"
)

func (c *Client) GetSupermicroBiosVersion(ctx context.Context) (response *oem.CommandGetSupermicroBiosVersionResponse, err error) {
	request := &oem.CommandGetSupermicroBiosVersionRequest{}
	response = &oem.CommandGetSupermicroBiosVersionResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

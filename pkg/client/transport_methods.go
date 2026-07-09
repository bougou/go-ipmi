package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// BuildIPMIRequest creates an IPMIRequest for a Command Request, filling in
// checksum fields.
func (c *Client) BuildIPMIRequest(ctx context.Context, reqCmd types.Request) (*types.IPMIRequest, error) {
	c.lock()
	defer c.unlock()

	ipmiReq := &types.IPMIRequest{
		ResponderAddr: c.responderAddr,

		NetFn:        reqCmd.Command().NetFn,
		ResponderLUN: c.responderLUN,

		RequesterAddr: c.requesterAddr,

		RequesterSequence: c.session.ipmiSeq,
		RequesterLUN:      c.requesterLUN,

		Command:     reqCmd.Command().ID,
		CommandData: reqCmd.Pack(),
	}

	commandContext := GetCommandContext(ctx)
	if commandContext != nil {
		c.Debug("Got CommandContext:", commandContext)

		if commandContext.responderAddr != nil {
			ipmiReq.ResponderAddr = *commandContext.responderAddr
		}
		if commandContext.responderLUN != nil {
			ipmiReq.ResponderLUN = *commandContext.responderLUN
		}
		if commandContext.requesterAddr != nil {
			ipmiReq.RequesterAddr = *commandContext.requesterAddr
		}
		if commandContext.requesterLUN != nil {
			ipmiReq.RequesterLUN = *commandContext.requesterLUN
		}
	}

	c.session.ipmiSeq += 1
	if c.session.ipmiSeq > types.IPMIRequesterSequenceMax {
		c.session.ipmiSeq = 1
	}

	ipmiReq.ComputeChecksum()

	return ipmiReq, nil
}

// BuildRmcpRequest builds an RMCP packet for the given command request.
func (c *Client) BuildRmcpRequest(ctx context.Context, reqCmd types.Request) (*types.Rmcp, error) {
	payloadType, rawPayload, err := c.buildRawPayload(ctx, reqCmd)
	if err != nil {
		return nil, fmt.Errorf("buildRawPayload failed, err: %w", err)
	}
	c.DebugBytes("rawPayload", rawPayload, 16)

	// ASF ping
	if _, ok := reqCmd.(*RmcpPingRequest); ok {
		return &types.Rmcp{
			RmcpHeader: types.NewRmcpHeaderASF(),
			ASF: &types.ASF{
				IANA:        4542,
				MessageType: uint8(types.MessageTypePing),
				MessageTag:  0,
				DataLength:  0,
				Data:        rawPayload,
			},
		}, nil
	}

	// IPMI 2.0
	if c.v20 {
		session20, err := c.genSession20(payloadType, rawPayload)
		if err != nil {
			return nil, fmt.Errorf("genSession20 failed, err: %w", err)
		}
		return &types.Rmcp{RmcpHeader: types.NewRmcpHeader(), Session20: session20}, nil
	}

	// IPMI 1.5
	session15, err := c.genSession15(rawPayload)
	if err != nil {
		return nil, fmt.Errorf("genSession15 failed, err: %w", err)
	}
	return &types.Rmcp{RmcpHeader: types.NewRmcpHeader(), Session15: session15}, nil
}

// ParseRmcpResponse parses a raw RMCP response message into the given Response.
func (c *Client) ParseRmcpResponse(ctx context.Context, msg []byte, response types.Response) error {
	rmcp := &types.Rmcp{}
	if err := rmcp.Unpack(msg); err != nil {
		return fmt.Errorf("unpack rmcp failed, err: %w", err)
	}
	c.Debug("<<<<<< RMCP Response", rmcp)

	if rmcp.ASF != nil {
		if int(rmcp.ASF.DataLength) != len(rmcp.ASF.Data) {
			return fmt.Errorf("asf Data Length not equal")
		}
		if err := response.Unpack(rmcp.ASF.Data); err != nil {
			return fmt.Errorf("unpack asf response failed, err: %w", err)
		}
		return nil
	}

	if rmcp.Session15 != nil {
		ipmiPayload := rmcp.Session15.Payload

		ipmiRes := types.IPMIResponse{}
		if err := ipmiRes.Unpack(ipmiPayload); err != nil {
			return fmt.Errorf("unpack ipmiRes failed, err: %w", err)
		}
		c.Debug("<<<< IPMI Response", ipmiRes)

		ccode := ipmiRes.CompletionCode
		if ccode != 0x00 {
			return types.NewResponseError(
				types.CompletionCode(ccode),
				fmt.Sprintf("ipmiRes CompletionCode (%#02x) is not normal: %s", ccode, types.StrCC(response, ccode)),
			)
		}
		if err := response.Unpack(ipmiRes.Data); err != nil {
			return types.NewResponseError(0x00, fmt.Sprintf("unpack response failed, err: %s", err))
		}
		return nil
	}

	if rmcp.Session20 != nil {
		sessionHdr := rmcp.Session20.SessionHeader20

		switch sessionHdr.PayloadType {
		case
			types.PayloadTypeRmcpOpenSessionResponse,
			types.PayloadTypeRAKPMessage2,
			types.PayloadTypeRAKPMessage4:
			if err := response.Unpack(rmcp.Session20.SessionPayload); err != nil {
				return fmt.Errorf("unpack session setup response failed, err: %w", err)
			}
			return nil

		case types.PayloadTypeSOL:
			payload := rmcp.Session20.SessionPayload
			if sessionHdr.PayloadEncrypted {
				c.DebugBytes("decrypting SOL payload", payload, 16)
				d, err := c.decryptPayload(payload)
				if err != nil {
					return fmt.Errorf("decrypt SOL session payload failed, err: %w", err)
				}
				payload = d
				c.DebugBytes("decrypted SOL payload", payload, 16)
			}
			if err := response.Unpack(payload); err != nil {
				return fmt.Errorf("unpack SOL payload response failed, err: %w", err)
			}
			return nil

		case types.PayloadTypeIPMI:
			ipmiPayload := rmcp.Session20.SessionPayload
			if sessionHdr.PayloadEncrypted {
				c.DebugBytes("decrypting", ipmiPayload, 16)
				d, err := c.decryptPayload(rmcp.Session20.SessionPayload)
				if err != nil {
					return fmt.Errorf("decrypt session payload failed, err: %w", err)
				}
				ipmiPayload = d
				c.DebugBytes("decrypted", ipmiPayload, 16)
			}

			ipmiRes := types.IPMIResponse{}
			if err := ipmiRes.Unpack(ipmiPayload); err != nil {
				return fmt.Errorf("unpack ipmiRes failed, err: %w", err)
			}
			c.Debug("<<<< IPMI Response", ipmiRes)

			ccode := ipmiRes.CompletionCode
			if ccode != 0x00 {
				return types.NewResponseError(
					types.CompletionCode(ccode),
					fmt.Sprintf("ipmiRes CompletionCode (%#02x) is not normal: %s", ccode, types.StrCC(response, ccode)),
				)
			}
			if err := response.Unpack(ipmiRes.Data); err != nil {
				return types.NewResponseError(0x00, fmt.Sprintf("unpack response failed, err: %s", err))
			}
			return nil
		}
	}

	return fmt.Errorf("not an IPMI response")
}

func (c *Client) parseIPMIResponseFromRmcp(rmcp *types.Rmcp) (ipmiRes *types.IPMIResponse, err error) {
	if rmcp.ASF != nil {
		return nil, fmt.Errorf("not an IPMI response (ASF)")
	}

	if rmcp.Session15 != nil {
		ipmiRes := &types.IPMIResponse{}
		if err := ipmiRes.Unpack(rmcp.Session15.Payload); err != nil {
			return nil, fmt.Errorf("unpack ipmi(15) payload failed, err: %w", err)
		}
		return ipmiRes, nil
	}

	if rmcp.Session20 != nil {
		sessionHdr := rmcp.Session20.SessionHeader20
		if sessionHdr.PayloadType != types.PayloadTypeIPMI {
			return nil, fmt.Errorf("not an IPMI response (%s)", sessionHdr.PayloadType)
		}

		ipmiPayload := rmcp.Session20.SessionPayload
		if sessionHdr.PayloadEncrypted {
			c.DebugBytes("decrypting", ipmiPayload, 16)
			d, err := c.decryptPayload(rmcp.Session20.SessionPayload)
			if err != nil {
				return nil, fmt.Errorf("decrypt ipmi(20) session payload failed, err: %w", err)
			}
			ipmiPayload = d
			c.DebugBytes("decrypted", ipmiPayload, 16)
		}

		ipmiRes := &types.IPMIResponse{}
		if err := ipmiRes.Unpack(ipmiPayload); err != nil {
			return nil, fmt.Errorf("unpack ipmi(20) payload failed, err: %w", err)
		}
		return ipmiRes, nil
	}

	return nil, fmt.Errorf("not an invalid Rmcp payload")
}

// findBestCipherSuites queries the BMC for available cipher suites and returns
// them sorted by preference.
func (c *Client) findBestCipherSuites(ctx context.Context) []types.CipherSuiteID {
	cipherSuiteRecords, err := c.GetAllChannelCipherSuites(ctx, types.ChannelNumberSelf)
	if err != nil {
		return types.PreferredCiphers
	}
	cipherSuiteIDs := make([]types.CipherSuiteID, len(cipherSuiteRecords))
	for i, cipherSuiteRecord := range cipherSuiteRecords {
		cipherSuiteIDs[i] = cipherSuiteRecord.CipherSuitID
	}
	return types.SortCipherSuites(cipherSuiteIDs)
}

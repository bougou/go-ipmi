package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

// buildRawPayload returns the PayloadType and the raw payload bytes for Command Request.
// Most command requests are of IPMI PayloadType, but some requests like RAKP messages are not.
func (c *Client) buildRawPayload(ctx context.Context, reqCmd types.Request) (types.PayloadType, []byte, error) {
	var payloadType types.PayloadType
	if _, ok := reqCmd.(*rmcpplus.OpenSessionRequest); ok {
		payloadType = types.PayloadTypeRmcpOpenSessionRequest
	} else if _, ok := reqCmd.(*rmcpplus.RAKPMessage1); ok {
		payloadType = types.PayloadTypeRAKPMessage1
	} else if _, ok := reqCmd.(*rmcpplus.RAKPMessage3); ok {
		payloadType = types.PayloadTypeRAKPMessage3
	} else if _, ok := reqCmd.(*types.SOLPayloadRequest); ok {
		payloadType = types.PayloadTypeSOL
	} else {
		payloadType = types.PayloadTypeIPMI
	}

	var rawPayload []byte
	switch payloadType {
	case types.PayloadTypeRmcpOpenSessionRequest, types.PayloadTypeRAKPMessage1, types.PayloadTypeRAKPMessage3:
		// Session Setup Payload Types

		rawPayload = reqCmd.Pack()

	case types.PayloadTypeSOL:
		rawPayload = reqCmd.Pack()

	case types.PayloadTypeIPMI:
		// Standard Payload Types
		ipmiReq, err := c.BuildIPMIRequest(ctx, reqCmd)
		if err != nil {
			return 0, nil, fmt.Errorf("BuildIPMIRequest failed, err: %w", err)
		}

		c.Debug(">>>> IPMI Request", ipmiReq)
		rawPayload = ipmiReq.Pack()
	}

	return payloadType, rawPayload, nil
}

// isIPMIPayloadLANRequest reports whether buildRawPayload uses PayloadTypeIPMI for this request.
// Session-setup and SOL payloads are excluded; those are not paired by IPMB rqSeq + command like
// standard IPMI commands.
func isIPMIPayloadLANRequest(req types.Request) bool {
	switch req.(type) {
	case *rmcpplus.OpenSessionRequest, *rmcpplus.RAKPMessage1, *rmcpplus.RAKPMessage3, *types.SOLPayloadRequest, *RmcpPingRequest:
		return false
	default:
		return true
	}
}

// tryMatchIPMILANResponse returns true if recv is an RMCP+ (or v1.5) packet whose IPMI payload
// matches the pending request identified by rqSeq and command.
func (c *Client) tryMatchIPMILANResponse(recv []byte, wantSeq, wantCmd uint8) (bool, error) {
	rmcp := &types.Rmcp{}
	if err := rmcp.Unpack(recv); err != nil {
		c.DebugfYellow("drop recv: rmcp unpack failed: %s\n", err)
		c.DebugBytes("dropped recv (rmcp unpack failed)", recv, 16)
		return false, nil
	}
	ipmiRes, err := c.parseIPMIResponseFromRmcp(rmcp)
	if err != nil {
		c.DebugfYellow("drop recv: parseIPMIResponseFromRmcp failed: %s\n", err)
		c.DebugBytes("dropped recv (ipmi unpack failed)", recv, 16)
		return false, nil
	}
	if ipmiRes.RequesterSequence != wantSeq || ipmiRes.Command != wantCmd {
		c.DebugfYellow("drop recv: mismatch (got rqSeq %#02x cmd %#02x, want rqSeq %#02x cmd %#02x)\n",
			ipmiRes.RequesterSequence, ipmiRes.Command, wantSeq, wantCmd)
		c.DebugBytes("dropped recv (rqSeq/cmd mismatch)", recv, 16)
		return false, nil
	}
	return true, nil
}

// tryMatchSOLResponse returns true if recv is an RMCP+ SOL payload packet
// acknowledging the pending request. A data request is acknowledged by
// echoing its sequence number (spec v2.0 §15.9/§15.11), so the response to
// the request with sequence number wantAck always carries
// AckedSequenceNumber == wantAck. Asynchronous BMC output may acknowledge an
// earlier request, so it is preserved for the active stream to process after
// the pending request completes.
//
// ACK-only requests (sequence 0h) carry no number for the BMC to echo: it
// answers with the last accepted data sequence instead (which is non-zero
// once any keystroke has flowed), so no acked-based filter exists and the
// first SOL packet on the wire is taken as their response.
func (c *Client) tryMatchSOLResponse(recv []byte, wantAck uint8) (bool, error) {
	rmcp := &types.Rmcp{}
	if err := rmcp.Unpack(recv); err != nil {
		c.DebugfYellow("drop recv: rmcp unpack failed: %s\n", err)
		c.DebugBytes("dropped recv (rmcp unpack failed)", recv, 16)
		return false, nil
	}
	if rmcp.Session20 == nil {
		return false, nil
	}
	hdr := rmcp.Session20.SessionHeader20
	if hdr.PayloadType != types.PayloadTypeSOL {
		return false, nil
	}
	payload := rmcp.Session20.SessionPayload
	if hdr.PayloadEncrypted {
		d, err := c.decryptPayload(payload)
		if err != nil {
			c.DebugfYellow("drop recv: decrypt SOL payload failed: %s\n", err)
			return false, nil
		}
		payload = d
	}
	var sol types.SOLPayloadPacket
	if err := sol.Unpack(payload); err != nil {
		c.DebugfYellow("drop recv: SOL payload unpack failed: %s\n", err)
		return false, nil
	}
	if wantAck != 0 && sol.AckedSequenceNumber != wantAck {
		if sol.SequenceNumber != 0 {
			c.deliverSOLOutput(&types.SOLPayloadResponse{SOLPayloadPacket: sol})
		}
		c.DebugfYellow("drop recv: SOL ack mismatch (got ack %d, want ack %d)\n", sol.AckedSequenceNumber, wantAck)
		return false, nil
	}
	return true, nil
}

func (c *Client) exchangeLAN(ctx context.Context, request types.Request, response types.Response) error {
	c.Debug(">> Command Request", request)

	var wantSeq, wantCmd uint8
	applyIPMIMatch := isIPMIPayloadLANRequest(request)
	if applyIPMIMatch {
		c.lock()
		wantSeq = c.session.ipmiSeq
		wantCmd = request.Command().ID
		c.unlock()
	}
	// SOL responses are matched by the acked sequence number echoing the
	// request's sequence number (§15.9/§15.11), not by rqSeq/cmd. ACK-only
	// requests (sequence 0h) get no echo and match the first SOL packet
	// (see tryMatchSOLResponse).
	var wantSOLAck uint8
	solReq, isSOL := request.(*types.SOLPayloadRequest)
	if isSOL {
		wantSOLAck = solReq.SequenceNumber
	}

	rmcp, err := c.BuildRmcpRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("build RMCP+ request msg failed, err: %w", err)
	}
	c.Debug(">>>>>> RMCP Request", rmcp)
	sent := rmcp.Pack()
	c.DebugBytes("sent", sent, 16)

	var recv []byte
	attempts := c.retryCount + 1 // initial try plus retries
	c.Debugf("exchange LAN (attempts: %d)\n", attempts)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		c.Debugf("attempt %d/%d, ", attempt, attempts)

		switch {
		case applyIPMIMatch:
			matcher := func(p []byte) (bool, error) {
				return c.tryMatchIPMILANResponse(p, wantSeq, wantCmd)
			}
			recv, err = c.udpClient.ExchangeUntilMatch(ctx, bytes.NewReader(sent), matcher)
		case isSOL:
			// The socket also carries the server's unsolicited SOL
			// retransmissions; a bare Exchange would consume one of those as
			// the response to this request.
			matcher := func(p []byte) (bool, error) {
				return c.tryMatchSOLResponse(p, wantSOLAck)
			}
			recv, err = c.udpClient.ExchangeUntilMatch(ctx, bytes.NewReader(sent), matcher)
		default:
			recv, err = c.udpClient.Exchange(ctx, bytes.NewReader(sent))
		}
		lastErr = err
		if err == nil {
			c.DebugfGreen("udp exchange success\n")
			// Success, break out of retry loop
			break
		}

		// Retry when no datagram matched the pending IPMI request (stray SOL/async) or on UDP read timeout.
		if errors.Is(err, errNoDatagramMatched) {
			c.DebugfRed("udp exchange: no matching IPMI response (want seq %#02x cmd %#02x), retry\n", wantSeq, wantCmd)
			continue
		}
		var netErr *net.OpError
		if errors.As(err, &netErr) && netErr.Timeout() {
			c.DebugfRed("udp exchange failed, timeout error: %v\n", err)
			continue
		}

		// Non-timeout error or final attempt, don't retry
		c.DebugfRed("udp exchange failed, error: %s\n", err)
		break
	}

	if lastErr != nil {
		return wrapExchangeLANError(attempts, applyIPMIMatch, wantSeq, wantCmd, lastErr)
	}

	c.DebugBytes("recv", recv, 16)

	if err := c.ParseRmcpResponse(ctx, recv, request.Command(), response); err != nil {
		return err
	}

	c.Debug("<< Command Response", response)
	return nil
}

// wrapExchangeLANError maps the last UDP-layer error to a user-facing IPMI LAN message while
// preserving the underlying error chain for errors.Is / errors.As.
func wrapExchangeLANError(attempts int, applyIPMIMatch bool, wantSeq, wantCmd uint8, err error) error {
	if err == nil {
		return nil
	}
	var detail string
	switch {
	case errors.Is(err, context.Canceled):
		detail = "operation canceled"
	case errors.Is(err, context.DeadlineExceeded):
		detail = "deadline exceeded while exchanging with BMC"
	case errors.Is(err, errNoDatagramMatched):
		if applyIPMIMatch {
			detail = fmt.Sprintf("no matching IPMI response (expected rqSeq %#02x, command %#02x) before UDP read deadline", wantSeq, wantCmd)
		} else {
			detail = "no UDP datagram matched before read deadline"
		}
	default:
		var netErr *net.OpError
		if errors.As(err, &netErr) && netErr.Timeout() {
			detail = "UDP read timed out waiting for BMC response"
		} else {
			detail = "UDP exchange error"
		}
	}
	return fmt.Errorf("IPMI LAN: %s after %d attempt(s): %w", detail, attempts, err)
}

// 13.14
// IPMI v1.5 LAN Session Activation
// 1. RmcpPresencePing - RmcpPresencePong
// 2. Get Channel Authentication Capabilities
// 3. Get Session Challenge
// 4. Activate Session
func (c *Client) Connect15(ctx context.Context) error {
	var (
		err           error
		channelNumber uint8 = types.ChannelNumberSelf
	)

	if c.maxPrivilegeLevel == types.PrivilegeLevelUnspecified {
		c.maxPrivilegeLevel = types.PrivilegeLevelAdministrator
	}

	_, err = c.GetChannelAuthenticationCapabilities(ctx, channelNumber, c.maxPrivilegeLevel)
	if err != nil {
		return fmt.Errorf("GetChannelAuthenticationCapabilities failed, err: %w", err)
	}

	_, err = c.GetSessionChallenge(ctx)
	if err != nil {
		return fmt.Errorf("GetSessionChallenge failed, err: %w", err)
	}

	c.session.v15.preSession = true

	_, err = c.ActivateSession(ctx)
	if err != nil {
		return fmt.Errorf("ActivateSession failed, err: %w", err)
	}

	_, err = c.SetSessionPrivilegeLevel(ctx, c.maxPrivilegeLevel)
	if err != nil {
		return fmt.Errorf("SetSessionPrivilegeLevel to (%s) failed, err: %w", c.maxPrivilegeLevel, err)
	}

	// The Connect context bounds setup. Client.Close owns the established session lifetime.
	go c.keepSessionAlive(context.WithoutCancel(ctx), DefaultKeepAliveIntervalSec)

	return nil

}

// see 13.15 IPMI v2.0/RMCP+ Session Activation
func (c *Client) Connect20(ctx context.Context) error {
	var (
		err           error
		channelNumber uint8 = types.ChannelNumberSelf
	)

	if c.maxPrivilegeLevel == types.PrivilegeLevelUnspecified {
		c.maxPrivilegeLevel = types.PrivilegeLevelAdministrator
	}

	// Per IPMI 2.0 spec §13.15, the initial Get Channel Authentication
	// Capabilities command is sent in IPMI 1.5 packet format before the
	// RMCP+ session is established.  Some BMC implementations (notably
	// ipmi_sim) do not respond to this command when framed inside an
	// RMCP+ session-zero (AuthType 0x06) header.
	c.v20 = false
	_, err = c.GetChannelAuthenticationCapabilities(ctx, channelNumber, c.maxPrivilegeLevel)
	c.v20 = true
	if err != nil {
		return fmt.Errorf("cmd: Get Channel Authentication Capabilities failed, err: %w", err)
	}

	var tryCiphers []types.CipherSuiteID

	c.session.v20.customSuiteIDs = slices.DeleteFunc(c.session.v20.customSuiteIDs, func(id types.CipherSuiteID) bool {
		return id == types.CipherSuiteIDReserved
	})
	if len(c.session.v20.customSuiteIDs) > 0 {
		tryCiphers = c.session.v20.customSuiteIDs
	} else if c.session.v20.cipherSuiteID != types.CipherSuiteIDReserved {
		// client explicitly specified a cipher suite to use
		tryCiphers = []types.CipherSuiteID{c.session.v20.cipherSuiteID}
	} else {
		tryCiphers = c.findBestCipherSuites(ctx)
	}

	c.DebugfGreen("\n\ntry ciphers (%v)\n", tryCiphers)

	var success bool
	errs := []error{}

	// try different cipher suites for opensession/rakp1/rakp3
	for _, cipherSuiteID := range tryCiphers {
		c.DebugfGreen("\n\ntry cipher suite id (%v)\n\n\n", cipherSuiteID)

		c.session.v20.cipherSuiteID = cipherSuiteID

		_, err = c.OpenSession(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("cmd: RMCP+ Open Session failed with cipher suite id (%v), err: %w", cipherSuiteID, err))
			continue
		}

		_, err = c.RAKPMessage1(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("cmd: rakp1 failed with cipher suite id (%v), err: %w", cipherSuiteID, err))
			continue
		}

		_, err = c.RAKPMessage3(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("cmd: rakp3 failed with cipher suite id (%v), err: %w", cipherSuiteID, err))
			continue
		}

		c.DebugfGreen("\n\nconnect20 success with cipher suite id (%v)\n\n\n", cipherSuiteID)
		success = true
		break
	}

	if !success {
		return fmt.Errorf("connect20 failed after try all cipher suite ids (%v), errs: \n%v", tryCiphers, errors.Join(errs...))
	}

	_, err = c.SetSessionPrivilegeLevel(ctx, c.maxPrivilegeLevel)
	if err != nil {
		return fmt.Errorf("SetSessionPrivilegeLevel to (%s) failed, err: %w", c.maxPrivilegeLevel, err)
	}

	// The Connect context bounds setup. Client.Close owns the established session lifetime.
	go c.keepSessionAlive(context.WithoutCancel(ctx), DefaultKeepAliveIntervalSec)

	return nil
}

// ConnectAuto detects the IPMI version supported by BMC by using
// GetChannelAuthenticationCapabilities command, then decide to use v1.5 or v2.0
// for subsequent requests.
func (c *Client) ConnectAuto(ctx context.Context) error {
	var (
		err error

		channelNumber uint8 = types.ChannelNumberSelf

		privilegeLevel types.PrivilegeLevel = types.PrivilegeLevelAdministrator
	)

	// force use IPMI v1.5 first
	c.v20 = false
	cap, err := c.GetChannelAuthenticationCapabilities(ctx, channelNumber, privilegeLevel)
	if err != nil {
		return fmt.Errorf("cmd: Get Channel Authentication Capabilities failed, err: %w", err)
	}
	if cap.SupportIPMIv20 {
		c.v20 = true
		return c.Connect20(ctx)
	}
	if cap.SupportIPMIv15 {
		return c.Connect15(ctx)
	}
	return fmt.Errorf("client does not support IPMI v1.5 and IPMI v.20")
}

// closeLAN closes session used in LAN communication.
func (c *Client) closeLAN(ctx context.Context) error {
	// close the channel to notify the keepAliveSession goroutine to stop
	close(c.closedCh)

	// Closing the network connection must not depend on the BMC replying to
	// close session, it always needs to be done or we will have a resource leak.
	// For example a timed-out or unreachable BMC must not leave the socket open.
	var sessionID uint32
	if c.v20 {
		sessionID = c.session.v20.bmcSessionID
	} else {
		sessionID = c.session.v15.sessionID
	}

	request := &app.CloseSessionRequest{
		SessionID: sessionID,
	}
	_, sessionErr := c.CloseSession(ctx, request)
	if sessionErr != nil {
		sessionErr = fmt.Errorf("CloseSession failed, err: %w", sessionErr)
	}

	connectionErr := c.udpClient.Close()
	if connectionErr != nil {
		connectionErr = fmt.Errorf("close UDP connection failed, err: %w", connectionErr)
	}

	return errors.Join(sessionErr, connectionErr)
}

// 6.12.15 Session Inactivity Timeouts
func (c *Client) keepSessionAlive(ctx context.Context, intervalSec int) {
	var period = time.Duration(intervalSec) * time.Second
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	c.Debugf("keepSessionAlive started\n")
	for {
		select {
		case <-ticker.C:
			if _, err := c.GetCurrentSessionInfo(ctx); err != nil {
				c.DebugfRed("keepSessionAlive failed, GetCurrentSessionInfo failed, err: %w", err)
			}
		case <-c.closedCh:
			c.Debugf("got close signal, keepSessionAlive stopped\n")
			return
		}
	}
}

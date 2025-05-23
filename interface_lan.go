package ipmi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	IPMIVersion15 = 0x15
	IPMIVersion20 = 0x20
)

// session holds data exchanged during Session Activation stage when using lan/lanplus interface.
// see: 13.14 IPMI v1.5 LAN Session Activation, 13.15 IPMI v2.0/RMCP+ Session Activation
type session struct {
	// filled after GetChannelAuthenticationCapabilities
	authType AuthType
	ipmiSeq  uint8
	v20      v20
	v15      v15
}

type v15 struct {
	// indicate whether or not the session is in Pre-Session stage,
	// that is between "GetSessionChallenge" and "ActivateSession"
	preSession bool

	// indicate whether or not the IPMI 1.5 session is activated.
	active bool

	sessionID uint32

	// Sequence number that BMC wants remote console to use for subsequent messages in the session.
	// Remote console use "inSeq" and increment it when sending Request to BMC.
	// "inSeq" is first updated by returned ActivateSession response.
	inSeq uint32

	// "outSeq" is set by Remote Console to indicate the sequence number should picked by BMC.
	// 6.12.12 IPMI v1.5 Outbound Session Sequence Number Tracking and Handling.
	outSeq uint32

	challenge [16]byte
}

type v20 struct {
	// specific to IPMI v2 / RMCP+ sessions
	state    SessionState
	sequence uint32 // session sequence number

	// the cipher suite used during OpenSessionRequest
	cipherSuiteID CipherSuiteID

	// filled by RmcpOpenSessionRequest
	requestedAuthAlg      AuthAlg
	requestedIntegrityAlg IntegrityAlg
	requestedEncryptAlg   CryptAlg

	// filled by RmcpOpenSessionResponse
	// RMCP Open Session is used for exchanging session ids
	authAlg      AuthAlg
	integrityAlg IntegrityAlg
	cryptAlg     CryptAlg

	role             uint8 // whole byte of privilege level in RAKP1, will be used for computing authcode of rakp2, rakp3
	consoleSessionID uint32
	bmcSessionID     uint32

	// values required for RAKP messages

	// filed in rakp1
	consoleRand [16]byte // Random number generated by the console

	// filled after rakp2
	bmcRand         [16]byte // Random number generated by the BMC
	bmcGUID         [16]byte // bmc GUID
	sik             []byte   // SIK, session integrity key
	k1              []byte   // K1 key
	k2              []byte   // K2 key
	rakp2ReturnCode uint8    // will be used in rakp3 message

	// see 13.33
	// Kuid vs Kg
	//  - ipmi user password (the pre-shared key), known as Kuid, which are set using the Set User Password command.
	//  - BMC key, known as Kg, Kg is set using the Set Channel Security Keys command.
	bmcKey []byte

	accumulatedPayloadSize uint32

	// for xRC4 encryption
	rc4EncryptIV [16]byte
	rc4DecryptIV [16]byte
}

// buildRawPayload returns the PayloadType and the raw payload bytes for Command Request.
// Most command requests are of IPMI PayloadType, but some requests like RAKP messages are not.
func (c *Client) buildRawPayload(ctx context.Context, reqCmd Request) (PayloadType, []byte, error) {
	var payloadType PayloadType
	if _, ok := reqCmd.(*OpenSessionRequest); ok {
		payloadType = PayloadTypeRmcpOpenSessionRequest
	} else if _, ok := reqCmd.(*RAKPMessage1); ok {
		payloadType = PayloadTypeRAKPMessage1
	} else if _, ok := reqCmd.(*RAKPMessage3); ok {
		payloadType = PayloadTypeRAKPMessage3
	} else {
		payloadType = PayloadTypeIPMI
	}

	var rawPayload []byte
	switch payloadType {
	case
		PayloadTypeRmcpOpenSessionRequest,
		PayloadTypeRAKPMessage1,
		PayloadTypeRAKPMessage3:
		// Session Setup Payload Types

		rawPayload = reqCmd.Pack()

	case PayloadTypeIPMI:
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

func (c *Client) exchangeLAN(ctx context.Context, request Request, response Response) error {
	c.Debug(">> Command Request", request)

	rmcp, err := c.BuildRmcpRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("build RMCP+ request msg failed, err: %w", err)
	}
	c.Debug(">>>>>> RMCP Request", rmcp)
	sent := rmcp.Pack()
	c.DebugBytes("sent", sent, 16)

	var recv []byte
	attempts := c.retryCount + 1 // initial try plus retries
	c.Debugf("exchangeLAN timeout attempts: %d\n", attempts)

	attemptCount := 0
	for attempt := 1; attempt <= attempts; attempt += 1 {
		attemptCount = attempt
		c.Debugf("Attempt %d/%d\n", attempt, attempts)
		recv, err = c.udpClient.Exchange(ctx, bytes.NewReader(sent))
		if err != nil {
			var netErr *net.OpError
			if errors.As(err, &netErr) {
				c.Debugf("udp exchange error is net error, %s\n", err)

				if netErr.Timeout() {
					c.Debugf("udp exchange error is net timeout error: %v\n", err)

					if attempt < attempts {
						c.Debugf("Attempt %d/%d: timeout error: %v. Retrying...\n", attempt, attempts, err)
						time.Sleep(c.retryInterval)
						continue
					}
				}

				c.Debugf("udp exchange error is net error but not timeout error, %s\n", err)
				break
			}

			c.Debugf("udp exchange error is not net error, %s\n", err)
			break
		}

		// no error
		break
	}

	if err != nil {
		return fmt.Errorf("client udp exchange msg failed, attempts %d times, err: %w", attemptCount, err)
	}
	c.DebugBytes("recv", recv, 16)

	if err := c.ParseRmcpResponse(ctx, recv, response); err != nil {
		return err
	}

	c.Debug("<< Command Response", response)
	return nil

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
		channelNumber uint8 = ChannelNumberSelf
	)

	if c.maxPrivilegeLevel == PrivilegeLevelUnspecified {
		c.maxPrivilegeLevel = PrivilegeLevelAdministrator
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

	go func() {
		c.keepSessionAlive(ctx, DefaultKeepAliveIntervalSec)
	}()

	return nil

}

// see 13.15 IPMI v2.0/RMCP+ Session Activation
func (c *Client) Connect20(ctx context.Context) error {
	var (
		err           error
		channelNumber uint8 = ChannelNumberSelf
	)

	if c.maxPrivilegeLevel == PrivilegeLevelUnspecified {
		c.maxPrivilegeLevel = PrivilegeLevelAdministrator
	}

	_, err = c.GetChannelAuthenticationCapabilities(ctx, channelNumber, c.maxPrivilegeLevel)
	if err != nil {
		return fmt.Errorf("cmd: Get Channel Authentication Capabilities failed, err: %w", err)
	}

	tryCiphers := c.findBestCipherSuites(ctx)

	if c.session.v20.cipherSuiteID != CipherSuiteIDReserved {
		// client explicitly specified a cipher suite to use
		tryCiphers = []CipherSuiteID{c.session.v20.cipherSuiteID}
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

	go func() {
		c.keepSessionAlive(ctx, DefaultKeepAliveIntervalSec)
	}()

	return nil
}

// ConnectAuto detects the IPMI version supported by BMC by using
// GetChannelAuthenticationCapabilities command, then decide to use v1.5 or v2.0
// for subsequent requests.
func (c *Client) ConnectAuto(ctx context.Context) error {
	var (
		err error

		channelNumber uint8 = ChannelNumberSelf

		privilegeLevel PrivilegeLevel = PrivilegeLevelAdministrator
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

	var sessionID uint32
	if c.v20 {
		sessionID = c.session.v20.bmcSessionID
	} else {
		sessionID = c.session.v15.sessionID
	}

	request := &CloseSessionRequest{
		SessionID: sessionID,
	}
	if _, err := c.CloseSession(ctx, request); err != nil {
		return fmt.Errorf("CloseSession failed, err: %w", err)
	}

	if err := c.udpClient.Close(); err != nil {
		return fmt.Errorf("close udp connection failed, err: %w", err)
	}

	return nil
}

// 6.12.15 Session Inactivity Timeouts
func (c *Client) keepSessionAlive(ctx context.Context, intervalSec int) {
	var period = time.Duration(intervalSec) * time.Second
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	c.Debugf("keepSessionAlive started")
	for {
		select {
		case <-ticker.C:
			if _, err := c.GetCurrentSessionInfo(ctx); err != nil {
				c.DebugfRed("keepSessionAlive failed, GetCurrentSessionInfo failed, err: %w", err)
			}
		case <-c.closedCh:
			c.Debugf("got close signal, keepSessionAlive stopped")
			return
		}
	}
}

package ipmi

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

type Interface string

const (
	InterfaceLan     Interface = "lan"
	InterfaceLanplus Interface = "lanplus"

	DefaultExchangeTimeoutSec int = 20
	DefaultBufferSize         int = 1024
)

type Client struct {
	Host      string
	Port      int
	Username  string // length must <= 16
	Password  string
	Interface Interface

	debug bool

	// this flags controls which IPMI version (1.5 or 2.0) be used by Client to send Request
	v20 bool

	// holds data exchanged during Session Activation stage.
	// see: 13.14 IPMI v1.5 LAN Session Activation, 13.15 IPMI v2.0/RMCP+ Session Activation
	session *session

	udpClient  *UDPClient
	timeout    time.Duration
	bufferSize int
}

func NewClient(host string, port int, user string, pass string) (*Client, error) {
	if len(user) > IPMI_MAX_USER_NAME_LENGTH {
		return nil, fmt.Errorf("user name (%s) too long, exceed (%d) characters", user, IPMI_MAX_USER_NAME_LENGTH)
	}

	if len(user) == 0 {
		return nil, fmt.Errorf("empty username")
	}

	if len(pass) == 0 {
		return nil, fmt.Errorf("empty password")
	}

	c := &Client{
		Host:      host,
		Port:      port,
		Username:  user,
		Password:  pass,
		Interface: "",

		v20:        true,
		timeout:    time.Second * time.Duration(DefaultExchangeTimeoutSec),
		bufferSize: DefaultBufferSize,

		session: &session{
			// IPMI Request Sequence, start from 1
			ipmiSeq: 1,
			v20: v20{
				state: SessionStatePreSession,
			},
			v15: v15{
				active: false,
			},
		},
	}

	c.udpClient = &UDPClient{
		Host:       host,
		Port:       port,
		timeout:    c.timeout,
		bufferSize: c.bufferSize,
	}

	return c, nil
}

func (c *Client) WithInterface(intf Interface) *Client {
	c.Interface = intf
	return c
}

func (c *Client) WithDebug(debug bool) *Client {
	c.debug = debug
	return c
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.timeout = timeout
	c.udpClient.timeout = timeout
	return c
}

func (c *Client) WithBufferSize(bufferSize int) *Client {
	c.bufferSize = bufferSize
	c.udpClient.bufferSize = bufferSize
	return c
}

func (c *Client) SessionPrivilegeLevel() PrivilegeLevel {
	return c.session.v20.maxPrivilegeLevel
}

func (c *Client) Connect() error {
	// Optional RMCP Ping/Pong mechanism
	// pongRes, err := c.RmcpPing()
	// if err != nil {
	// 	return fmt.Errorf("RMCP Ping failed, err: %s", err)
	// }
	// if pongRes.IPMISupported {
	// 	return fmt.Errorf("ipmi not supported")
	// }

	if c.Interface == "" {
		return c.ConnectAuto()
	}

	if c.Interface == InterfaceLanplus {
		c.v20 = true
		return c.Connect20()
	}
	if c.Interface == InterfaceLan {
		c.v20 = false
		return c.Connect15()
	}

	return fmt.Errorf("not supported interface (lan or lanplus)")
}

// 13.14
// IPMI v1.5 LAN Session Activation
// 1. RmcpresencePing - PMCPPresencePong
// 2. Get Channel Authentication Capabilities
// 3. Get Session Challenge
// 4. Activate Session
func (c *Client) Connect15() error {
	var (
		err            error
		channelNumber  uint8          = 0x0e // Eh = retrieve information for channel this request was issued on
		privilegeLevel PrivilegeLevel = PrivilegeLevelAdministrator
	)

	_, err = c.GetChannelAuthenticationCapabilities(channelNumber, privilegeLevel)
	if err != nil {
		return fmt.Errorf("GetChannelAuthenticationCapabilities failed, err: %s", err)
	}

	_, err = c.GetSessionChallenge()
	if err != nil {
		return fmt.Errorf("GetSessionChallenge failed, err: %s", err)
	}

	c.session.v15.preSession = true

	_, err = c.ActivateSession()
	if err != nil {
		return fmt.Errorf("ActivateSession failed, err: %s", err)
	}

	_, err = c.SetSessionPrivilegeLevel(PrivilegeLevelAdministrator)
	if err != nil {
		return fmt.Errorf("SetSessionPrivilegeLevel failed, err: %s", err)
	}

	return nil

}

// see 13.15 IPMI v2.0/RMCP+ Session Activation
func (c *Client) Connect20() error {
	var (
		err error

		// 0h-Bh,Fh = specific channel number
		// Eh = retrieve information for channel this request was issued on
		channelNumber uint8 = 0x0e

		privilegeLevel PrivilegeLevel = PrivilegeLevelAdministrator
	)

	_, err = c.GetChannelAuthenticationCapabilities(channelNumber, privilegeLevel)
	if err != nil {
		return fmt.Errorf("cmd: Get Channel Authentication Capabilities failed, err: %s", err)
	}

	// Todo, retry for opensession/rakp1/rakp3
	_, err = c.OpenSession()
	if err != nil {
		return fmt.Errorf("cmd: RMCP+ Open Session failed, err: %s", err)
	}

	_, err = c.RAKPMessage1()
	if err != nil {
		return fmt.Errorf("cmd: rakp1 failed, err: %s", err)
	}

	_, err = c.RAKPMessage3()
	if err != nil {
		return fmt.Errorf("cmd: rakp3 failed, err: %s", err)
	}

	return nil
}

// ConnectAuto detects the IPMI version supported by BMC by using
// GetChannelAuthenticaitonCapabilities commmand, then decide to use v1.5 or v2.0
// for subsequent requests.
func (c *Client) ConnectAuto() error {
	var (
		err error

		// 0h-Bh,Fh = specific channel number
		// Eh = retrieve information for channel this request was issued on
		channelNumber uint8 = 0x0e

		privilegeLevel PrivilegeLevel = PrivilegeLevelAdministrator
	)

	// force use IPMI v1.5 first
	c.v20 = false
	cap, err := c.GetChannelAuthenticationCapabilities(channelNumber, privilegeLevel)
	if err != nil {
		return fmt.Errorf("cmd: Get Channel Authentication Capabilities failed, err: %s", err)
	}
	if cap.SupportIPMIv20 {
		c.v20 = true
		return c.Connect20()
	}
	if cap.SupportIPMIv15 {
		return c.Connect15()
	}
	return fmt.Errorf("client does not support IPMI v1.5 and IPMI v.20")
}

func (c *Client) Close() error {
	var sessionID uint32
	if c.v20 {
		sessionID = c.session.v20.bmcSessionID
	} else {
		sessionID = c.session.v15.sessionID
	}

	request := &CloseSessionRequest{
		SessionID: sessionID,
	}
	if _, err := c.CloseSession(request); err != nil {
		return fmt.Errorf("CloseSession failed, err: %s", err)
	}

	if err := c.udpClient.Close(); err != nil {
		return fmt.Errorf("close udp connection failed, err: %s", err)
	}

	return nil
}

func (c *Client) Exchange(request Request, response Response) error {
	c.Debug(">> Command Request", request)

	rmcp, err := c.BuildRmcpRequest(request)
	if err != nil {
		return fmt.Errorf("build RMCP+ request msg failed, err: %s", err)
	}
	c.Debug(">>>>>> RMCP Request", rmcp)
	sent := rmcp.Pack()
	c.DebugBytes("sent", sent, 16)

	ctx := context.Background()
	recv, err := c.udpClient.Exchange(ctx, bytes.NewReader(sent))
	if err != nil {
		return fmt.Errorf("client udp exchange msg failed, err: %s", err)
	}
	c.DebugBytes("recv", recv, 16)

	if err := c.ParseRmcpResponse(recv, response); err != nil {
		// Warn, must directly return err.
		// The error returned by ParseRmcpResponse might be of *ResponseError type.
		return err
	}

	c.Debug("<< Commmand Response", response)
	return nil
}

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

	maxPrivilegeLevel PrivilegeLevel
	sessionID         uint32

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

	// filled by RmcpOpenSessionRequest
	requestedAuthAlg      AuthAlg
	requestedIntegrityAlg IntegrityAlg
	requestedEncryptAlg   CryptAlg

	// filled by RmcpOpenSessionResponse
	// RMCP Open Session is used for exchanging session ids
	authAlg           AuthAlg
	integrityAlg      IntegrityAlg
	cryptAlg          CryptAlg
	maxPrivilegeLevel PrivilegeLevel // uint8 requestedRole sent in RAKP 1 message
	role              uint8          // whole byte of priviledge level in RAKP1, will be used for computing authcode of rakp2, rakp3
	consoleSessionID  uint32
	bmcSessionID      uint32

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

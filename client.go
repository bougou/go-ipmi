package ipmi

import (
	"bytes"
	"context"
	"fmt"
	"time"
)

const (
	DefaultExchangeTimeoutSec int = 120
	DefaultBufferSize         int = 1024
)

type Client struct {
	Host      string
	Port      int
	Username  string // length must <= 16
	Password  string
	Interface string

	debug bool

	v20 bool

	usernamePad16 []byte
	passwordPad16 []byte
	passwordPad20 []byte

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
		Interface: "lanplus",

		v20:        true,
		timeout:    time.Second * time.Duration(DefaultExchangeTimeoutSec),
		bufferSize: DefaultBufferSize,

		session: &session{
			v20: v20{
				state: SessionStatePreSession,
			},
			v15: v15{
				active: false,
			},
		},
	}

	l := len([]byte(c.Password))
	if l >= 20 {
		c.passwordPad20 = []byte(c.Password)[:20]
	} else {
		c.passwordPad20 = []byte(c.Password)
		for i := 0; i < 20-l; i++ {
			c.passwordPad20 = append(c.passwordPad20, 0x00)
		}
	}
	if l >= 16 {
		c.passwordPad16 = []byte(c.Password)[:16]
	} else {
		c.passwordPad16 = []byte(c.Password)
		for i := 0; i < 16-l; i++ {
			c.passwordPad16 = append(c.passwordPad16, 0x00)
		}
	}

	lu := len([]byte(c.Username))
	if lu >= 16 {
		c.usernamePad16 = []byte(c.Username)[:16]
	} else {
		c.usernamePad16 = []byte(c.Username)
		for i := 0; i < 20-lu; i++ {
			c.usernamePad16 = append(c.usernamePad16, 0x00)
		}
	}

	c.udpClient = &UDPClient{
		Host:       host,
		Port:       port,
		timeout:    c.timeout,
		bufferSize: c.bufferSize,
	}

	return c, nil
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

	if c.Interface == "lanplus" {
		c.v20 = true
		return c.Connect20()
	}
	if c.Interface == "lan" {
		c.v20 = false

		return c.Connect15()
	}

	// Todo, if c.Interface not specified,
	// first try v1.5 to find detect version from GetChannelAuthenticaitonCapabilities
	// then decide to use v1.5 or v2.0
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

	_, err = c.ActivateSession()
	if err != nil {
		return fmt.Errorf("ActivateSession failed, err: %s", err)
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
	recv, err := c.udpClient.Exchanged(ctx, bytes.NewReader(sent))
	if err != nil {
		return fmt.Errorf("client udp exchange msg failed, err: %s", err)
	}
	c.DebugBytes("recv", recv, 16)

	if err := c.ParseRmcpResponse(recv, response); err != nil {
		return fmt.Errorf("build rmcp response failed, err: %s", err)
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
	active            bool
	maxPrivilegeLevel PrivilegeLevel
	sessionID         uint32
	inSeq             uint32
	outSeq            uint32 // 6.12.12 IPMI v1.5 Outbound Session Sequence Number Tracking and Handling
	challenge         [16]byte
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
	// User passwords (keys, Kuid) are set using the Set User Password command.
	// Kg is set using the Set Channel Security Keys command.
	bmcKey []byte // BMC key, known as Kg
	// ipmi user password, the pre-shared key, known as Kuid

	accumulatedPayloadSize uint32

	// for xRC4 encryption
	rc4EncryptIV [16]byte
	rc4DecryptIV [16]byte
}

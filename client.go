package ipmi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type Interface string

const (
	InterfaceLan     Interface = "lan"
	InterfaceLanplus Interface = "lanplus"
	InterfaceOpen    Interface = "open"
	InterfaceTool    Interface = "tool"

	DefaultExchangeTimeoutSec   int = 20
	DefaultKeepAliveIntervalSec int = 30
	DefaultBufferSize           int = 1024
)

type Client struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Interface Interface

	debug bool

	maxPrivilegeLevel PrivilegeLevel

	responderAddr uint8
	responderLUN  uint8
	requesterAddr uint8
	requesterLUN  uint8

	openipmi *openipmi
	session  *session

	// this flags controls which IPMI version (1.5 or 2.0) be used by Client to send Request
	v20 bool

	udpClient  *UDPClient
	timeout    time.Duration
	bufferSize int

	// retryCount specifies the number of additional attempts to make after an initial failure.
	// For lan/lanplus interfaces, retries only occur when UDP exchanges timeout.
	// A value of 0 means no retries (only one attempt), 1 means one retry (two attempts total), etc.
	retryCount    int
	retryInterval time.Duration

	l sync.Mutex

	// closedCh is closed when Client.Close() is called.
	// used to notify other goroutines that Client is closed.
	closedCh chan bool
}

func NewOpenClient() (*Client, error) {
	myAddr := BMC_SA

	return &Client{
		Interface:  InterfaceOpen,
		timeout:    time.Second * time.Duration(DefaultExchangeTimeoutSec),
		bufferSize: DefaultBufferSize,

		openipmi: &openipmi{
			myAddr:     myAddr,
			targetAddr: myAddr,
		},
	}, nil
}

// NewToolClient creates an IPMI client based ipmitool.
// You should pass the file path of ipmitool binary or path of a wrapper script
// that would be executed.
func NewToolClient(path string) (*Client, error) {

	return &Client{
		Host:      path,
		Interface: InterfaceTool,
	}, nil
}

func NewClient(host string, port int, user string, pass string) (*Client, error) {
	if len(user) > IPMI_MAX_USER_NAME_LENGTH {
		return nil, fmt.Errorf("user name (%s) too long, exceed (%d) characters", user, IPMI_MAX_USER_NAME_LENGTH)
	}

	c := &Client{
		Host:      host,
		Port:      port,
		Username:  user,
		Password:  pass,
		Interface: InterfaceLanplus,

		v20:        true,
		timeout:    time.Second * time.Duration(DefaultExchangeTimeoutSec),
		bufferSize: DefaultBufferSize,

		retryCount:    0,
		retryInterval: 0,

		maxPrivilegeLevel: PrivilegeLevelUnspecified,

		responderAddr: BMC_SA,
		responderLUN:  uint8(IPMB_LUN_BMC),
		requesterAddr: RemoteConsole_SWID,
		requesterLUN:  0x00,

		session: &session{
			// IPMI Request Sequence, start from 1
			ipmiSeq: 1,
			v20: v20{
				state:         SessionStatePreSession,
				cipherSuiteID: CipherSuiteIDReserved,
			},
			v15: v15{
				active: false,
			},
		},

		closedCh: make(chan bool),
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

func (c *Client) WithUDPProxy(proxy proxy.Dialer) *Client {
	if c.udpClient != nil {
		c.udpClient.SetProxy(proxy)
	}
	return c
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.timeout = timeout

	if c.udpClient != nil {
		c.udpClient.timeout = timeout
	}
	return c
}

func (c *Client) WithBufferSize(bufferSize int) *Client {
	c.bufferSize = bufferSize

	if c.udpClient != nil {
		c.udpClient.bufferSize = bufferSize
	}
	return c
}

func (c *Client) WithRetry(retryCount int, retryInterval time.Duration) *Client {
	c.retryCount = retryCount
	c.retryInterval = retryInterval
	return c
}

// WithCipherSuiteID sets a custom cipher suite which is used during OpenSession command.
// It is only valid for client with IPMI lanplus interface.
// For the custom cipherSuiteID to take effect, you must call WithCipherSuiteID before calling Connect method.
func (c *Client) WithCipherSuiteID(cipherSuiteID ...CipherSuiteID) *Client {
	if c.session != nil {
		if len(cipherSuiteID) > 1 {
			c.session.v20.customSuiteIDs = cipherSuiteID
		} else {
			c.session.v20.cipherSuiteID = cipherSuiteID[0]
		}
	}
	return c
}

// WithMaxPrivilegeLevel sets a specified session privilege level to use.
func (c *Client) WithMaxPrivilegeLevel(privilegeLevel PrivilegeLevel) *Client {
	c.maxPrivilegeLevel = privilegeLevel
	return c
}

func (c *Client) WithResponderAddr(responderAddr, responderLUN uint8) {
	c.responderAddr = responderAddr
	c.responderLUN = responderLUN
}
func (c *Client) WithRequesterAddr(requesterAddr, requesterLUN uint8) {
	c.requesterAddr = requesterAddr
	c.requesterLUN = requesterLUN
}

func (c *Client) SessionPrivilegeLevel() PrivilegeLevel {
	return c.maxPrivilegeLevel
}

// Connect connects to the bmc by specified Interface.
func (c *Client) Connect(ctx context.Context) error {
	// Optional RMCP Ping/Pong mechanism
	// pongRes, err := c.RmcpPing()
	// if err != nil {
	// return fmt.Errorf("RMCP Ping failed, err: %w", err)
	// }
	// if pongRes.IPMISupported {
	// return fmt.Errorf("ipmi not supported")
	// }

	switch c.Interface {
	case "", InterfaceOpen:
		var devnum int32 = 0
		return c.ConnectOpen(ctx, devnum)

	case InterfaceTool:
		var devnum int32 = 0
		return c.ConnectTool(ctx, devnum)

	case InterfaceLanplus:
		c.v20 = true
		return c.Connect20(ctx)

	case InterfaceLan:
		c.v20 = false
		return c.Connect15(ctx)

	default:
		return fmt.Errorf("not supported interface, supported: lan,lanplus,open")
	}
}

func (c *Client) Close(ctx context.Context) error {
	switch c.Interface {
	case "", InterfaceOpen:
		return c.closeOpen(ctx)

	case InterfaceTool:
		return c.closeTool(ctx)

	case InterfaceLan, InterfaceLanplus:
		return c.closeLAN(ctx)
	}

	return nil
}

func (c *Client) Exchange(ctx context.Context, request Request, response Response) error {
	switch c.Interface {
	case "", InterfaceOpen:
		return c.exchangeOpen(ctx, request, response)

	case InterfaceTool:
		return c.exchangeTool(ctx, request, response)

	case InterfaceLan, InterfaceLanplus:
		return c.exchangeLAN(ctx, request, response)

	}

	return nil
}

func (c *Client) lock() {
	c.l.Lock()
}

func (c *Client) unlock() {
	c.l.Unlock()
}

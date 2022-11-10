package ipmi

import (
	"fmt"
	"time"
)

type Interface string

const (
	InterfaceLan     Interface = "lan"
	InterfaceLanplus Interface = "lanplus"
	InterfaceOpen    Interface = "open"
	InterfaceTool    Interface = "tool"

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

	openipmi *openipmi
	session  *session

	// this flags controls which IPMI version (1.5 or 2.0) be used by Client to send Request
	v20 bool

	udpClient  *UDPClient
	timeout    time.Duration
	bufferSize int
}

func NewOpenClient() (*Client, error) {
	myAddr := BMC_SA

	return &Client{
		Interface: "open",

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
		Interface: "tool",
	}, nil
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

// Connect connects to the bmc by specified Interface.
func (c *Client) Connect() error {
	// Optional RMCP Ping/Pong mechanism
	// pongRes, err := c.RmcpPing()
	// if err != nil {
	// 	return fmt.Errorf("RMCP Ping failed, err: %s", err)
	// }
	// if pongRes.IPMISupported {
	// 	return fmt.Errorf("ipmi not supported")
	// }

	switch c.Interface {
	case "", InterfaceOpen:
		var devnum int32 = 0
		return c.ConnectOpen(devnum)

	case InterfaceTool:
		var devnum int32 = 0
		return c.ConnectTool(devnum)

	case InterfaceLanplus:
		c.v20 = true
		return c.Connect20()

	case InterfaceLan:
		c.v20 = false
		return c.Connect15()

	default:
		return fmt.Errorf("not supported interface, supported: lan,lanplus,open")
	}
}

func (c *Client) Close() error {
	switch c.Interface {
	case "", InterfaceOpen:
		return c.closeOpen()

	case InterfaceTool:
		return c.closeTool()

	case InterfaceLan, InterfaceLanplus:
		return c.closeLAN()
	}

	return nil
}

func (c *Client) Exchange(request Request, response Response) error {
	switch c.Interface {
	case "", InterfaceOpen:
		return c.exchangeOpen(request, response)

	case InterfaceTool:
		return c.exchangeTool(request, response)

	case InterfaceLan, InterfaceLanplus:
		return c.exchangeLAN(request, response)

	}

	return nil
}

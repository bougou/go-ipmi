package ipmi

import (
	client "github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	DefaultBufferSize              = client.DefaultBufferSize
	DefaultKeepAliveIntervalSec    = client.DefaultKeepAliveIntervalSec
	DefaultLanRetries              = client.DefaultLanRetries
	DefaultLanTimeoutSec           = client.DefaultLanTimeoutSec
	DefaultLanplusRetries          = client.DefaultLanplusRetries
	DefaultLanplusTimeoutSec       = client.DefaultLanplusTimeoutSec
	DefaultOpenRetries             = client.DefaultOpenRetries
	DefaultOpenTimeoutSec          = client.DefaultOpenTimeoutSec
	IPMIVersion15                  = client.IPMIVersion15
	IPMIVersion20                  = client.IPMIVersion20
	IPMI_MAX_USER_NAME_LENGTH      = rmcpplus.MaxUserNameLength
	IPMI_RAKP1_MESSAGE_SIZE        = rmcpplus.RAKP1MessageSize
	InterfaceLan                   = client.InterfaceLan
	InterfaceLanplus               = client.InterfaceLanplus
	InterfaceOpen                  = client.InterfaceOpen
	InterfaceTool                  = client.InterfaceTool
	RmcpOpenSessionRequestSize     = rmcpplus.OpenSessionRequestSize
	RmcpOpenSessionResponseMinSize = rmcpplus.OpenSessionResponseMinSize
	RmcpOpenSessionResponseSize    = rmcpplus.OpenSessionResponseSize
)

var (
	ErrDCMIGroupExtensionIDMismatch = types.ErrDCMIGroupExtensionIDMismatch
	ErrUnpackedDataTooShort         = types.ErrUnpackedDataTooShort
)

type (
	AuthCodeMultiSessionInput  = client.AuthCodeMultiSessionInput
	AuthCodeSingleSessionInput = client.AuthCodeSingleSessionInput
	AuthenticationPayload      = rmcpplus.AuthenticationPayload
	Client                     = client.Client
	CommandContext             = client.CommandContext
	CommandRawRequest          = client.CommandRawRequest
	CommandRawResponse         = client.CommandRawResponse
	ConfidentialityPayload     = rmcpplus.ConfidentialityPayload
	IntegrityPayload           = rmcpplus.IntegrityPayload
	Interface                  = client.Interface
	OpenSessionRequest         = rmcpplus.OpenSessionRequest
	OpenSessionResponse        = rmcpplus.OpenSessionResponse
	RAKPMessage1               = rmcpplus.RAKPMessage1
	RAKPMessage2               = rmcpplus.RAKPMessage2
	RAKPMessage3               = rmcpplus.RAKPMessage3
	RAKPMessage4               = rmcpplus.RAKPMessage4
	RmcpPingRequest            = client.RmcpPingRequest
	RmcpPingResponse           = client.RmcpPingResponse
	SensorFilterOption         = client.SensorFilterOption
	SOLActivateOptions         = client.SOLActivateOptions
	UDPClient                  = client.UDPClient
)

var (
	CheckDCMIGroupExenstionMatch        = types.CheckDCMIGroupExenstionMatch
	ErrDCMIGroupExtensionIDMismatchWith = types.ErrDCMIGroupExtensionIDMismatchWith
	ErrNotEnoughDataWith                = types.ErrNotEnoughDataWith
	ErrUnpackedDataTooShortWith         = types.ErrUnpackedDataTooShortWith
	GetCommandContext                   = client.GetCommandContext
	NewClient                           = client.NewClient
	NewOpenClient                       = client.NewOpenClient
	NewToolClient                       = client.NewToolClient
	NewUDPClient                        = client.NewUDPClient
	RenderTable                         = client.RenderTable
	RenderTableStream                   = client.RenderTableStream
	SensorFilterOptionIsReadingValid    = client.SensorFilterOptionIsReadingValid
	SensorFilterOptionIsSensorType      = client.SensorFilterOptionIsSensorType
	SensorFilterOptionIsThreshold       = client.SensorFilterOptionIsThreshold
	WithCommandContext                  = client.WithCommandContext
)

package ipmi

import client "github.com/bougou/go-ipmi/pkg/client"

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
	IPMI_MAX_USER_NAME_LENGTH      = client.IPMI_MAX_USER_NAME_LENGTH
	IPMI_RAKP1_MESSAGE_SIZE        = client.IPMI_RAKP1_MESSAGE_SIZE
	InterfaceLan                   = client.InterfaceLan
	InterfaceLanplus               = client.InterfaceLanplus
	InterfaceOpen                  = client.InterfaceOpen
	InterfaceTool                  = client.InterfaceTool
	RmcpOpenSessionRequestSize     = client.RmcpOpenSessionRequestSize
	RmcpOpenSessionResponseMinSize = client.RmcpOpenSessionResponseMinSize
	RmcpOpenSessionResponseSize    = client.RmcpOpenSessionResponseSize
)

var (
	ErrDCMIGroupExtensionIDMismatch = client.ErrDCMIGroupExtensionIDMismatch
	ErrUnpackedDataTooShort         = client.ErrUnpackedDataTooShort
)

type (
	AuthCodeMultiSessionInput  = client.AuthCodeMultiSessionInput
	AuthCodeSingleSessionInput = client.AuthCodeSingleSessionInput
	AuthenticationPayload      = client.AuthenticationPayload
	Client                     = client.Client
	CommandContext             = client.CommandContext
	CommandRawRequest          = client.CommandRawRequest
	CommandRawResponse         = client.CommandRawResponse
	ConfidentialityPayload     = client.ConfidentialityPayload
	IntegrityPayload           = client.IntegrityPayload
	Interface                  = client.Interface
	OpenSessionRequest         = client.OpenSessionRequest
	OpenSessionResponse        = client.OpenSessionResponse
	RAKPMessage1               = client.RAKPMessage1
	RAKPMessage2               = client.RAKPMessage2
	RAKPMessage3               = client.RAKPMessage3
	RAKPMessage4               = client.RAKPMessage4
	RmcpPingRequest            = client.RmcpPingRequest
	RmcpPingResponse           = client.RmcpPingResponse
	SensorFilterOption         = client.SensorFilterOption
	SOLActivateOptions         = client.SOLActivateOptions
	UDPClient                  = client.UDPClient
)

var (
	CheckDCMIGroupExenstionMatch        = client.CheckDCMIGroupExenstionMatch
	ErrDCMIGroupExtensionIDMismatchWith = client.ErrDCMIGroupExtensionIDMismatchWith
	ErrNotEnoughDataWith                = client.ErrNotEnoughDataWith
	ErrUnpackedDataTooShortWith         = client.ErrUnpackedDataTooShortWith
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

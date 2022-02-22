package ipmi

// Command is the field in an IPMI message
type Command struct {
	ID    uint8
	NetFn NetFn
	Name  string
}

type Request interface {
	// Pack encodes the object to data bytes
	Pack() []byte
	// Command return the IPMI command info (NetFn/Cmd).
	// All IPMI specification specified commands are already predefined in this file.
	Command() Command
}

type Response interface {
	// Unpack decodes the object from data bytes
	Unpack(data []byte) error
	// CompletionCodes returns a map of command-specific completion codes
	CompletionCodes() map[uint8]string
	// Format return a formatted human friendly string
	Format() string
}

// ResponseError encapsulate the CompletionCode of IPMI Response Msg
// alongside with error description.
type ResponseError struct {
	completionCode CompletionCode
	description    string
}

// Error implements the error interface
func (e *ResponseError) Error() string {
	return e.description
}

func (e *ResponseError) CompletionCode() CompletionCode {
	return e.completionCode
}

// Appendix G - Command Assignments
// Command Number Assignments (Appendix G, table G-1)
var (
	// a faked command for RAKP messages
	CommandNone = Command{}

	// IPM Device Global Commands
	CommandGetDeviceID                        = Command{ID: 0x01, NetFn: NetFnAppRequest, Name: "Get Device ID"}
	CommandColdReset                          = Command{ID: 0x02, NetFn: NetFnAppRequest, Name: "Cold Reset"}
	CommandWarmReset                          = Command{ID: 0x03, NetFn: NetFnAppRequest, Name: "Warm Reset"}
	CommandGetSelfTestResults                 = Command{ID: 0x04, NetFn: NetFnAppRequest, Name: "Get Self Test Results"}
	CommandManufacturingTestOn                = Command{ID: 0x05, NetFn: NetFnAppRequest, Name: "Manufacturing Test On"}
	CommandSetACPIPowerState                  = Command{ID: 0x06, NetFn: NetFnAppRequest, Name: "Set ACPI Power State"}
	CommandGetACPIPowerState                  = Command{ID: 0x07, NetFn: NetFnAppRequest, Name: "Get ACPI Power State"}
	CommandGetDeviceGUID                      = Command{ID: 0x08, NetFn: NetFnAppRequest, Name: "Get Device GUID"}
	CommandGetNetFnSupport                    = Command{ID: 0x09, NetFn: NetFnAppRequest, Name: "Get NetFn Support"}
	CommandGetCommandSupport                  = Command{ID: 0x0a, NetFn: NetFnAppRequest, Name: "Get Command Support"}
	CommandGetCommandSubfunctionSupport       = Command{ID: 0x0b, NetFn: NetFnAppRequest, Name: "Get Command Sub-function Support"}
	CommandGetConfigurableCommands            = Command{ID: 0x0c, NetFn: NetFnAppRequest, Name: "Get Configurable Commands"}
	CommandGetConfigurableCommandSubfunctions = Command{ID: 0x0d, NetFn: NetFnAppRequest, Name: "Get Configurable Command Sub-functions"} // 0Eh - 0Fh reserved
	CommandSetCommandEnables                  = Command{ID: 0x60, NetFn: NetFnAppRequest, Name: "Set Command Enables"}
	CommandGetCommandEnables                  = Command{ID: 0x61, NetFn: NetFnAppRequest, Name: "Get Command Enables"}
	CommandSetCommandSubfunctionsEnables      = Command{ID: 0x62, NetFn: NetFnAppRequest, Name: "Set Command Sub-function Enables"}
	CommandGetCommandSubfunctionsEnables      = Command{ID: 0x63, NetFn: NetFnAppRequest, Name: "Get Command Sub-function Enables"}
	CommandGetOEMNetFnIanaSupport             = Command{ID: 0x64, NetFn: NetFnAppRequest, Name: "Get OEM NetFn IANA Support"}

	// BMC Watchdog Timer Commands
	CommandResetWatchdogTimer = Command{ID: 0x22, NetFn: NetFnAppRequest, Name: "Reset Watchdog Timer"}
	CommandSetWatchdogTimer   = Command{ID: 0x24, NetFn: NetFnAppRequest, Name: "Set Watchdog Timer"}
	CommandGetWatchdogTimer   = Command{ID: 0x25, NetFn: NetFnAppRequest, Name: "Get Watchdog Timer"}

	// BMC Device and Messaging Commands
	CommandSetBMCGlobalEnables            = Command{ID: 0x2e, NetFn: NetFnAppRequest, Name: "Set BMC Global Enables"}
	CommandGetBMCGlobalEnables            = Command{ID: 0x2f, NetFn: NetFnAppRequest, Name: "Get BMC Global Enables"}
	CommandClearMessageFlags              = Command{ID: 0x30, NetFn: NetFnAppRequest, Name: "Clear Message Flags"}
	CommandGetMessageFlags                = Command{ID: 0x31, NetFn: NetFnAppRequest, Name: "Get Message Flags"}
	CommandEnableMessageChannelReceive    = Command{ID: 0x32, NetFn: NetFnAppRequest, Name: "Enable Message Channel Receive"}
	CommandGetMessage                     = Command{ID: 0x33, NetFn: NetFnAppRequest, Name: "Get Message"}
	CommandSendMessage                    = Command{ID: 0x34, NetFn: NetFnAppRequest, Name: "Send Message"}
	CommandReadEventMessageBuffer         = Command{ID: 0x35, NetFn: NetFnAppRequest, Name: "Read Event Message Buffer"}
	CommandGetBTInterfaceCapabilities     = Command{ID: 0x36, NetFn: NetFnAppRequest, Name: "Get BT Interface Capabilities"}
	CommandGetSystemGUID                  = Command{ID: 0x37, NetFn: NetFnAppRequest, Name: "Get System GUID"}
	CommandSetSystemInfoParameters        = Command{ID: 0x58, NetFn: NetFnAppRequest, Name: "Set System Info Parameters"}
	CommandGetSystemInfoParameters        = Command{ID: 0x59, NetFn: NetFnAppRequest, Name: "Get System Info Parameters"}
	CommandGetChannelAuthCapabilities     = Command{ID: 0x38, NetFn: NetFnAppRequest, Name: "Get Channel Authentication Capabilities"}
	CommandGetSessionChallenge            = Command{ID: 0x39, NetFn: NetFnAppRequest, Name: "Get Session Challenge"}
	CommandActivateSession                = Command{ID: 0x3a, NetFn: NetFnAppRequest, Name: "Activate Session"}
	CommandSetSessionPrivilegeLevel       = Command{ID: 0x3b, NetFn: NetFnAppRequest, Name: "Set Session Privilege Level"}
	CommandCloseSession                   = Command{ID: 0x3c, NetFn: NetFnAppRequest, Name: "Close Session"}
	CommandGetSessionInfo                 = Command{ID: 0x3d, NetFn: NetFnAppRequest, Name: "Get Session Info"} // 3e unassigned
	CommandGetAuthCode                    = Command{ID: 0x3f, NetFn: NetFnAppRequest, Name: "Get AuthCode"}
	CommandSetChannelAccess               = Command{ID: 0x40, NetFn: NetFnAppRequest, Name: "Set Channel Access"}
	CommandGetChannelAccess               = Command{ID: 0x41, NetFn: NetFnAppRequest, Name: "Get Channel Access"}
	CommandGetChannelInfo                 = Command{ID: 0x42, NetFn: NetFnAppRequest, Name: "Get Channel Info Command"}
	CommandSetUserAccess                  = Command{ID: 0x43, NetFn: NetFnAppRequest, Name: "Set User Access Command"}
	CommandGetUserAccess                  = Command{ID: 0x44, NetFn: NetFnAppRequest, Name: "Get User Access Command"}
	CommandSetUsername                    = Command{ID: 0x45, NetFn: NetFnAppRequest, Name: "Set User Name"}
	CommandGetUsername                    = Command{ID: 0x46, NetFn: NetFnAppRequest, Name: "Get User Name Command"}
	CommandSetUserPassword                = Command{ID: 0x47, NetFn: NetFnAppRequest, Name: "Set User Password Command"}
	CommandActivatePayload                = Command{ID: 0x48, NetFn: NetFnAppRequest, Name: "Activate Payload"}
	CommandDeactivatePayload              = Command{ID: 0x49, NetFn: NetFnAppRequest, Name: "Deactivate Payload"}
	CommandGetPayloadActivationStatus     = Command{ID: 0x4a, NetFn: NetFnAppRequest, Name: "Get Payload Activation Status"}
	CommandGetPayloadInstanceInfo         = Command{ID: 0x4b, NetFn: NetFnAppRequest, Name: "Get Payload Instance Info"}
	CommandSetUserPayloadAccess           = Command{ID: 0x4c, NetFn: NetFnAppRequest, Name: "Set User Payload Access"}
	CommandGetUserPayloadAccess           = Command{ID: 0x4d, NetFn: NetFnAppRequest, Name: "Get User Payload Access"}
	CommandGetChannelPayloadSupport       = Command{ID: 0x4e, NetFn: NetFnAppRequest, Name: "Get Channel Payload Support"}
	CommandGetChannelPayloadVersion       = Command{ID: 0x4f, NetFn: NetFnAppRequest, Name: "Get Channel Payload Version"}
	CommandGetChannelOEMPayloadInfo       = Command{ID: 0x50, NetFn: NetFnAppRequest, Name: "Get Channel OEM Payload Info"} // 51 unassigned
	CommandMasterWriteRead                = Command{ID: 0x52, NetFn: NetFnAppRequest, Name: "Master Write-Read"}            // 53 unassigned
	CommandGetChannelCipherSuites         = Command{ID: 0x54, NetFn: NetFnAppRequest, Name: "Get Channel Cipher Suites"}
	CommandSuspendOrResumeEncryption      = Command{ID: 0x55, NetFn: NetFnAppRequest, Name: "Suspend/Resume Payload Encryption"}
	CommandSetChannelCipherSuites         = Command{ID: 0x56, NetFn: NetFnAppRequest, Name: "Set Channel Security Keys"}
	CommandGetSystemInterfaceCapabilities = Command{ID: 0x57, NetFn: NetFnAppRequest, Name: "Get System Interface Capabilities"}

	// Chassis Device Commands
	CommandGetChassisCapabilities = Command{ID: 0x00, NetFn: NetFnChassisRequest, Name: "Get Chassis Capabilities"}
	CommandGetChassisStatus       = Command{ID: 0x01, NetFn: NetFnChassisRequest, Name: "Get Chassis Status"}
	CommandChassisControl         = Command{ID: 0x02, NetFn: NetFnChassisRequest, Name: "Chassis Control"}
	CommandChassisReset           = Command{ID: 0x03, NetFn: NetFnChassisRequest, Name: "Chassis Reset"}
	CommandChassisIdentify        = Command{ID: 0x04, NetFn: NetFnChassisRequest, Name: "Chassis Identify"}
	CommandSetChassisCapabilities = Command{ID: 0x05, NetFn: NetFnChassisRequest, Name: "Set Chassis Capabilities"}
	CommandSetPowerRestorePolicy  = Command{ID: 0x06, NetFn: NetFnChassisRequest, Name: "Set Power Restore Policy"}
	CommandGetSystemRestartCause  = Command{ID: 0x07, NetFn: NetFnChassisRequest, Name: "Get System Restart Cause"}
	CommandSetSystemBootOptions   = Command{ID: 0x08, NetFn: NetFnChassisRequest, Name: "Set System Boot Options"}
	CommandGetSystemBootOptions   = Command{ID: 0x09, NetFn: NetFnChassisRequest, Name: "Get System Boot Options"}
	CommandSetFrontPanelEnables   = Command{ID: 0x0a, NetFn: NetFnChassisRequest, Name: "Set Front Panel Button Enables"}
	CommandSetPowerCycleInterval  = Command{ID: 0x0b, NetFn: NetFnChassisRequest, Name: "Set Power Cycle Interval"} // 0ch -0eh unassigned
	CommandGetPOHCounter          = Command{ID: 0x0f, NetFn: NetFnChassisRequest, Name: "Get POH Counter"}

	// Event Commands
	CommandSetEventReceiver = Command{ID: 0x00, NetFn: NetFnSensorEventRequest, Name: "Set Event Receiver"}
	CommandGetEventReceiver = Command{ID: 0x01, NetFn: NetFnSensorEventRequest, Name: "Get Event Receiver"}
	CommandEventMessage     = Command{ID: 0x02, NetFn: NetFnSensorEventRequest, Name: "Platform Event (Event Message)"} // 03h -0fh unassigned

	// PEF and Alerting Commands
	CommandGetPefCapabilities      = Command{ID: 0x10, NetFn: NetFnSensorEventRequest, Name: "Get PEF Capabilities"}
	CommandArmPefPostponeTimer     = Command{ID: 0x11, NetFn: NetFnSensorEventRequest, Name: "Arm PEF Postpone Timer"}
	CommandSetPefConfigParameters  = Command{ID: 0x12, NetFn: NetFnSensorEventRequest, Name: "Set PEF Configuration Parameters"}
	CommandGetPefConfigParameters  = Command{ID: 0x13, NetFn: NetFnSensorEventRequest, Name: "Get PEF Configuration Parameters"}
	CommandSetLastProcessedEventId = Command{ID: 0x14, NetFn: NetFnSensorEventRequest, Name: "Set Last Processed Event ID"}
	CommandGetLastProcessedEventId = Command{ID: 0x15, NetFn: NetFnSensorEventRequest, Name: "Get Last Processed Event ID"}
	CommandAlertImmediate          = Command{ID: 0x16, NetFn: NetFnSensorEventRequest, Name: "Alert Immediate"}
	CommandPetAck                  = Command{ID: 0x17, NetFn: NetFnSensorEventRequest, Name: "PET Acknowledge"}

	// Sensor Device Commands
	CommandGetDeviceSDRInfo               = Command{ID: 0x20, NetFn: NetFnSensorEventRequest, Name: "Get Device SDR Info"}
	CommandGetDeviceSDR                   = Command{ID: 0x21, NetFn: NetFnSensorEventRequest, Name: "Get Device SDR"}
	CommandReserveDeviceSDRRepo           = Command{ID: 0x22, NetFn: NetFnSensorEventRequest, Name: "Reserve Device SDR Repository"}
	CommandGetSensorReadingFactors        = Command{ID: 0x23, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Reading Factors"}
	CommandSetSensorHysteresis            = Command{ID: 0x24, NetFn: NetFnSensorEventRequest, Name: "Set Sensor Hysteresis"}
	CommandGetSensorHysteresis            = Command{ID: 0x25, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Hysteresis"}
	CommandSetSensorThresholds            = Command{ID: 0x26, NetFn: NetFnSensorEventRequest, Name: "Set Sensor Threshold"}
	CommandGetSensorThresholds            = Command{ID: 0x27, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Threshold"}
	CommandSetSensorEventEnable           = Command{ID: 0x28, NetFn: NetFnSensorEventRequest, Name: "Set Sensor Event Enable"}
	CommandGetSensorEventEnable           = Command{ID: 0x29, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Event Enable"}
	CommandRearmSensorEvents              = Command{ID: 0x2a, NetFn: NetFnSensorEventRequest, Name: "Re-arm Sensor Events"}
	CommandGetSensorEventStatus           = Command{ID: 0x2b, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Event Status"} // no 2c
	CommandGetSensorReading               = Command{ID: 0x2d, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Reading"}
	CommandSetSensorType                  = Command{ID: 0x2e, NetFn: NetFnSensorEventRequest, Name: "Set Sensor Type"}
	CommandGetSensorType                  = Command{ID: 0x2f, NetFn: NetFnSensorEventRequest, Name: "Get Sensor Type"}
	CommandSetSensorReadingAndEventStatus = Command{ID: 0x30, NetFn: NetFnSensorEventRequest, Name: "Set Sensor Reading And Event Status"}

	// FRU Device Commands
	CommandGetFRUInventoryAreaInfo = Command{ID: 0x10, NetFn: NetFnStorageRequest, Name: "Get FRU Inventory Area Info"}
	CommandReadFRUData             = Command{ID: 0x11, NetFn: NetFnStorageRequest, Name: "Read FRU Data"}
	CommandWriteFRUData            = Command{ID: 0x12, NetFn: NetFnStorageRequest, Name: "Write FRU Data"}

	// SDR Device Commands
	CommandGetSDRRepoInfo         = Command{ID: 0x20, NetFn: NetFnStorageRequest, Name: "Get SDR Repository Info"}
	CommandGetSDRRepoAllocInfo    = Command{ID: 0x21, NetFn: NetFnStorageRequest, Name: "Get SDR Repository Allocation Info"}
	CommandReserveSDRRepo         = Command{ID: 0x22, NetFn: NetFnStorageRequest, Name: "Reserve SDR Repository"}
	CommandGetSDR                 = Command{ID: 0x23, NetFn: NetFnStorageRequest, Name: "Get SDR"}
	CommandAddSDR                 = Command{ID: 0x24, NetFn: NetFnStorageRequest, Name: "Add SDR"}
	CommandPartialAddSDR          = Command{ID: 0x25, NetFn: NetFnStorageRequest, Name: "Partial Add SDR"}
	CommandDeleteSDR              = Command{ID: 0x26, NetFn: NetFnStorageRequest, Name: "Delete SDR"}
	CommandClearSDRRepo           = Command{ID: 0x27, NetFn: NetFnStorageRequest, Name: "Clear SDR Repository"}
	CommandGetSDRRepoTime         = Command{ID: 0x28, NetFn: NetFnStorageRequest, Name: "Get SDR Repository Time"}
	CommandSetSDRRepoTime         = Command{ID: 0x29, NetFn: NetFnStorageRequest, Name: "Set SDR Repository Time"}
	CommandEnterSDRRepoUpateMode  = Command{ID: 0x2a, NetFn: NetFnStorageRequest, Name: "Enter SDR Repository Update Mode"}
	CommandExitSDRRepoUpdateMode  = Command{ID: 0x2b, NetFn: NetFnStorageRequest, Name: "Exit SDR Repository Update Mode"}
	CommandRunInitializationAgent = Command{ID: 0x2c, NetFn: NetFnStorageRequest, Name: "Run Initialization Agent"}

	// SEL Device Commands
	CommandGetSELInfo          = Command{ID: 0x40, NetFn: NetFnStorageRequest, Name: "Get SEL Info"}
	CommandGetSELAllocInfo     = Command{ID: 0x41, NetFn: NetFnStorageRequest, Name: "Get SEL Allocation Info"}
	CommandReserveSEL          = Command{ID: 0x42, NetFn: NetFnStorageRequest, Name: "Reserve SEL"}
	CommandGetSELEntry         = Command{ID: 0x43, NetFn: NetFnStorageRequest, Name: "Get SEL Entry"}
	CommandAddSELEntry         = Command{ID: 0x44, NetFn: NetFnStorageRequest, Name: "Add SEL Entry"}
	CommandPartialAddSELEntry  = Command{ID: 0x45, NetFn: NetFnStorageRequest, Name: "Partial Add SEL Entry"}
	CommandDeleteSELEntry      = Command{ID: 0x46, NetFn: NetFnStorageRequest, Name: "Delete SEL Entry"}
	CommandClearSEL            = Command{ID: 0x47, NetFn: NetFnStorageRequest, Name: "Clear SEL"}
	CommandGetSELTime          = Command{ID: 0x48, NetFn: NetFnStorageRequest, Name: "Get SEL Time"}
	CommandSetSELTime          = Command{ID: 0x49, NetFn: NetFnStorageRequest, Name: "Set SEL Time"}
	CommandGetAuxLogStatus     = Command{ID: 0x5a, NetFn: NetFnStorageRequest, Name: "Get Auxiliary Log Status"}
	CommandSetAuxLogStatus     = Command{ID: 0x5b, NetFn: NetFnStorageRequest, Name: "Set Auxiliary Log Status"}
	CommandGetSELTimeUTCOffset = Command{ID: 0x5c, NetFn: NetFnStorageRequest, Name: "Get SEL Time UTC Offset"}
	CommandSetSELTimeUTCOffset = Command{ID: 0x5d, NetFn: NetFnStorageRequest, Name: "Set SEL Time UTC Offset"}

	// LAN Device Commands
	CommandSetLanConfigParams = Command{ID: 0x01, NetFn: NetFnTransportRequest, Name: "Set LAN Configuration Parameters"}
	CommandGetLanConfigParams = Command{ID: 0x02, NetFn: NetFnTransportRequest, Name: "Get LAN Configuration Parameters"}
	CommandSuspendARPs        = Command{ID: 0x03, NetFn: NetFnTransportRequest, Name: "Suspend BMC ARPs"}
	CommandGetIpStatistics    = Command{ID: 0x04, NetFn: NetFnTransportRequest, Name: "Get IP/UDP/RMCP Statistics"}

	// Serial/Modem Device Commands
	CommandSetSerialConfig        = Command{ID: 0x10, NetFn: NetFnTransportRequest, Name: "Set Serial/Modem Configuration"}
	CommandGetSerialConfig        = Command{ID: 0x11, NetFn: NetFnTransportRequest, Name: "Get Serial/Modem Configuration"}
	CommandSetSerialMux           = Command{ID: 0x12, NetFn: NetFnTransportRequest, Name: "Set Serial/Modem Mux"}
	CommandGetTapResponseCodes    = Command{ID: 0x13, NetFn: NetFnTransportRequest, Name: "Get TAP Response Codes"}
	CommandSetPPPTransmitData     = Command{ID: 0x14, NetFn: NetFnTransportRequest, Name: "Set PPP UDP Proxy Transmit Data"}
	CommandGetPPPTransmitData     = Command{ID: 0x15, NetFn: NetFnTransportRequest, Name: "Get PPP UDP Proxy Transmit Data"}
	CommandSendPPPPacket          = Command{ID: 0x16, NetFn: NetFnTransportRequest, Name: "Send PPP UDP Proxy Packet"}
	CommandGetPPPReceiveData      = Command{ID: 0x17, NetFn: NetFnTransportRequest, Name: "Get PPP UDP Proxy Receive Data"}
	CommandSerialConnectionActive = Command{ID: 0x18, NetFn: NetFnTransportRequest, Name: "Serial/Modem Connection Active"}
	CommandCallback               = Command{ID: 0x19, NetFn: NetFnTransportRequest, Name: "Callback"}
	CommandSetUserCallbackOptions = Command{ID: 0x1a, NetFn: NetFnTransportRequest, Name: "Set User Callback Options"}
	CommandGetUserCallbackOptions = Command{ID: 0x1b, NetFn: NetFnTransportRequest, Name: "Get User Callback Options"}
	CommandSetSerialRoutingMux    = Command{ID: 0x1c, NetFn: NetFnTransportRequest, Name: "Set Serial Routing Mux"}
	CommandSOLActivating          = Command{ID: 0x20, NetFn: NetFnTransportRequest, Name: "SOL Activating"}
	CommandSetSOLConfigParams     = Command{ID: 0x21, NetFn: NetFnTransportRequest, Name: "Set SOL Configuration Parameters"}
	CommandGetSOLConfigParams     = Command{ID: 0x22, NetFn: NetFnTransportRequest, Name: "Get SOL Configuration Parameters"}

	// Command Forwarding Commands
	CommandFowarded        = Command{ID: 0x30, NetFn: NetFnTransportRequest, Name: "Forwarded Command"}
	CommandSetForwarded    = Command{ID: 0x31, NetFn: NetFnTransportRequest, Name: "Set Forwarded Commands"}
	CommandGetForwarded    = Command{ID: 0x32, NetFn: NetFnTransportRequest, Name: "Get Forwarded Commands"}
	CommandEnableForwarded = Command{ID: 0x33, NetFn: NetFnTransportRequest, Name: "Enable Forwarded Commands"}

	// Bridge Management Commands (ICMB)
	CommandGetBridgeState        = Command{ID: 0x00, NetFn: NetFnBridgeRequest, Name: "Get Bridge State"}
	CommandSetBridgeState        = Command{ID: 0x01, NetFn: NetFnBridgeRequest, Name: "Set Bridge State"}
	CommandGetICMBAddress        = Command{ID: 0x02, NetFn: NetFnBridgeRequest, Name: "Get ICMB Address"}
	CommandSetICMBAddress        = Command{ID: 0x03, NetFn: NetFnBridgeRequest, Name: "Set ICMB Address"}
	CommandSetBridgeProxyAddress = Command{ID: 0x04, NetFn: NetFnBridgeRequest, Name: "Set Bridge ProxyAddress"}
	CommandGetBridgeStatistics   = Command{ID: 0x05, NetFn: NetFnBridgeRequest, Name: "Get Bridge Statistics"}
	CommandGetICMBCapabilities   = Command{ID: 0x06, NetFn: NetFnBridgeRequest, Name: "Get ICMB Capabilities"}
	CommandClearBridgeStatistics = Command{ID: 0x08, NetFn: NetFnBridgeRequest, Name: "Clear Bridge Statistics"}
	CommandGetBridgeProxyAddress = Command{ID: 0x09, NetFn: NetFnBridgeRequest, Name: "Get Bridge Proxy Address"}
	CommandGetICMBConnectorInfo  = Command{ID: 0x0a, NetFn: NetFnBridgeRequest, Name: "Get ICMB Connector Info"}
	CommandGetICMBConnectionID   = Command{ID: 0x0b, NetFn: NetFnBridgeRequest, Name: "Get ICMB Connection ID"}
	CommandSendICMBConnectionID  = Command{ID: 0x0c, NetFn: NetFnBridgeRequest, Name: "Send ICMB Connection ID"}

	// Discovery Commands (ICMB)
	CommandPrepareForDiscovery = Command{ID: 0x10, NetFn: NetFnBridgeRequest, Name: "PrepareForDiscovery"}
	CommandGetAddresses        = Command{ID: 0x11, NetFn: NetFnBridgeRequest, Name: "GetAddresses"}
	CommandSetDiscovered       = Command{ID: 0x12, NetFn: NetFnBridgeRequest, Name: "SetDiscovered"}
	CommandGetChassisDeviceId  = Command{ID: 0x13, NetFn: NetFnBridgeRequest, Name: "GetChassisDeviceId"}
	CommandSetChassisDeviceId  = Command{ID: 0x14, NetFn: NetFnBridgeRequest, Name: "SetChassisDeviceId"}

	// Bridging Commands (ICMB)
	CommandBridgeRequest = Command{ID: 0x20, NetFn: NetFnBridgeRequest, Name: "BridgeRequest"}
	CommandBridgeMessage = Command{ID: 0x21, NetFn: NetFnBridgeRequest, Name: "BridgeMessage"}

	// Event Commands (ICMB)
	CommandGetEventCount          = Command{ID: 0x30, NetFn: NetFnBridgeRequest, Name: "GetEventCount"}
	CommandSetEventDestination    = Command{ID: 0x31, NetFn: NetFnBridgeRequest, Name: "SetEventDestination"}
	CommandSetEventReceptionState = Command{ID: 0x32, NetFn: NetFnBridgeRequest, Name: "SetEventReceptionState"}
	CommandSendICMBEventMessage   = Command{ID: 0x33, NetFn: NetFnBridgeRequest, Name: "SendICMBEventMessage"}
	CommandGetEventDestination    = Command{ID: 0x34, NetFn: NetFnBridgeRequest, Name: "GetEventDestination (optional)"}
	CommandGetEventReceptionState = Command{ID: 0x35, NetFn: NetFnBridgeRequest, Name: "GetEventReceptionState (optional)"}

	// OEM Commands for Bridge NetFn
	// C0h-FEh

	// Other Bridge Commands
	CommandErrorReport = Command{ID: 0xff, NetFn: NetFnBridgeRequest, Name: "Error Report (optional)"}
)

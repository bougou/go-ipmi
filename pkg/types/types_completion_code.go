package types

import "fmt"

type CompletionCode uint8

// IPMI v2.0 Rev 1.1, section 5.2 Table 5-2, Completion Codes.
//
// IPMI v2.0 is backward-compatible with v1.5 for all completion codes.
// The only differences from v1.5 Rev 1.2 Table 5-2 are:
//   - D4h (modified): expanded from "Insufficient privilege level" to include
//     "or other security-based restriction (e.g. disabled for firmware firewall)"
//   - D6h (new): "Cannot execute command. Parameter is illegal because command
//     sub-function has been disabled or is unavailable"
//
// Completion Code ranges:
//   - 00h, C0h-FFh: Generic completion codes
//   - 01h-7Eh:      Device-specific (OEM) codes
//   - 80h-BEh:      Standard command-specific codes (per-command definitions)
//   - All other:     Reserved
const (
	// GENERIC COMPLETION CODES 00h, C0h-FFh

	CodeOK                                   CompletionCode = 0x00 // 00h: Command Completed Normally.
	CodeNormal                                              = CodeOK
	CodeNodeBusy                             CompletionCode = 0xC0 // C0h: Node Busy.
	CodeInvalidCommand                       CompletionCode = 0xC1 // C1h: Invalid Command.
	CodeInvalidCommandForLUN                 CompletionCode = 0xC2 // C2h: Command invalid for given LUN.
	CodeProcessTimeout                       CompletionCode = 0xC3 // C3h: Timeout while processing command.
	CodeOutOfSpace                           CompletionCode = 0xC4 // C4h: Out of space.
	CodeReservationCanceled                  CompletionCode = 0xC5 // C5h: Reservation Canceled or Invalid Reservation ID.
	CodeRequestDataTruncated                 CompletionCode = 0xC6 // C6h: Request data truncated.
	CodeRequestDataLengthInvalid             CompletionCode = 0xC7 // C7h: Request data length invalid.
	CodeRequestDataLengthLimitExceeded       CompletionCode = 0xC8 // C8h: Request data field length limit exceeded.
	CodeParameterOutOfRange                  CompletionCode = 0xC9 // C9h: Parameter out of range.
	CodeCannotReturnRequestedDataBytes       CompletionCode = 0xCA // CAh: Cannot return number of requested data bytes.
	CodeRequestedDataNotPresent              CompletionCode = 0xCB // CBh: Requested Sensor, data, or record not present.
	CodeRequestDataFieldInvalid              CompletionCode = 0xCC // CCh: Invalid data field in Request.
	CodeIllegalCommand                       CompletionCode = 0xCD // CDh: Command illegal for specified sensor or record type.
	CodeCannotProvideResponse                CompletionCode = 0xCE // CEh: Command response could not be provided.
	CodeCannotExecuteDuplicatedRequest       CompletionCode = 0xCF // CFh: Cannot execute duplicated request.
	CodeCannotProvideResponseSDRRInUpdate    CompletionCode = 0xD0 // D0h: SDR Repository in update mode.
	CodeCannotProvideResponseFirmwareUpdate  CompletionCode = 0xD1 // D1h: Device in firmware update mode.
	CodeCannotProvideResponseBMCInitialize   CompletionCode = 0xD2 // D2h: BMC initialization in progress.
	CodeDestinationUnavailable               CompletionCode = 0xD3 // D3h: Destination unavailable.
	CodeInsufficientPrivilege                CompletionCode = 0xD4 // D4h: Cannot execute command due to insufficient privilege level or other security-based restriction.
	CodeCannotExecuteCommandSecurityRestrict                = CodeInsufficientPrivilege
	CodeNotSupported                         CompletionCode = 0xD5 // D5h: Cannot execute command. Command, or request parameter(s), not supported in present state.
	CodeCannotExecuteCommandNotSupported                    = CodeNotSupported
	CodeSubFnDisabled                        CompletionCode = 0xD6 // D6h: Cannot execute command. Parameter is illegal because command sub-function has been disabled or is unavailable.
	CodeCannotExecuteCommandSubFnDisabled                   = CodeSubFnDisabled
	CodeUnspecifiedError                     CompletionCode = 0xFF // FFh: Unspecified error.

	// DEVICE-SPECIFIC (OEM) CODES 01h-7Eh — interpretation requires a-priori device knowledge.

	// COMMAND-SPECIFIC CODES 80h-BEh — defined per-command in the relevant command specification sections.
	//
	// CodeParameterNotSupported (80h) is defined for Set/Get System Boot Options
	// (v2.0§28.12 / §28.13) and other config-parameter commands (e.g. §22.14a/b
	// Set/Get System Info Parameters) when the parameter selector is not implemented.
	CodeParameterNotSupported CompletionCode = 0x80
)

var CC = map[uint8]string{
	0x00: "Command completed normally",
	0xc0: "Node busy",
	0xc1: "Invalid command",
	0xc2: "Invalid command on LUN",
	0xc3: "Timeout",
	0xc4: "Out of space",
	0xc5: "Reservation cancelled or invalid",
	0xc6: "Request data truncated",
	0xc7: "Request data length invalid",
	0xc8: "Request data field length limit exceeded",
	0xc9: "Parameter out of range",
	0xca: "Cannot return number of requested data bytes",
	0xcb: "Requested sensor, data, or record not found",
	0xcc: "Invalid data field in request",
	0xcd: "Command illegal for specified sensor or record type",
	0xce: "Command response could not be provided",
	0xcf: "Cannot execute duplicated request",
	0xd0: "SDR Repository in update mode",
	0xd1: "Device firmware in update mode",
	0xd2: "BMC initialization in progress",
	0xd3: "Destination unavailable",
	0xd4: "Cannot execute command, insufficient privilege level or other security-based restriction",
	0xd5: "Cannot execute command, command or request parameters not supported in present state",
	0xd6: "Cannot execute command, command disabled or is unavailable",
	0xff: "Unspecified error",
}

// String return description of generic completion code.
// Please use StrCC function to get description for any completion code
// returned for specific command response.
func (cc CompletionCode) String() string {
	if s, ok := CC[uint8(cc)]; ok {
		return s
	}
	return fmt.Sprintf("0x%02x", uint8(cc))
}

// commandSpecificCC records the command-specific completion codes (80h-BEh,
// v2.0§5.2 Table 5-2) defined in each command's own spec section. Generic
// codes live in CC; this table holds only the per-command additions, keyed by
// Command so that code holding just the dispatched command (server-side
// dispatch/middleware, without a decoded Response) can still name a code.
// Names follow the spec's wording with the first letter capitalized, matching
// the style of CC.

// Shared command-specific sets. The parameter-configuration commands repeat
// the same 80h-83h pattern across their spec sections; the Set commands split
// into those without write-only parameters (System Info §22.16a, Boot Options
// §28.12) and those with them (LAN §23.2, Serial/Modem §25.2, SOL §26.3, PEF
// §30.4); every Get counterpart (§22.16b, §23.3, §25.3, §26.4, §28.13, §30.5)
// defines 80h only.
var (
	paramConfigSetCC = map[uint8]string{
		0x80: "Parameter not supported",
		0x81: "Attempt to set 'set in progress' value (in parameter #0) when not in 'set complete' state",
		0x82: "Attempt to write read-only parameter",
	}
	paramConfigSetRWCC = map[uint8]string{
		0x80: "Parameter not supported",
		0x81: "Attempt to set 'set in progress' value (in parameter #0) when not in 'set complete' state",
		0x82: "Attempt to write read-only parameter",
		0x83: "Attempt to read write-only parameter",
	}
	paramConfigGetCC = map[uint8]string{
		0x80: "Parameter not supported",
	}
	// selEraseCC is returned by the SEL Device commands that cannot run while
	// a Clear SEL is in progress (Tables 31-2, 31-4, 31-5, 31-7) and by
	// Set/Get Last Processed Event ID (Tables 30-7, 30-8).
	selEraseCC = map[uint8]string{
		0x81: "Cannot execute command, SEL erase in progress",
	}
	// fruBusyCC is returned by Read/Write FRU Data (Tables 34-3, 34-4) when the
	// logical FRU device is temporarily unavailable; software may retry.
	fruBusyCC = map[uint8]string{
		0x81: "FRU device busy",
	}
	// queueEmptyCC is returned by Get Message (Table 22-7) and Read Event
	// Message Buffer (Table 22-11) when no message is waiting.
	queueEmptyCC = map[uint8]string{
		0x80: "Data not available (queue / buffer empty)",
	}
)

var commandSpecificCC = map[Command]map[uint8]string{
	// §21 BMC Device and LUN commands.
	CommandSetCommandEnables: {
		0x80: "Attempt to enable an unsupported or un-configurable command",
	},
	CommandSetCommandSubfunctionEnables: {
		// Spec Table 21-9, "Set Configurable Command Sub-function Enables".
		0x80: "Attempt to enable an unsupported or un-configurable sub-function",
	},
	CommandGetConfigurableCommandSubfunctions: {
		// Table 21-10 lists the Set wording verbatim; a Get can only hit this
		// for an unsupported or un-configurable command number.
		0x80: "Attempt to enable an unsupported or un-configurable sub-function",
	},

	// §22 messaging and session commands.
	CommandGetMessage:             queueEmptyCC,
	CommandReadEventMessageBuffer: queueEmptyCC,
	CommandSendMessage: {
		0x80: "Invalid Session Handle",
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
	},
	CommandMasterWriteRead: {
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
		0x84: "Truncated Read",
	},
	CommandSetSystemInfoParam: paramConfigSetCC,
	CommandGetSystemInfoParam: paramConfigGetCC,
	CommandGetSessionChallenge: {
		0x81: "Invalid user name",
		0x82: "Null user name (User 1) not enabled",
	},
	CommandActivateSession: {
		0x81: "No session slot available (BMC cannot accept any more sessions)",
		0x82: "No slot available for given user",
		0x83: "No slot available to support user due to maximum privilege capability",
		0x84: "Session sequence number out-of-range",
		0x85: "Invalid Session ID in request",
		0x86: "Requested maximum privilege level exceeds user and/or channel privilege limit",
	},
	CommandSetSessionPrivilegeLevel: {
		0x80: "Requested level not available for this user",
		0x81: "Requested level exceeds Channel and/or User Privilege Limit",
		0x82: "Cannot disable User Level authentication",
	},
	CommandCloseSession: {
		0x87: "Invalid Session ID in request",
		0x88: "Invalid Session Handle in request",
	},
	CommandSetChannelAccess: {
		0x82: "Set not supported on selected channel",
		0x83: "Access mode not supported",
	},
	CommandGetChannelAccess: {
		0x82: "Command not supported for selected channel",
	},
	CommandSetChannelSecurityKeys: {
		0x80: "Cannot perform set / confirm. Key is locked",
		0x81: "Insufficient key bytes",
		0x82: "Too many key bytes",
		0x83: "Key value does not meet criteria for specified type of key",
		0x84: "KR is not used",
	},
	CommandSetUserPassword: {
		0x80: "Password test failed. Password size correct, but password data does not match stored value",
		0x81: "Password test failed. Wrong password size was used",
	},

	// §23 LAN commands.
	CommandSetLanConfigParam: paramConfigSetRWCC,
	CommandGetLanConfigParam: paramConfigGetCC,

	// §24 payload commands.
	CommandActivatePayload: {
		0x80: "Payload already active on another session",
		0x81: "Payload type is disabled",
		0x82: "Payload activation limit reached",
		0x83: "Cannot activate payload with encryption",
		0x84: "Cannot activate payload without encryption",
	},
	CommandDeactivatePayload: {
		0x80: "Payload already deactivated",
		0x81: "Payload type is disabled",
	},
	CommandSuspendResumePayloadEncryption: {
		0x80: "Operation not supported for given payload type",
		0x81: "Operation not allowed under present configuration for given payload type",
		0x82: "Encryption is not available for session that payload type is active under",
		0x83: "The payload instance is not presently active",
	},
	CommandGetChannelPayloadVersion: {
		0x80: "Payload type not available on given channel",
	},
	CommandGetChannelOEMPayloadInfo: {
		0x80: "OEM Payload IANA and/or Payload ID not supported",
	},

	// §25 serial/modem and PPP commands.
	CommandSetSerialConfig: paramConfigSetRWCC,
	CommandGetSerialConfig: paramConfigGetCC,
	CommandSendPPPPacket: {
		0x80: "PPP Link is not up",
		0x81: "IP Protocol is not up",
	},
	CommandGetPPPReceiveData: {
		0x80: "No packet data available",
	},
	CommandCallback: {
		0x81: "Callback rejected due to alert in progress on this channel",
		0x82: "Callback rejected due to IPMI messaging session active on the callback channel",
	},

	// §26 SOL commands.
	CommandSetSOLConfigParam: paramConfigSetRWCC,
	CommandGetSOLConfigParam: paramConfigGetCC,

	// §27 watchdog commands.
	CommandResetWatchdogTimer: {
		0x80: "Attempt to start un-initialized watchdog",
	},

	// §28 chassis commands.
	CommandSetSystemBootOptions: paramConfigSetCC,
	CommandGetSystemBootOptions: paramConfigGetCC,

	// §30 PEF and alerting commands.
	CommandSetPEFConfigParam:       paramConfigSetRWCC,
	CommandGetPEFConfigParam:       paramConfigGetCC,
	CommandSetLastProcessedEventId: selEraseCC,
	CommandGetLastProcessedEventId: selEraseCC,
	CommandAlertImmediate: {
		0x81: "Alert Immediate rejected due to alert already in progress",
		0x82: "Alert Immediate rejected due to IPMI messaging session active on this channel",
		0x83: "Platform Event Parameters (4:11) not supported",
	},

	// §31 SEL Device commands. Add/Delete additionally reject Record Types
	// they cannot store/remove with 80h (Tables 31-6, 31-8).
	CommandGetSELInfo:         selEraseCC,
	CommandReserveSEL:         selEraseCC,
	CommandGetSELEntry:        selEraseCC,
	CommandPartialAddSELEntry: selEraseCC,
	CommandAddSELEntry: {
		0x80: "Operation not supported for this Record Type",
		0x81: "Cannot execute command, SEL erase in progress",
	},
	CommandDeleteSELEntry: {
		0x80: "Operation not supported for this Record Type",
		0x81: "Cannot execute command, SEL erase in progress",
	},

	// §33 SDR Repository commands.
	CommandPartialAddSDR: {
		0x80: "Record rejected due to mismatch between record length in header data and number of bytes written",
	},

	// §34 FRU commands.
	CommandReadFRUData:  fruBusyCC,
	CommandWriteFRUData: fruBusyCC,

	// §35 sensor commands.
	CommandGetDeviceSDR: {
		0x80: "Record changed",
	},
	CommandSetSensorReadingAndEventStatus: {
		0x80: "Attempt to change reading or set or clear status bits that are not settable via this command",
		0x81: "Attempted to set Event Data Bytes, but setting Event Data Bytes is not supported for this sensor",
	},

	// §35b ICMB bridging commands.
	CommandForwarded: {
		0x80: "Target controller unavailable",
	},

	// DCMI commands (dcmi-spec-v1.5 §6). DCMI §8 Table 8-1 only mirrors the
	// IPMI generic codes; the command-specific ones are called out in the
	// command tables themselves. Set/Get DCMI Configuration Parameters follow
	// the parameter-configuration pattern (Tables 6-4/6-5).
	CommandGetDCMIPowerLimit: {
		0x80: "No Active Set Power Limit",
	},
	CommandSetDCMIPowerLimit: {
		0x84: "Power Limit out of range",
		0x85: "Correction Time out of range",
		0x89: "Statistics Reporting Period out of range",
	},
	CommandSetDCMIThermalLimit: {
		0x84: "Thermal Limit out of range",
		0x85: "Exception Time out of range",
	},
	CommandSetDCMIConfigParam: paramConfigSetCC,
	CommandGetDCMIConfigParam: paramConfigGetCC,
}

// StrCCForCommand returns the description of ccode for the specified command:
// the command-specific name when the command defines one for ccode, otherwise
// the generic name, otherwise the hex string. Unlike StrCC it needs no decoded
// Response, so it works wherever only the Command is available.
func StrCCForCommand(cmd Command, ccode uint8) string {
	if s, ok := commandSpecificCC[cmd][ccode]; ok {
		return s
	}
	return CompletionCode(ccode).String()
}

// Error makes CompletionCode implement the error interface.
func (cc CompletionCode) Error() string {
	return fmt.Sprintf("IPMI completion code %s", cc.String())
}

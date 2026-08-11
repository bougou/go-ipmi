package types

import (
	"fmt"
	"maps"
)

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

	// Activate Payload command completion codes (v2.0§24.1, Table 24-2).
	CodeActivatePayloadAlreadyActive                   CompletionCode = 0x80
	CodeActivatePayloadTypeDisabled                    CompletionCode = 0x81
	CodeActivatePayloadActivationLimitReached          CompletionCode = 0x82
	CodeActivatePayloadCannotActivateWithEncryption    CompletionCode = 0x83
	CodeActivatePayloadCannotActivateWithoutEncryption CompletionCode = 0x84

	// Deactivate Payload command completion codes (v2.0§24.2, Table 24-3).
	CodeDeactivatePayloadAlreadyDeactivated CompletionCode = 0x80
	CodeDeactivatePayloadTypeDisabled       CompletionCode = 0x81
)

// String return description of generic completion code.
// Prefer [StrCC] when the command is known so that command-specific codes
// (80h-BEh) resolve to names.
func (cc CompletionCode) String() string {
	if s, ok := genericCC[uint8(cc)]; ok {
		return s
	}
	return fmt.Sprintf("0x%02x", uint8(cc))
}

// Error makes CompletionCode implement the error interface.
func (cc CompletionCode) Error() string {
	return fmt.Sprintf("IPMI completion code %s", cc.String())
}

// CommandSpecificCC returns the command-specific completion-code name map for
// cmd. Returns nil when the command defines none. The returned map is shared
// package state; callers must treat it as read-only.
func CommandSpecificCC(cmd Command) map[uint8]string {
	return commandSpecificCC[cmd.Key()]
}

// AllCC returns a new map of generic completion codes merged with any
// command-specific codes defined for cmd. Prefer [StrCC] when only one
// code needs a name.
func AllCC(cmd Command) map[uint8]string {
	out := map[uint8]string{}
	maps.Copy(out, genericCC)
	maps.Copy(out, CommandSpecificCC(cmd))
	return out
}

// StrCC returns the description of ccode for cmd: command-specific name when
// defined, otherwise the generic name, otherwise hex.
func StrCC(cmd Command, ccode uint8) string {
	if m := CommandSpecificCC(cmd); m != nil {
		if s, ok := m[ccode]; ok {
			return s
		}
	}
	return CompletionCode(ccode).String()
}

var genericCC = map[uint8]string{
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

// commandSpecificCC is the single source of truth for command-specific
// completion codes (80h-BEh, v2.0§5.2 Table 5-2) defined in each command's
// own spec section. Generic codes live in genericCC. Keys are [CommandKey] so a
// Name-less {NetFn, ID} from the wire matches the same entry as a named
// [Command] table constant.
//
// Shared sets: parameter-configuration commands repeat the same 80h-83h
// pattern; Set commands without write-only parameters (System Info §22.14a,
// Boot Options §28.12) omit 83h; Set commands with them (LAN §23.2,
// Serial/Modem §25.2, SOL §26.3, PEF §30.4) include 83h; every Get
// counterpart defines 80h only.
var (
	paramConfigSetCC = map[uint8]string{
		0x80: "Parameter not supported",
		0x81: "Attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state",
		0x82: "Attempt to write read-only parameter",
	}
	paramConfigSetRWCC = map[uint8]string{
		0x80: "Parameter not supported",
		0x81: "Attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state",
		0x82: "Attempt to write read-only parameter",
		0x83: "Attempt to read write-only parameter",
	}
	paramConfigGetCC = map[uint8]string{
		0x80: "Parameter not supported",
	}
	// selEraseCC: SEL Device commands that cannot run while Clear SEL is in
	// progress (Tables 31-2, 31-4, 31-5) and Set/Get Last Processed Event ID
	// (Tables 30-7, 30-8).
	selEraseCC = map[uint8]string{
		0x81: "Cannot execute command, SEL erase in progress",
	}
	// queueEmptyCC: Get Message (Table 22-7) and Read Event Message Buffer
	// (Table 22-11).
	queueEmptyCC = map[uint8]string{
		0x80: "Data not available (queue / buffer empty)",
	}
	// recordMismatchCC: Partial Add SDR (Table 33-8) and Partial Add SEL Entry
	// (Table 31-7) 80h.
	recordMismatchCC = map[uint8]string{
		0x80: "Record rejected due to mismatch between record length in header data and number of bytes written",
	}
)

var commandSpecificCC = map[CommandKey]map[uint8]string{
	// v2.0§21 BMC Device and Messaging Firewall commands.
	CommandSetCommandEnables.Key(): {
		0x80: "Attempt to enable an unsupported or un-configurable command", // Table 21-7
	},
	CommandSetCommandSubfunctionEnables.Key(): {
		0x80: "Attempt to enable an unsupported or un-configurable sub-function", // Table 21-9
	},
	CommandGetCommandSubfunctionEnables.Key(): {
		0x80: "Attempt to get an unsupported or un-configurable sub-function", // Table 21-10
	},

	// v2.0§22 messaging and session commands.
	CommandGetMessage.Key():             queueEmptyCC, // Table 22-7
	CommandReadEventMessageBuffer.Key(): queueEmptyCC, // Table 22-11
	CommandSendMessage.Key(): { // Table 22-9
		0x80: "Invalid Session Handle. The session handle does not match up with any currently active sessions for this channel",
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
	},
	CommandMasterWriteRead.Key(): { // Table 22-14
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
		0x84: "Truncated Read",
	},
	CommandSetSystemInfoParam.Key(): paramConfigSetCC, // §22.14a
	CommandGetSystemInfoParam.Key(): paramConfigGetCC, // §22.14b
	CommandGetSessionChallenge.Key(): { // Table 22-21
		0x81: "Invalid user name",
		0x82: "Null user name (User 1) not enabled",
	},
	CommandActivateSession.Key(): { // Table 22-22
		0x81: "No session slot available (BMC cannot accept any more sessions)",
		0x82: "No slot available for given user. (Limit of user sessions allowed under that name has been reached)",
		0x83: "No slot available to support user due to maximum privilege capability",
		0x84: "Session sequence number out-of-range",
		0x85: "Invalid Session ID in request",
		0x86: "Requested maximum privilege level exceeds user and/or channel privilege limit",
	},
	CommandSetSessionPrivilegeLevel.Key(): { // Table 22-24
		0x80: "Requested level not available for this user",
		0x81: "Requested level exceeds Channel and/or User Privilege Limit",
		0x82: "Cannot disable User Level authentication",
	},
	CommandCloseSession.Key(): { // Table 22-25
		0x87: "Invalid Session ID in request",
		0x88: "Invalid Session Handle in request",
	},
	CommandSetChannelAccess.Key(): { // Table 22-27
		0x82: "Set not supported on selected channel (e.g. channel is session-less)",
		0x83: "Access mode not supported",
	},
	CommandGetChannelAccess.Key(): { // Table 22-28
		0x82: "Command not supported for selected channel (e.g. channel is session-less)",
	},
	CommandSetChannelSecurityKeys.Key(): { // Table 22-30
		0x80: "Cannot perform set / confirm. Key is locked",
		0x81: "Insufficient key bytes",
		0x82: "Too many key bytes",
		0x83: "Key value does not meet criteria for specified type of key",
		0x84: "KR is not used",
	},
	CommandSetUserPassword.Key(): { // Table 22-35
		0x80: "Password test failed. Password size correct, but password data does not match stored value",
		0x81: "Password test failed. Wrong password size was used",
	},

	// v2.0§23 LAN commands.
	CommandSetLanConfigParam.Key(): paramConfigSetRWCC, // Table 23-2
	CommandGetLanConfigParam.Key(): paramConfigGetCC,   // Table 23-3
	// Suspend BMC ARPs (Table 23-5) defines no command-specific codes.

	// v2.0§24 payload commands.
	CommandActivatePayload.Key(): { // Table 24-2
		uint8(CodeActivatePayloadAlreadyActive):                   "Payload already active on another session",
		uint8(CodeActivatePayloadTypeDisabled):                    "Payload type is disabled",
		uint8(CodeActivatePayloadActivationLimitReached):          "Payload activation limit reached",
		uint8(CodeActivatePayloadCannotActivateWithEncryption):    "Cannot activate payload with encryption",
		uint8(CodeActivatePayloadCannotActivateWithoutEncryption): "Cannot activate payload without encryption",
	},
	CommandDeactivatePayload.Key(): { // Table 24-3
		uint8(CodeDeactivatePayloadAlreadyDeactivated): "Payload already deactivated",
		uint8(CodeDeactivatePayloadTypeDisabled):       "Payload type is disabled",
	},
	CommandSuspendResumePayloadEncryption.Key(): { // Table 24-5
		0x80: "Operation not supported for given payload type",
		0x81: "Operation not allowed under present configuration for given payload type",
		0x82: "Encryption is not available for session that payload type is active under",
		0x83: "The payload instance is not presently active",
	},
	CommandGetChannelPayloadVersion.Key(): { // Table 24-11
		0x80: "Payload type not available on given channel",
	},
	CommandGetChannelOEMPayloadInfo.Key(): { // Table 24-12
		0x80: "OEM Payload IANA and/or Payload ID not supported",
	},

	// v2.0§25 serial/modem and PPP commands.
	CommandSetSerialConfig.Key(): paramConfigSetRWCC, // Table 25-2
	CommandGetSerialConfig.Key(): paramConfigGetCC,   // Table 25-3
	CommandSendPPPPacket.Key(): { // Table 25-9
		0x80: "PPP Link is not up",
		0x81: "IP Protocol is not up",
	},
	CommandGetPPPReceiveData.Key(): { // Table 25-10
		0x80: "No packet data available",
	},
	CommandCallback.Key(): { // Table 25-12
		0x81: "Callback rejected due to alert in progress on this channel",
		0x82: "Callback rejected due to IPMI messaging session active on the callback channel",
	},

	// v2.0§26 SOL commands.
	CommandSetSOLConfigParam.Key(): paramConfigSetRWCC, // Table 26-3
	CommandGetSOLConfigParam.Key(): paramConfigGetCC,   // Table 26-4

	// v2.0§27 watchdog commands.
	CommandResetWatchdogTimer.Key(): { // Table 27-2
		0x80: "Attempt to start un-initialized watchdog",
	},

	// v2.0§28 chassis commands.
	CommandSetSystemBootOptions.Key(): paramConfigSetCC, // Table 28-12
	CommandGetSystemBootOptions.Key(): paramConfigGetCC, // Table 28-13

	// v2.0§30 PEF and alerting commands.
	CommandSetPEFConfigParam.Key():       paramConfigSetRWCC, // Table 30-4
	CommandGetPEFConfigParam.Key():       paramConfigGetCC,   // Table 30-5
	CommandSetLastProcessedEventId.Key(): selEraseCC,         // Table 30-7
	CommandGetLastProcessedEventId.Key(): selEraseCC,         // Table 30-8
	CommandAlertImmediate.Key(): { // Table 30-9
		0x81: "Alert Immediate rejected due to alert already in progress",
		0x82: "Alert Immediate rejected due to IPMI messaging session active on this channel",
		0x83: "Platform Event Parameters (4:11) not supported",
	},
	// PET Acknowledge (Table 30-10) defines no command-specific codes.

	// v2.0§31 SEL Device commands.
	CommandGetSELInfo.Key():  selEraseCC, // Table 31-2
	CommandReserveSEL.Key():  selEraseCC, // Table 31-4
	CommandGetSELEntry.Key(): selEraseCC, // Table 31-5
	CommandAddSELEntry.Key(): { // Table 31-6
		0x80: "Operation not supported for this Record Type",
		0x81: "Cannot execute command, SEL erase in progress",
	},
	CommandPartialAddSELEntry.Key(): { // Table 31-7
		0x80: "Record rejected due to mismatch between record length in header data and number of bytes written",
		0x81: "Cannot execute command, SEL erase in progress",
	},
	CommandDeleteSELEntry.Key(): { // Table 31-8
		0x80: "Operation not supported for this Record Type",
		0x81: "Cannot execute command, SEL erase in progress",
	},

	// v2.0§33 SDR Repository commands.
	CommandPartialAddSDR.Key(): recordMismatchCC, // Table 33-8

	// v2.0§34 FRU commands.
	CommandReadFRUData.Key(): { // Table 34-3
		0x81: "FRU device busy",
	},
	CommandWriteFRUData.Key(): { // Table 34-4
		0x80: "Write-protected offset",
		0x81: "FRU device busy",
	},

	// v2.0§35 sensor commands.
	CommandGetDeviceSDR.Key(): { // Table 35-3
		0x80: "Record changed",
	},
	CommandSetSensorReadingAndEventStatus.Key(): { // Table 35-xx Set Sensor Reading and Event Status
		0x80: "Attempt to change reading or set or clear status bits that are not settable via this command",
		0x81: "Attempted to set Event Data Bytes, but setting Event Data Bytes is not supported for this sensor",
	},

	// v2.0§35b ICMB bridging.
	CommandForwarded.Key(): { // Table 35b-5
		0x80: "Target controller unavailable",
	},

	// DCMI (dcmi-spec-v1.5 §6). Only codes listed in the command tables; §8
	// Table 8-1 mirrors IPMI generic codes only. Set/Get DCMI Configuration
	// Parameters (Tables 6-4/6-5) and Get Temperature Readings (Table 6-22)
	// define no command-specific codes.
	CommandGetDCMIAssetTag.Key(): { // dcmi§6.4.2 Table 6-8
		0x80: "Encoding type in FRU is binary / unspecified",
		0x81: "Encoding type in FRU is BCD Plus",
		0x82: "Encoding type in FRU is 6-bit ASCII Packed",
		0x83: "Encoding type in FRU is set to ASCII+Latin1, but language code is not set to English (indicating data is 2-byte UNICODE)",
	},
	CommandGetDCMIPowerLimit.Key(): { // dcmi§6.6.2 Table 6-17
		0x80: "No Active Set Power Limit",
	},
	CommandSetDCMIPowerLimit.Key(): { // dcmi§6.6.3 Table 6-18
		0x84: "Power Limit out of range",
		0x85: "Correction Time out of range",
		0x89: "Statistics Reporting Period out of range",
	},
	CommandSetDCMIThermalLimit.Key(): { // dcmi§6.7.2 Table 6-21
		0x84: "Thermal Limit out of range",
		0x85: "Exception Time out of range",
	},
}

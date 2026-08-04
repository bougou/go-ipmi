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
var commandSpecificCC = map[Command]map[uint8]string{
	// v2.0§28.12 Table 28-12 / §28.13 Table 28-13.
	CommandSetSystemBootOptions: {
		0x80: "Parameter not supported",
		0x81: "Attempt to set 'set in progress' when not in 'set complete' state",
		0x82: "Attempt to write read-only parameter",
	},
	CommandGetSystemBootOptions: {
		0x80: "Parameter not supported",
	},
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

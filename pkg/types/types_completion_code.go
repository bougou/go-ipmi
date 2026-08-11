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

	CodeOK                                  CompletionCode = 0x00
	CodeNodeBusy                            CompletionCode = 0xC0
	CodeInvalidCommand                      CompletionCode = 0xC1
	CodeInvalidCommandForLUN                CompletionCode = 0xC2
	CodeProcessTimeout                      CompletionCode = 0xC3
	CodeOutOfSpace                          CompletionCode = 0xC4
	CodeReservationCanceled                 CompletionCode = 0xC5
	CodeRequestDataTruncated                CompletionCode = 0xC6
	CodeRequestDataLengthInvalid            CompletionCode = 0xC7
	CodeRequestDataLengthLimitExceeded      CompletionCode = 0xC8
	CodeParameterOutOfRange                 CompletionCode = 0xC9
	CodeCannotReturnRequestedDataBytes      CompletionCode = 0xCA
	CodeRequestedDataNotPresent             CompletionCode = 0xCB
	CodeRequestDataFieldInvalid             CompletionCode = 0xCC
	CodeIllegalCommand                      CompletionCode = 0xCD
	CodeCannotProvideResponse               CompletionCode = 0xCE
	CodeCannotExecuteDuplicatedRequest      CompletionCode = 0xCF
	CodeCannotProvideResponseSDRRInUpdate   CompletionCode = 0xD0
	CodeCannotProvideResponseFirmwareUpdate CompletionCode = 0xD1
	CodeCannotProvideResponseBMCInitialize  CompletionCode = 0xD2
	CodeDestinationUnavailable              CompletionCode = 0xD3
	CodeInsufficientPrivilege               CompletionCode = 0xD4
	CodeNotSupported                        CompletionCode = 0xD5
	CodeSubFnDisabled                       CompletionCode = 0xD6
	CodeUnspecifiedError                    CompletionCode = 0xFF

	// DEVICE-SPECIFIC (OEM) CODES 01h-7Eh — interpretation requires a-priori device knowledge.

	// COMMAND-SPECIFIC CODES 80h-BEh — defined per-command in each command's own
	// spec section. The same numeric value has different meanings per command
	// (e.g. 0x80 repeats throughout), so names encode the command:
	// Code{Command}{Meaning}. To classify an error, compare
	// [ResponseError.CompletionCode] against these names together with the
	// command that produced the error, since a bare value is ambiguous.
	//
	// The human-readable description of each code lives in [commandSpecificCC]
	// (and [genericCC] for generic codes); [StrCC] resolves a (command, code)
	// pair to that description.

	// Shared command-specific codes: reused across multiple commands that define
	// identical meanings for the same value. Referenced by the shared sub-maps
	// below ([paramConfigSetCC], [paramConfigSetRWCC], [paramConfigGetCC],
	// [selEraseCC], [queueEmptyCC], [recordMismatchCC]).
	CodeParameterNotSupported            CompletionCode = 0x80 // Parameter-configuration commands (§22.14, §23, §25, §26, §28, §30): parameter selector not implemented.
	CodeParamConfigSetInProgressConflict CompletionCode = 0x81 // Parameter Set commands: 'set in progress' conflict in parameter #0.
	CodeParamConfigSetReadOnly           CompletionCode = 0x82 // Parameter Set commands: attempt to write a read-only parameter.
	CodeParamConfigReadWriteOnly         CompletionCode = 0x83 // Parameter Set commands with write-only params: attempt to read a write-only parameter.
	CodeSELEraseInProgress               CompletionCode = 0x81 // SEL Device commands and Set/Get Last Processed Event ID: SEL erase in progress.
	CodeDataNotAvailable                 CompletionCode = 0x80 // Get Message / Read Event Message Buffer: queue or buffer empty.
	CodePartialAddRecordMismatch         CompletionCode = 0x80 // Partial Add SDR / Partial Add SEL Entry: record length mismatch.

	// v2.0§21 BMC Device and Messaging Firewall commands.
	CodeSetCommandEnablesUnsupported            CompletionCode = 0x80
	CodeSetCommandSubfunctionEnablesUnsupported CompletionCode = 0x80
	CodeGetCommandSubfunctionEnablesUnsupported CompletionCode = 0x80

	// v2.0§22 messaging and session commands.
	CodeSendMessageInvalidSessionHandle           CompletionCode = 0x80
	CodeSendMessageLostArbitration                CompletionCode = 0x81
	CodeSendMessageBusError                       CompletionCode = 0x82
	CodeSendMessageNAKOnWrite                     CompletionCode = 0x83
	CodeMasterWriteReadLostArbitration            CompletionCode = 0x81
	CodeMasterWriteReadBusError                   CompletionCode = 0x82
	CodeMasterWriteReadNAKOnWrite                 CompletionCode = 0x83
	CodeMasterWriteReadTruncatedRead              CompletionCode = 0x84
	CodeGetSessionChallengeInvalidUserName        CompletionCode = 0x81
	CodeGetSessionChallengeNullUserNotEnabled     CompletionCode = 0x82
	CodeActivateSessionNoSlotAvailable            CompletionCode = 0x81
	CodeActivateSessionNoSlotForUser              CompletionCode = 0x82
	CodeActivateSessionNoSlotForPrivilege         CompletionCode = 0x83
	CodeActivateSessionSequenceOutOfRange         CompletionCode = 0x84
	CodeActivateSessionInvalidSessionID           CompletionCode = 0x85
	CodeActivateSessionPrivilegeExceedsLimit      CompletionCode = 0x86
	CodeSetSessionPrivilegeLevelLevelNotAvailable CompletionCode = 0x80
	CodeSetSessionPrivilegeLevelExceedsLimit      CompletionCode = 0x81
	CodeSetSessionPrivilegeLevelCannotDisableAuth CompletionCode = 0x82
	CodeCloseSessionInvalidSessionID              CompletionCode = 0x87
	CodeCloseSessionInvalidSessionHandle          CompletionCode = 0x88
	CodeSetChannelAccessSetNotSupported           CompletionCode = 0x82
	CodeSetChannelAccessModeNotSupported          CompletionCode = 0x83
	CodeGetChannelAccessNotSupported              CompletionCode = 0x82
	CodeSetChannelSecurityKeysKeyLocked           CompletionCode = 0x80
	CodeSetChannelSecurityKeysInsufficientBytes   CompletionCode = 0x81
	CodeSetChannelSecurityKeysTooManyBytes        CompletionCode = 0x82
	CodeSetChannelSecurityKeysKeyValueInvalid     CompletionCode = 0x83
	CodeSetChannelSecurityKeysKRNotUsed           CompletionCode = 0x84
	CodeSetUserPasswordDataMismatch               CompletionCode = 0x80
	CodeSetUserPasswordWrongSize                  CompletionCode = 0x81

	// v2.0§24 payload commands.
	CodeActivatePayloadAlreadyActive                   CompletionCode = 0x80
	CodeActivatePayloadTypeDisabled                    CompletionCode = 0x81
	CodeActivatePayloadActivationLimitReached          CompletionCode = 0x82
	CodeActivatePayloadCannotActivateWithEncryption    CompletionCode = 0x83
	CodeActivatePayloadCannotActivateWithoutEncryption CompletionCode = 0x84
	CodeDeactivatePayloadAlreadyDeactivated            CompletionCode = 0x80
	CodeDeactivatePayloadTypeDisabled                  CompletionCode = 0x81
	CodeSuspendResumePayloadEncryptionNotSupported     CompletionCode = 0x80
	CodeSuspendResumePayloadEncryptionNotAllowed       CompletionCode = 0x81
	CodeSuspendResumePayloadEncryptionNotAvailable     CompletionCode = 0x82
	CodeSuspendResumePayloadEncryptionNotActive        CompletionCode = 0x83
	CodeGetChannelPayloadVersionNotAvailable           CompletionCode = 0x80
	CodeGetChannelOEMPayloadInfoNotSupported           CompletionCode = 0x80

	// v2.0§25 serial/modem and PPP commands.
	CodeSendPPPPacketLinkNotUp     CompletionCode = 0x80
	CodeSendPPPPacketProtocolNotUp CompletionCode = 0x81
	CodeGetPPPReceiveDataNoData    CompletionCode = 0x80
	CodeCallbackAlertInProgress    CompletionCode = 0x81
	CodeCallbackSessionActive      CompletionCode = 0x82

	// v2.0§27 watchdog commands.
	CodeResetWatchdogTimerUninitialized CompletionCode = 0x80

	// v2.0§30 PEF and alerting commands.
	CodeAlertImmediateAlreadyInProgress  CompletionCode = 0x81
	CodeAlertImmediateSessionActive      CompletionCode = 0x82
	CodeAlertImmediateParamsNotSupported CompletionCode = 0x83

	// v2.0§31 SEL Device commands.
	CodeAddSELEntryRecordTypeNotSupported    CompletionCode = 0x80
	CodeDeleteSELEntryRecordTypeNotSupported CompletionCode = 0x80

	// v2.0§34 FRU commands.
	CodeReadFRUDataDeviceBusy      CompletionCode = 0x81
	CodeWriteFRUDataWriteProtected CompletionCode = 0x80
	CodeWriteFRUDataDeviceBusy     CompletionCode = 0x81

	// v2.0§35 sensor commands.
	CodeGetDeviceSDRRecordChanged             CompletionCode = 0x80
	CodeSetSensorReadingNotSettable           CompletionCode = 0x80
	CodeSetSensorReadingEventDataNotSupported CompletionCode = 0x81

	// v2.0§35b ICMB bridging.
	CodeForwardedTargetUnavailable CompletionCode = 0x80

	// DCMI (dcmi-spec-v1.5 §6).
	CodeGetDCMIAssetTagBinaryEncoding              CompletionCode = 0x80
	CodeGetDCMIAssetTagBCDPlus                     CompletionCode = 0x81
	CodeGetDCMIAssetTag6BitASCII                   CompletionCode = 0x82
	CodeGetDCMIAssetTagUnicodeNonEnglish           CompletionCode = 0x83
	CodeGetDCMIPowerLimitNoActiveLimit             CompletionCode = 0x80
	CodeSetDCMIPowerLimitOutOfRange                CompletionCode = 0x84
	CodeSetDCMIPowerLimitCorrectionTimeOutOfRange  CompletionCode = 0x85
	CodeSetDCMIPowerLimitStatsPeriodOutOfRange     CompletionCode = 0x89
	CodeSetDCMIThermalLimitOutOfRange              CompletionCode = 0x84
	CodeSetDCMIThermalLimitExceptionTimeOutOfRange CompletionCode = 0x85
)

// String return description of generic completion code.
// Prefer [StrCC] when the command is known so that command-specific codes
// (80h-BEh) resolve to names.
func (cc CompletionCode) String() string {
	if s, ok := genericCC[cc]; ok {
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
func CommandSpecificCC(cmd Command) map[CompletionCode]string {
	return commandSpecificCC[cmd.Key()]
}

// AllCC returns a new map of generic completion codes merged with any
// command-specific codes defined for cmd. Prefer [StrCC] when only one
// code needs a name.
func AllCC(cmd Command) map[CompletionCode]string {
	out := map[CompletionCode]string{}
	maps.Copy(out, genericCC)
	maps.Copy(out, CommandSpecificCC(cmd))
	return out
}

// StrCC returns the description of ccode for cmd: command-specific name when
// defined, otherwise the generic name, otherwise hex.
func StrCC(cmd Command, ccode uint8) string {
	if m := CommandSpecificCC(cmd); m != nil {
		if s, ok := m[CompletionCode(ccode)]; ok {
			return s
		}
	}
	return CompletionCode(ccode).String()
}

var genericCC = map[CompletionCode]string{
	CodeOK:                                  "Command completed normally",
	CodeNodeBusy:                            "Node busy",
	CodeInvalidCommand:                      "Invalid command",
	CodeInvalidCommandForLUN:                "Invalid command on LUN",
	CodeProcessTimeout:                      "Timeout",
	CodeOutOfSpace:                          "Out of space",
	CodeReservationCanceled:                 "Reservation cancelled or invalid",
	CodeRequestDataTruncated:                "Request data truncated",
	CodeRequestDataLengthInvalid:            "Request data length invalid",
	CodeRequestDataLengthLimitExceeded:      "Request data field length limit exceeded",
	CodeParameterOutOfRange:                 "Parameter out of range",
	CodeCannotReturnRequestedDataBytes:      "Cannot return number of requested data bytes",
	CodeRequestedDataNotPresent:             "Requested sensor, data, or record not found",
	CodeRequestDataFieldInvalid:             "Invalid data field in request",
	CodeIllegalCommand:                      "Command illegal for specified sensor or record type",
	CodeCannotProvideResponse:               "Command response could not be provided",
	CodeCannotExecuteDuplicatedRequest:      "Cannot execute duplicated request",
	CodeCannotProvideResponseSDRRInUpdate:   "SDR Repository in update mode",
	CodeCannotProvideResponseFirmwareUpdate: "Device firmware in update mode",
	CodeCannotProvideResponseBMCInitialize:  "BMC initialization in progress",
	CodeDestinationUnavailable:              "Destination unavailable",
	CodeInsufficientPrivilege:               "Cannot execute command, insufficient privilege level or other security-based restriction",
	CodeNotSupported:                        "Cannot execute command, command or request parameters not supported in present state",
	CodeSubFnDisabled:                       "Cannot execute command, parameter illegal: sub-function disabled or unavailable",
	CodeUnspecifiedError:                    "Unspecified error",
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
	paramConfigSetCC = map[CompletionCode]string{
		CodeParameterNotSupported:            "Parameter not supported",
		CodeParamConfigSetInProgressConflict: "Attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state",
		CodeParamConfigSetReadOnly:           "Attempt to write read-only parameter",
	}
	paramConfigSetRWCC = map[CompletionCode]string{
		CodeParameterNotSupported:            "Parameter not supported",
		CodeParamConfigSetInProgressConflict: "Attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state",
		CodeParamConfigSetReadOnly:           "Attempt to write read-only parameter",
		CodeParamConfigReadWriteOnly:         "Attempt to read write-only parameter",
	}
	paramConfigGetCC = map[CompletionCode]string{
		CodeParameterNotSupported: "Parameter not supported",
	}
	// selEraseCC: SEL Device commands that cannot run while Clear SEL is in
	// progress (Tables 31-2, 31-4, 31-5) and Set/Get Last Processed Event ID
	// (Tables 30-7, 30-8).
	selEraseCC = map[CompletionCode]string{
		CodeSELEraseInProgress: "Cannot execute command, SEL erase in progress",
	}
	// queueEmptyCC: Get Message (Table 22-7) and Read Event Message Buffer
	// (Table 22-11).
	queueEmptyCC = map[CompletionCode]string{
		CodeDataNotAvailable: "Data not available (queue / buffer empty)",
	}
	// recordMismatchCC: Partial Add SDR (Table 33-8) and Partial Add SEL Entry
	// (Table 31-7) 80h.
	recordMismatchCC = map[CompletionCode]string{
		CodePartialAddRecordMismatch: "Record rejected due to mismatch between record length in header data and number of bytes written",
	}
)

var commandSpecificCC = map[CommandKey]map[CompletionCode]string{
	// v2.0§21 BMC Device and Messaging Firewall commands.
	CommandSetCommandEnables.Key(): {
		CodeSetCommandEnablesUnsupported: "Attempt to enable an unsupported or un-configurable command", // Table 21-7
	},
	CommandSetCommandSubfunctionEnables.Key(): {
		CodeSetCommandSubfunctionEnablesUnsupported: "Attempt to enable an unsupported or un-configurable sub-function", // Table 21-9
	},
	CommandGetCommandSubfunctionEnables.Key(): {
		CodeGetCommandSubfunctionEnablesUnsupported: "Attempt to get an unsupported or un-configurable sub-function", // Table 21-10
	},

	// v2.0§22 messaging and session commands.
	CommandGetMessage.Key():             queueEmptyCC, // Table 22-7
	CommandReadEventMessageBuffer.Key(): queueEmptyCC, // Table 22-11
	CommandSendMessage.Key(): { // Table 22-9
		CodeSendMessageInvalidSessionHandle: "Invalid Session Handle. The session handle does not match up with any currently active sessions for this channel",
		CodeSendMessageLostArbitration:      "Lost Arbitration",
		CodeSendMessageBusError:             "Bus Error",
		CodeSendMessageNAKOnWrite:           "NAK on Write",
	},
	CommandMasterWriteRead.Key(): { // Table 22-14
		CodeMasterWriteReadLostArbitration: "Lost Arbitration",
		CodeMasterWriteReadBusError:        "Bus Error",
		CodeMasterWriteReadNAKOnWrite:      "NAK on Write",
		CodeMasterWriteReadTruncatedRead:   "Truncated Read",
	},
	CommandSetSystemInfoParam.Key(): paramConfigSetCC, // §22.14a
	CommandGetSystemInfoParam.Key(): paramConfigGetCC, // §22.14b
	CommandGetSessionChallenge.Key(): { // Table 22-21
		CodeGetSessionChallengeInvalidUserName:    "Invalid user name",
		CodeGetSessionChallengeNullUserNotEnabled: "Null user name (User 1) not enabled",
	},
	CommandActivateSession.Key(): { // Table 22-22
		CodeActivateSessionNoSlotAvailable:       "No session slot available (BMC cannot accept any more sessions)",
		CodeActivateSessionNoSlotForUser:         "No slot available for given user. (Limit of user sessions allowed under that name has been reached)",
		CodeActivateSessionNoSlotForPrivilege:    "No slot available to support user due to maximum privilege capability",
		CodeActivateSessionSequenceOutOfRange:    "Session sequence number out-of-range",
		CodeActivateSessionInvalidSessionID:      "Invalid Session ID in request",
		CodeActivateSessionPrivilegeExceedsLimit: "Requested maximum privilege level exceeds user and/or channel privilege limit",
	},
	CommandSetSessionPrivilegeLevel.Key(): { // Table 22-24
		CodeSetSessionPrivilegeLevelLevelNotAvailable: "Requested level not available for this user",
		CodeSetSessionPrivilegeLevelExceedsLimit:      "Requested level exceeds Channel and/or User Privilege Limit",
		CodeSetSessionPrivilegeLevelCannotDisableAuth: "Cannot disable User Level authentication",
	},
	CommandCloseSession.Key(): { // Table 22-25
		CodeCloseSessionInvalidSessionID:     "Invalid Session ID in request",
		CodeCloseSessionInvalidSessionHandle: "Invalid Session Handle in request",
	},
	CommandSetChannelAccess.Key(): { // Table 22-27
		CodeSetChannelAccessSetNotSupported:  "Set not supported on selected channel (e.g. channel is session-less)",
		CodeSetChannelAccessModeNotSupported: "Access mode not supported",
	},
	CommandGetChannelAccess.Key(): { // Table 22-28
		CodeGetChannelAccessNotSupported: "Command not supported for selected channel (e.g. channel is session-less)",
	},
	CommandSetChannelSecurityKeys.Key(): { // Table 22-30
		CodeSetChannelSecurityKeysKeyLocked:         "Cannot perform set / confirm. Key is locked",
		CodeSetChannelSecurityKeysInsufficientBytes: "Insufficient key bytes",
		CodeSetChannelSecurityKeysTooManyBytes:      "Too many key bytes",
		CodeSetChannelSecurityKeysKeyValueInvalid:   "Key value does not meet criteria for specified type of key",
		CodeSetChannelSecurityKeysKRNotUsed:         "KR is not used",
	},
	CommandSetUserPassword.Key(): { // Table 22-35
		CodeSetUserPasswordDataMismatch: "Password test failed. Password size correct, but password data does not match stored value",
		CodeSetUserPasswordWrongSize:    "Password test failed. Wrong password size was used",
	},

	// v2.0§23 LAN commands.
	CommandSetLanConfigParam.Key(): paramConfigSetRWCC, // Table 23-2
	CommandGetLanConfigParam.Key(): paramConfigGetCC,   // Table 23-3
	// Suspend BMC ARPs (Table 23-5) defines no command-specific codes.

	// v2.0§24 payload commands.
	CommandActivatePayload.Key(): { // Table 24-2
		CodeActivatePayloadAlreadyActive:                   "Payload already active on another session",
		CodeActivatePayloadTypeDisabled:                    "Payload type is disabled",
		CodeActivatePayloadActivationLimitReached:          "Payload activation limit reached",
		CodeActivatePayloadCannotActivateWithEncryption:    "Cannot activate payload with encryption",
		CodeActivatePayloadCannotActivateWithoutEncryption: "Cannot activate payload without encryption",
	},
	CommandDeactivatePayload.Key(): { // Table 24-3
		CodeDeactivatePayloadAlreadyDeactivated: "Payload already deactivated",
		CodeDeactivatePayloadTypeDisabled:       "Payload type is disabled",
	},
	CommandSuspendResumePayloadEncryption.Key(): { // Table 24-5
		CodeSuspendResumePayloadEncryptionNotSupported: "Operation not supported for given payload type",
		CodeSuspendResumePayloadEncryptionNotAllowed:   "Operation not allowed under present configuration for given payload type",
		CodeSuspendResumePayloadEncryptionNotAvailable: "Encryption is not available for session that payload type is active under",
		CodeSuspendResumePayloadEncryptionNotActive:    "The payload instance is not presently active",
	},
	CommandGetChannelPayloadVersion.Key(): { // Table 24-11
		CodeGetChannelPayloadVersionNotAvailable: "Payload type not available on given channel",
	},
	CommandGetChannelOEMPayloadInfo.Key(): { // Table 24-12
		CodeGetChannelOEMPayloadInfoNotSupported: "OEM Payload IANA and/or Payload ID not supported",
	},

	// v2.0§25 serial/modem and PPP commands.
	CommandSetSerialConfig.Key(): paramConfigSetRWCC, // Table 25-2
	CommandGetSerialConfig.Key(): paramConfigGetCC,   // Table 25-3
	CommandSendPPPPacket.Key(): { // Table 25-9
		CodeSendPPPPacketLinkNotUp:     "PPP Link is not up",
		CodeSendPPPPacketProtocolNotUp: "IP Protocol is not up",
	},
	CommandGetPPPReceiveData.Key(): { // Table 25-10
		CodeGetPPPReceiveDataNoData: "No packet data available",
	},
	CommandCallback.Key(): { // Table 25-12
		CodeCallbackAlertInProgress: "Callback rejected due to alert in progress on this channel",
		CodeCallbackSessionActive:   "Callback rejected due to IPMI messaging session active on the callback channel",
	},

	// v2.0§26 SOL commands.
	CommandSetSOLConfigParam.Key(): paramConfigSetRWCC, // Table 26-3
	CommandGetSOLConfigParam.Key(): paramConfigGetCC,   // Table 26-4

	// v2.0§27 watchdog commands.
	CommandResetWatchdogTimer.Key(): { // Table 27-2
		CodeResetWatchdogTimerUninitialized: "Attempt to start un-initialized watchdog",
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
		CodeAlertImmediateAlreadyInProgress:  "Alert Immediate rejected due to alert already in progress",
		CodeAlertImmediateSessionActive:      "Alert Immediate rejected due to IPMI messaging session active on this channel",
		CodeAlertImmediateParamsNotSupported: "Platform Event Parameters (4:11) not supported",
	},
	// PET Acknowledge (Table 30-10) defines no command-specific codes.

	// v2.0§31 SEL Device commands.
	CommandGetSELInfo.Key():  selEraseCC, // Table 31-2
	CommandReserveSEL.Key():  selEraseCC, // Table 31-4
	CommandGetSELEntry.Key(): selEraseCC, // Table 31-5
	CommandAddSELEntry.Key(): { // Table 31-6
		CodeAddSELEntryRecordTypeNotSupported: "Operation not supported for this Record Type",
		CodeSELEraseInProgress:                "Cannot execute command, SEL erase in progress",
	},
	CommandPartialAddSELEntry.Key(): { // Table 31-7
		CodePartialAddRecordMismatch: "Record rejected due to mismatch between record length in header data and number of bytes written",
		CodeSELEraseInProgress:       "Cannot execute command, SEL erase in progress",
	},
	CommandDeleteSELEntry.Key(): { // Table 31-8
		CodeDeleteSELEntryRecordTypeNotSupported: "Operation not supported for this Record Type",
		CodeSELEraseInProgress:                   "Cannot execute command, SEL erase in progress",
	},

	// v2.0§33 SDR Repository commands.
	CommandPartialAddSDR.Key(): recordMismatchCC, // Table 33-8

	// v2.0§34 FRU commands.
	CommandReadFRUData.Key(): { // Table 34-3
		CodeReadFRUDataDeviceBusy: "FRU device busy",
	},
	CommandWriteFRUData.Key(): { // Table 34-4
		CodeWriteFRUDataWriteProtected: "Write-protected offset",
		CodeWriteFRUDataDeviceBusy:     "FRU device busy",
	},

	// v2.0§35 sensor commands.
	CommandGetDeviceSDR.Key(): { // Table 35-3
		CodeGetDeviceSDRRecordChanged: "Record changed",
	},
	CommandSetSensorReadingAndEventStatus.Key(): { // Table 35-xx Set Sensor Reading and Event Status
		CodeSetSensorReadingNotSettable:           "Attempt to change reading or set or clear status bits that are not settable via this command",
		CodeSetSensorReadingEventDataNotSupported: "Attempted to set Event Data Bytes, but setting Event Data Bytes is not supported for this sensor",
	},

	// v2.0§35b ICMB bridging.
	CommandForwarded.Key(): { // Table 35b-5
		CodeForwardedTargetUnavailable: "Target controller unavailable",
	},

	// DCMI (dcmi-spec-v1.5 §6). Only codes listed in the command tables; §8
	// Table 8-1 mirrors IPMI generic codes only. Set/Get DCMI Configuration
	// Parameters (Tables 6-4/6-5) and Get Temperature Readings (Table 6-22)
	// define no command-specific codes.
	CommandGetDCMIAssetTag.Key(): { // dcmi§6.4.2 Table 6-8
		CodeGetDCMIAssetTagBinaryEncoding:    "Encoding type in FRU is binary / unspecified",
		CodeGetDCMIAssetTagBCDPlus:           "Encoding type in FRU is BCD Plus",
		CodeGetDCMIAssetTag6BitASCII:         "Encoding type in FRU is 6-bit ASCII Packed",
		CodeGetDCMIAssetTagUnicodeNonEnglish: "Encoding type in FRU is set to ASCII+Latin1, but language code is not set to English (indicating data is 2-byte UNICODE)",
	},
	CommandGetDCMIPowerLimit.Key(): { // dcmi§6.6.2 Table 6-17
		CodeGetDCMIPowerLimitNoActiveLimit: "No Active Set Power Limit",
	},
	CommandSetDCMIPowerLimit.Key(): { // dcmi§6.6.3 Table 6-18
		CodeSetDCMIPowerLimitOutOfRange:               "Power Limit out of range",
		CodeSetDCMIPowerLimitCorrectionTimeOutOfRange: "Correction Time out of range",
		CodeSetDCMIPowerLimitStatsPeriodOutOfRange:    "Statistics Reporting Period out of range",
	},
	CommandSetDCMIThermalLimit.Key(): { // dcmi§6.7.2 Table 6-21
		CodeSetDCMIThermalLimitOutOfRange:              "Thermal Limit out of range",
		CodeSetDCMIThermalLimitExceptionTimeOutOfRange: "Exception Time out of range",
	},
}

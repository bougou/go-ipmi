package ipmi

type CompletionCode uint8

// 5.2 Table 5 for generic completion codes
const (
	// GENERIC COMPLETION CODES 00h, C0h-FFh
	CompletionCodeNormal                               = CompletionCode(0x00)
	CompletionCodeNodeBusy                             = CompletionCode(0xC0)
	CompletionCodeInvalidCommand                       = CompletionCode(0xC1)
	CompletionCodeInvalidCommandForLUN                 = CompletionCode(0xC2)
	CompletionCodeProcessTimeout                       = CompletionCode(0xC3)
	CompletionCodeOutOfSpace                           = CompletionCode(0xC4)
	CompletionCodeReservationCanceled                  = CompletionCode(0xC5)
	CompletionCodeRequestDataTruncated                 = CompletionCode(0xC6)
	CompletionCodeRequestDataLengthInvalid             = CompletionCode(0xC7)
	CompletionCodeRequestDataLengthLimitExceeded       = CompletionCode(0xC8)
	CompletionCodeParameterOutOfRange                  = CompletionCode(0xC9)
	CompletionCodeCannotReturnRequestedDataBytes       = CompletionCode(0xCA)
	CompletionCodeRequestedDataNotPresent              = CompletionCode(0xCB)
	CompletionCodeRequestDataFieldInvalid              = CompletionCode(0xCC)
	CompletionCodeIllegalCommand                       = CompletionCode(0xCD)
	CompletionCodeCannotProvideResponse                = CompletionCode(0xCE)
	CompletionCodeCannotExecuteDuplicatedRequest       = CompletionCode(0xCF)
	CompletionCodeCannotProvideResponseSDRRInUpdate    = CompletionCode(0xD0)
	CompletionCodeCannotProvideResponseFirmwareUpdate  = CompletionCode(0xD1)
	CompletionCodeCannotProvideResponseBMCInitialize   = CompletionCode(0xD2)
	CompletionCodeDestinationUnavailable               = CompletionCode(0xD3)
	CompletionCodeCannotExecuteCommandSecurityRestrict = CompletionCode(0xD4)
	CompletionCodeCannotExecuteCommandNotSupported     = CompletionCode(0xD5)
	CompletionCodeCannotExecuteCommandSubFnDisabled    = CompletionCode(0xD6)
	CompletionCodeUnspecifiedError                     = CompletionCode(0xFF)

	// DEVICE-SPECIFIC (OEM) CODES 01h-7Eh

	// COMMAND-SPECIFIC CODES 80h-BEh
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
	0xd1: "Device firmeware in update mode",
	0xd2: "BMC initialization in progress",
	0xd3: "Destination unavailable",
	0xd4: "Insufficient privilege level",
	0xd5: "Command not supported in present state",
	0xd6: "Cannot execute command, command disabled",
	0xff: "Unspecified error",
}

// String return description of global completion code.
// Please use StrCC function to get description for any completion code
// returned for specific command response.
func (cc CompletionCode) String() string {
	if s, ok := CC[uint8(cc)]; ok {
		return s
	}
	return ""
}

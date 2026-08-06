package types

import (
	"errors"
	"fmt"
)

var (
	ErrUnpackedDataTooShort         = errors.New("unpacked data is too short")
	ErrDCMIGroupExtensionIDMismatch = errors.New("DCMI group extension ID mismatch")
)

// RmcpStatusError reports a non-zero RMCP+ Open Session or RAKP status code.
type RmcpStatusError struct {
	statusCode RmcpStatusCode
}

func (e *RmcpStatusError) Error() string {
	return fmt.Sprintf("RMCP+ status %#02x: %s", uint8(e.statusCode), e.statusCode)
}

// StatusCode returns the RMCP+ status code reported by the BMC.
func (e *RmcpStatusError) StatusCode() RmcpStatusCode {
	return e.statusCode
}

// NewRmcpStatusError creates an error for a non-zero RMCP+ status code.
func NewRmcpStatusError(statusCode RmcpStatusCode) *RmcpStatusError {
	return &RmcpStatusError{statusCode: statusCode}
}

func ErrUnpackedDataTooShortWith(actual int, expected int) error {
	return fmt.Errorf("%w (%d/%d)", ErrUnpackedDataTooShort, actual, expected)
}

func ErrNotEnoughDataWith(msg string, actual int, expected int) error {
	return fmt.Errorf("not enough data for %s (%d/%d)", msg, actual, expected)
}

func ErrDCMIGroupExtensionIDMismatchWith(expected uint8, actual uint8) error {
	return fmt.Errorf("%w: expected %#02x, got %#02x", ErrDCMIGroupExtensionIDMismatch, expected, actual)
}

func CheckDCMIGroupExenstionMatch(grpExt uint8) error {
	if grpExt != GroupExtensionDCMI {
		return ErrDCMIGroupExtensionIDMismatchWith(GroupExtensionDCMI, grpExt)
	}
	return nil
}

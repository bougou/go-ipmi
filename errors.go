package ipmi

import (
	"errors"
	"fmt"
)

var (
	ErrUnpackedDataTooShort         = errors.New("unpacked data is too short")
	ErrDCMIGroupExtensionIDMismatch = errors.New("DCMI group extension ID mismatch")
)

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

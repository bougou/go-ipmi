package ipmi

import (
	"errors"
	"fmt"
)

var (
	ErrUnpackedDataTooShort = errors.New("unpacked data is too short")
)

func ErrUnpackedDataTooShortWith(actual int, expected int) error {
	return fmt.Errorf("%s (%d/%d)", ErrUnpackedDataTooShort, actual, expected)
}

func ErrNotEnoughDataWith(msg string, actual int, expected int) error {
	return fmt.Errorf("not enough data for %s (%d/%d)", msg, actual, expected)
}

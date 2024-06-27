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

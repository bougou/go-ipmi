package ipmi

import (
	"errors"
)

var (
	ErrUnpackedDataTooShort = errors.New("unpacked data is too short")
)

package handlers

import (
	"errors"

	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// codeFromErr maps a HAL error to a completion code.
// If err carries a [types.CompletionCode] (e.g. HAL returns [types.CodeNodeBusy] directly),
// that code is extracted; otherwise the error is mapped to [types.CodeUnspecifiedError].
func codeFromErr(err error) types.CompletionCode {
	if err == nil {
		return types.CodeOK
	}
	var cc types.CompletionCode
	if errors.As(err, &cc) {
		return cc
	}
	return types.CodeUnspecifiedError
}

// codeFromHalErr maps a HAL error to a completion code.
// It delegates to [codeFromErr] and additionally maps [hal.ErrNotSupported]
// to [types.CodeParameterNotSupported].
func codeFromHalErr(err error) types.CompletionCode {
	if cc := codeFromErr(err); cc != types.CodeUnspecifiedError {
		return cc
	}
	if errors.Is(err, hal.ErrNotSupported) {
		return types.CodeParameterNotSupported
	}
	return types.CodeUnspecifiedError
}

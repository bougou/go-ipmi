package client

import "errors"

// ErrRAKPAuthentication reports that a RAKP authentication code or integrity check did not match.
var ErrRAKPAuthentication = errors.New("RAKP authentication failed")

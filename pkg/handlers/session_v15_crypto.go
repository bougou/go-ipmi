package handlers

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

// GenV15AuthCode computes the IPMI v1.5 multi-session AuthCode per spec
// v1.5§18.15.1 Figure 18-1 / v2.0§22.17.1 Figure 22-1.
func GenV15AuthCode(password []byte, authType bmc.V15AuthType, sessionID uint32, ipmiData []byte, sessionSeq uint32) []byte {
	return crypto.AuthCodeMultiSession(types.AuthType(authType), password, sessionID, sessionSeq, ipmiData)
}

// VerifyV15AuthCode returns true when got matches the expected AuthCode.
func VerifyV15AuthCode(password []byte, authType bmc.V15AuthType, sessionID uint32, ipmiData []byte, sessionSeq uint32, got []byte) bool {
	return crypto.VerifyMultiSessionAuthCode(types.AuthType(authType), password, sessionID, sessionSeq, ipmiData, got)
}

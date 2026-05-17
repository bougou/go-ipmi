package app

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 22.14b Get System Info Parameters Command
type GetSystemInfoParamRequest struct {
	GetParamRevisionOnly bool
	ParamSelector        ipmi.SystemInfoParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSystemInfoParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSystemInfoParamRequest) Pack() []byte {
	out := make([]byte, 4)

	var b uint8
	b = ipmi.SetOrClearBit7(b, req.GetParamRevisionOnly)
	out[0] = b

	out[1] = uint8(req.ParamSelector)
	out[2] = req.SetSelector
	out[3] = req.BlockSelector

	return out
}

func (req *GetSystemInfoParamRequest) Command() ipmi.Command {
	return ipmi.CommandGetSystemInfoParam
}

func (res *GetSystemInfoParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemInfoParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision, _, _ = ipmi.UnpackUint8(msg, 0)
	if len(msg) > 1 {
		res.ParamData, _, _ = ipmi.UnpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSystemInfoParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Param Revision : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param Data     : %v\n", res.ParamData)
}

// parameter not supported

// For the first block of string data (set selector = 0),
// the first two bytes indicate the encoding of the string and its overall length as follows.
// So, if the length is less than 2, it means there is no string data.

// Sets count is based on first two bytes + string length.

// 1 set per 16 bytes. Subtract 1 before dividing by 16, else multiples
// of 16 would get an extra set.

func getSystemInfoStringMeta(params []any) (s string, stringDataRaw []byte, stringDataType uint8, stringDataLength uint8) {
	if len(params) == 0 {
		return
	}

	array := make([]ipmi.SystemInfoParameter, 0)
	for _, param := range params {
		v, ok := param.(ipmi.SystemInfoParameter)
		if ok {
			array = append(array, v)
		}
	}

	allBlockData := make([]byte, 0)

	for _, p := range array {
		_, setSelector, _ := p.SystemInfoParameter()
		paramData := p.Pack()
		blockData := paramData[1:]
		if setSelector == 0 {
			stringDataType = blockData[0]
			stringDataLength = blockData[1]
		}
		allBlockData = append(allBlockData, blockData[:]...)
	}

	stringDataRaw = allBlockData[2 : stringDataLength+2]

	switch stringDataType {
	// 0h = ASCII+Latin1
	// 1h = UTF-8
	// 2h = UNICODE
	// all other = reserved.
	case 0x00:
		s = string(stringDataRaw)
	case 0x01:
		s = string(stringDataRaw)
	case 0x02:
		// here, suppose UTF-16
		u16 := make([]uint16, len(stringDataRaw)/2)
		for i := 0; i < len(u16); i++ {
			u16[i] = binary.BigEndian.Uint16(stringDataRaw[i*2 : i*2+2])
		}
		// Decode UTF-16 to UTF-8
		runes := utf16.Decode(u16)
		s = string(runes)
	default:
		s = string(stringDataRaw)
	}

	return
}

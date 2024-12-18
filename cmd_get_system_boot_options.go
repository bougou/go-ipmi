package ipmi

import (
	"context"
	"fmt"
)

// 28.13 Get System Boot Options Command
type GetSystemBootOptionsRequest struct {
	ParameterSelector BootOptionParameterSelector
	SetSelector       uint8
	BlockSelector     uint8
}

// Table 28-14, Boot Option Parameters

type GetSystemBootOptionsResponse struct {
	ParameterVersion uint8

	// [7] - 1b = mark parameter invalid / locked
	// 0b = mark parameter valid / unlocked
	ParameterInValid bool
	// [6:0] - boot option parameter selector
	ParameterSelector BootOptionParameterSelector

	parameterData []byte // origin parameter data

	// parameterData is automatically parsed to BootOptionParameter
	BootOptionParameter *BootOptionParameter
}

func (req *GetSystemBootOptionsRequest) Pack() []byte {
	out := make([]byte, 3)
	packUint8(uint8(req.ParameterSelector), out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)
	return out
}

func (req *GetSystemBootOptionsRequest) Command() Command {
	return CommandGetSystemBootOptions
}

func (res *GetSystemBootOptionsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemBootOptionsResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.ParameterVersion, _, _ = unpackUint8(msg, 0)
	b, _, _ := unpackUint8(msg, 1)
	res.ParameterInValid = isBit7Set(b)
	res.ParameterSelector = BootOptionParameterSelector(b & 0x7f) // clear bit 7

	if len(msg) > 2 {
		parameterData, _, _ := unpackBytes(msg, 2, len(msg)-2)
		res.parameterData = parameterData

		bop, err := ParseBootOptionParameterData(res.ParameterSelector, parameterData)
		if err != nil {
			return fmt.Errorf("parse ParameterData failed, err: %s", err)
		}
		res.BootOptionParameter = bop
	}

	return nil
}

func (res *GetSystemBootOptionsResponse) Format() string {
	return fmt.Sprintf(`Boot parameter version: %d
Boot parameter %d is %s
Boot parameter data: %02x
%s`,
		res.ParameterVersion,
		res.ParameterSelector, formatBool(res.ParameterInValid, "invalid/locked", "valid/unlocked"),
		res.parameterData,
		res.BootOptionParameter.Format(res.ParameterSelector))
}

// This command is used to set parameters that direct the system boot following a system power up or reset.
// The boot flags only apply for one system restart. It is the responsibility of the system BIOS
// to read these settings from the BMC and then clear the boot flags
func (c *Client) GetSystemBootOptions(ctx context.Context, parameterSelector BootOptionParameterSelector) (response *GetSystemBootOptionsResponse, err error) {
	request := &GetSystemBootOptionsRequest{
		ParameterSelector: parameterSelector,
		SetSelector:       0x00,
		BlockSelector:     0x00,
	}
	response = &GetSystemBootOptionsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

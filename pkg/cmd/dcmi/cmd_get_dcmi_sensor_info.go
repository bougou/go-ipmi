package dcmi

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5]: 6.5.2 Get DCMI Sensor Info Command
type GetDCMISensorInfoRequest struct {
	SensorType ipmi.SensorType
	EntityID   ipmi.EntityID

	// 00h Retrieve information about all instances associated with Entity ID
	// 01h - FFh Retrieve only the information about particular instance.
	EntityInstance      ipmi.EntityInstance
	EntityInstanceStart uint8
}

type GetDCMISensorInfoResponse struct {
	TotalEntityInstances uint8
	RecordsCount         uint8
	SDRRecordID          []uint16
}

func (req *GetDCMISensorInfoRequest) Pack() []byte {
	out := make([]byte, 5)
	ipmi.PackUint8(ipmi.GroupExtensionDCMI, out, 0)
	ipmi.PackUint8(uint8(req.SensorType), out, 1)
	ipmi.PackUint8(byte(req.EntityID), out, 2)
	ipmi.PackUint8(byte(req.EntityInstance), out, 3)
	ipmi.PackUint8(req.EntityInstanceStart, out, 4)

	return out
}

func (req *GetDCMISensorInfoRequest) Command() ipmi.Command {
	return ipmi.CommandGetDCMISensorInfo
}

func (res *GetDCMISensorInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMISensorInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	if err := ipmi.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalEntityInstances = msg[1]
	res.RecordsCount = msg[2]

	if len(msg) < 3+int(res.RecordsCount)*2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 3+int(res.RecordsCount)*2)
	}

	res.SDRRecordID = make([]uint16, res.RecordsCount)
	for i := 0; i < int(res.RecordsCount); i++ {
		res.SDRRecordID[i], _, _ = ipmi.UnpackUint16L(msg, 3+i*2)
	}

	return nil
}

func (res *GetDCMISensorInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Total entity instances : %d\n", res.TotalEntityInstances) +
		fmt.Sprintf("Number of records      : %d\n", res.RecordsCount) +
		fmt.Sprintf("SDR Record ID          : %v\n", res.SDRRecordID)
}

// GetDCMISensorInfo sends a DCMI "Get Power Reading" command.
// See [GetDCMISensorInfoRequest] for details.

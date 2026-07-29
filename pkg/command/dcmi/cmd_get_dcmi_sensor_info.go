package dcmi

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// [DCMI specification v1.5]: 6.5.2 Get DCMI Sensor Info Command
type GetDCMISensorInfoRequest struct {
	SensorType types.SensorType
	EntityID   types.EntityID

	// 00h Retrieve information about all instances associated with Entity ID
	// 01h - FFh Retrieve only the information about particular instance.
	EntityInstance      types.EntityInstance
	EntityInstanceStart uint8
}

type GetDCMISensorInfoResponse struct {
	TotalEntityInstances uint8
	RecordsCount         uint8
	SDRRecordID          []uint16
}

func (req *GetDCMISensorInfoRequest) Pack() []byte {
	out := make([]byte, 5)
	types.PackUint8(types.GroupExtensionDCMI, out, 0)
	types.PackUint8(uint8(req.SensorType), out, 1)
	types.PackUint8(byte(req.EntityID), out, 2)
	types.PackUint8(byte(req.EntityInstance), out, 3)
	types.PackUint8(req.EntityInstanceStart, out, 4)

	return out
}

func (req *GetDCMISensorInfoRequest) Command() types.Command {
	return types.CommandGetDCMISensorInfo
}

func (res *GetDCMISensorInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMISensorInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	if err := types.CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalEntityInstances = msg[1]
	res.RecordsCount = msg[2]

	if len(msg) < 3+int(res.RecordsCount)*2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3+int(res.RecordsCount)*2)
	}

	res.SDRRecordID = make([]uint16, res.RecordsCount)
	for i := 0; i < int(res.RecordsCount); i++ {
		res.SDRRecordID[i], _, _ = types.UnpackUint16L(msg, 3+i*2)
	}

	return nil
}

func (res *GetDCMISensorInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Total entity instances : %d\n", res.TotalEntityInstances) +
		fmt.Sprintf("Number of records      : %d\n", res.RecordsCount) +
		fmt.Sprintf("SDR Record ID          : %v\n", res.SDRRecordID)
}

package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5]: 6.5.2 Get DCMI Sensor Info Command
type GetDCMISensorInfoRequest struct {
	SensorType SensorType
	EntityID   EntityID

	// 00h Retrieve information about all instances associated with Entity ID
	// 01h - FFh Retrieve only the information about particular instance.
	EntityInstance      EntityInstance
	EntityInstanceStart uint8
}

type GetDCMISensorInfoResponse struct {
	TotalEntityInstances uint8
	NumberOfRecords      uint8
	SDRRecordID          []uint16
}

func (req *GetDCMISensorInfoRequest) Pack() []byte {
	out := make([]byte, 5)
	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.SensorType), out, 1)
	packUint8(byte(req.EntityID), out, 2)
	packUint8(byte(req.EntityInstance), out, 3)
	packUint8(req.EntityInstanceStart, out, 4)

	return out
}

func (req *GetDCMISensorInfoRequest) Command() Command {
	return CommandGetDCMISensorInfo
}

func (res *GetDCMISensorInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMISensorInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.TotalEntityInstances = msg[1]
	res.NumberOfRecords = msg[2]

	if len(msg) < 3+int(res.NumberOfRecords)*2 {
		return ErrUnpackedDataTooShortWith(len(msg), 3+int(res.NumberOfRecords)*2)
	}

	res.SDRRecordID = make([]uint16, res.NumberOfRecords)
	for i := 0; i < int(res.NumberOfRecords); i++ {
		res.SDRRecordID[i], _, _ = unpackUint16L(msg, 3+i*2)
	}

	return nil
}

func (res *GetDCMISensorInfoResponse) Format() string {
	return fmt.Sprintf(`
Total entity instances: %d
Number of records: %d
SDR Record ID: %v
`,
		res.TotalEntityInstances,
		res.NumberOfRecords,
		res.SDRRecordID,
	)
}

// GetDCMISensorInfo sends a DCMI "Get Power Reading" command.
// See [GetDCMISensorInfoRequest] for details.
func (c *Client) GetDCMISensorInfo(ctx context.Context, request *GetDCMISensorInfoRequest) (response *GetDCMISensorInfoResponse, err error) {
	response = &GetDCMISensorInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMISensors(ctx context.Context, entityIDs ...EntityID) ([]*SDR, error) {
	out := make([]*SDR, 0)

	for _, entityID := range entityIDs {
		request := &GetDCMISensorInfoRequest{
			SensorType:          SensorTypeTemperature,
			EntityID:            entityID,
			EntityInstance:      0x00,
			EntityInstanceStart: 0,
		}

		response := &GetDCMISensorInfoResponse{}
		if err := c.Exchange(ctx, request, response); err != nil {
			return nil, err
		}

		for _, recordID := range response.SDRRecordID {
			sdr, err := c.GetSDREnhanced(ctx, recordID)
			if err != nil {
				return nil, fmt.Errorf("GetSDRDetail failed for recordID (%#02x), err: %s", recordID, err)
			}
			out = append(out, sdr)
		}
	}

	return out, nil
}

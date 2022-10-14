package ipmi

// NetFn is Network Function
type NetFn uint8

// Network Function Codes, section 5.1 Table 5
// Even NetFn values are used for requests to the BMC,
// and odd NetFn values are returned in responses from the BMC.
//
// six-bit field identifying the function, so total 64 NetFn (32 NetFn pairs)
const (
	NetFnChassisRequest      NetFn = 0x00
	NetFnChassisResponse     NetFn = 0x01
	NetFnBridgeRequest       NetFn = 0x02
	NetFnBridgeResponse      NetFn = 0x03
	NetFnSensorEventRequest  NetFn = 0x04
	NetFnSensorEventResponse NetFn = 0x05
	NetFnAppRequest          NetFn = 0x06
	NetFnAppResponse         NetFn = 0x07
	NetFnFirmwareRequest     NetFn = 0x08
	NetFnFirmwareResponse    NetFn = 0x09
	NetFnStorageRequest      NetFn = 0x0a
	NetFnStorageResponse     NetFn = 0x0b
	NetFnTransportRequest    NetFn = 0x0c
	NetFnTransportResponse   NetFn = 0x0d

	// Reserverd  0E - 2B

	NetFnGroupExtensionRequest  NetFn = 0x2c
	NetFnGroupExtensionResponse NetFn = 0x2d
	NetFnOEMGroupRequest        NetFn = 0x2e
	NetFnOEMGroupResponse       NetFn = 0x2f

	// 30h-3Fh controller specific
	// Vendor specific (16 Network Functions [8 pairs]).

	NetFnOEMSupermicroRequest NetFn = 0x30
)

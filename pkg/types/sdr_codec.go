package types

import "fmt"

// PackSDR serializes a parsed [SDR] back to wire bytes per v2.0§43.
func PackSDR(sdr *SDR) ([]byte, error) {
	if sdr == nil || sdr.RecordHeader == nil {
		return nil, fmt.Errorf("types: PackSDR: nil sdr or record header")
	}

	recordID := sdr.RecordHeader.RecordID
	switch sdr.RecordHeader.RecordType {
	case SDRRecordTypeFullSensor:
		if sdr.Full == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Full sensor body")
		}
		return sdr.Full.Pack(recordID), nil
	case SDRRecordTypeCompactSensor:
		if sdr.Compact == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Compact sensor body")
		}
		return sdr.Compact.Pack(recordID), nil
	case SDRRecordTypeEventOnly:
		if sdr.EventOnly == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Event-only sensor body")
		}
		return sdr.EventOnly.Pack(recordID), nil
	case SDRRecordTypeEntityAssociation:
		if sdr.EntityAssociation == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Entity Association body")
		}
		return sdr.EntityAssociation.Pack(recordID), nil
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		if sdr.DeviceRelative == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Device-relative Entity Association body")
		}
		return sdr.DeviceRelative.Pack(recordID), nil
	case SDRRecordTypeGenericLocator:
		if sdr.GenericDeviceLocator == nil {
			return nil, fmt.Errorf("types: PackSDR: nil Generic Device Locator body")
		}
		return sdr.GenericDeviceLocator.Pack(recordID), nil
	case SDRRecordTypeFRUDeviceLocator:
		if sdr.FRUDeviceLocator == nil {
			return nil, fmt.Errorf("types: PackSDR: nil FRU Device Locator body")
		}
		return sdr.FRUDeviceLocator.Pack(recordID), nil
	case SDRRecordTypeManagementControllerDeviceLocator:
		if sdr.MgmtControllerDeviceLocator == nil {
			return nil, fmt.Errorf("types: PackSDR: nil MC Device Locator body")
		}
		return sdr.MgmtControllerDeviceLocator.Pack(recordID), nil
	case SDRRecordTypeManagementControllerConfirmation:
		if sdr.MgmtControllerConfirmation == nil {
			return nil, fmt.Errorf("types: PackSDR: nil MC Confirmation body")
		}
		return sdr.MgmtControllerConfirmation.Pack(recordID), nil
	case SDRRecordTypeBMCMessageChannelInfo:
		if sdr.BMCChannelInfo == nil {
			return nil, fmt.Errorf("types: PackSDR: nil BMC Message Channel Info body")
		}
		return sdr.BMCChannelInfo.Pack(recordID), nil
	case SDRRecordTypeOEM:
		if sdr.OEM == nil {
			return nil, fmt.Errorf("types: PackSDR: nil OEM body")
		}
		return sdr.OEM.Pack(recordID), nil
	default:
		return nil, fmt.Errorf("types: PackSDR: unsupported record type %#02x", sdr.RecordHeader.RecordType)
	}
}

// packSDRWire wraps a record body with the five-byte SDR header (v2.0§43).
func packSDRWire(recordID uint16, recordType SDRRecordType, body []byte) []byte {
	rec := make([]byte, SDRRecordHeaderSize+len(body))
	PackUint16L(recordID, rec, 0)
	rec[2] = SDRCommandSetVersion
	rec[3] = byte(recordType)
	rec[4] = byte(len(body))
	copy(rec[SDRRecordHeaderSize:], body)
	return rec
}

// packSensorRecordSharing encodes the 2-byte Sensor Record Sharing / Sensor
// Direction field used by Compact (v2.0§43.2) and Event-Only (v2.0§43.3) SDRs:
//
//	Byte1: [7:6] Direction, [5:4] ID String Instance Modifier Type, [3:0] Share Count
//	Byte2: [7] Entity Instance Sharing, [6:0] ID String Instance Modifier Offset
func packSensorRecordSharing(direction, modifierType, shareCount uint8, entitySharing bool, offset uint8) (b1, b2 uint8) {
	b1 = ((direction & 0x03) << 6) | ((modifierType & 0x03) << 4) | (shareCount & 0x0f)
	b2 = offset & 0x7f
	if entitySharing {
		b2 = SetBit7(b2)
	}
	return b1, b2
}

func unpackSensorRecordSharing(b1, b2 uint8) (direction, modifierType, shareCount uint8, entitySharing bool, offset uint8) {
	direction = (b1 >> 6) & 0x03
	modifierType = (b1 >> 4) & 0x03
	shareCount = b1 & 0x0f
	entitySharing = IsBit7Set(b2)
	offset = b2 & 0x7f
	return
}

func packASCIITypeLengthField(s string) []byte {
	if s == "" {
		return []byte{0xC0}
	}
	if len(s) > 0x3f {
		s = s[:0x3f]
	}
	out := make([]byte, 1+len(s))
	out[0] = 0xC0 | byte(len(s))
	copy(out[1:], []byte(s))
	return out
}

func packEntityInstanceByte(instance EntityInstance, logical bool) uint8 {
	b := uint8(instance) & 0x7f
	if logical {
		b |= 0x80
	}
	return b
}

func packSensorInitializationByte(s SensorInitialization) uint8 {
	var b uint8
	if s.Settable {
		b = SetBit7(b)
	}
	if s.InitScanning {
		b = SetBit6(b)
	}
	if s.InitEvents {
		b = SetBit5(b)
	}
	if s.InitThresholds {
		b = SetBit4(b)
	}
	if s.InitHysteresis {
		b = SetBit3(b)
	}
	if s.InitSensorType {
		b = SetBit2(b)
	}
	if s.EventGenerationEnabled {
		b = SetBit1(b)
	}
	if s.SensorScanningEnabled {
		b = SetBit0(b)
	}
	return b
}

func packSensorCapabilitiesByte(s SensorCapabilities) uint8 {
	var b uint8
	if s.IgnoreSensorIfNoEntity {
		b = SetBit7(b)
	}
	if s.AutoRearm {
		b = SetBit6(b)
	}
	b |= uint8(s.HysteresisAccess&0x03) << 4
	b |= uint8(s.ThresholdAccess&0x03) << 2
	b |= uint8(s.EventMessageControl & 0x03)
	return b
}

func packSensorUnitByte(u SensorUnit) (b20, b21, b22 uint8) {
	b20 = uint8(u.AnalogDataFormat&0x03) << 6
	b20 |= uint8(u.RateUnit&0x07) << 3
	b20 |= uint8(u.ModifierRelation&0x03) << 1
	if u.Percentage {
		b20 = SetBit0(b20)
	}
	b21 = uint8(u.BaseUnit)
	b22 = uint8(u.ModifierUnit)
	return b20, b21, b22
}

// The packMask* functions return a mask field in the bit numbering of the
// spec, to be written with PackUint16L; see [Mask].

func packDiscreteEventMask(e Mask_DiscreteEvent) uint16 {
	var b uint16
	set := func(n uint, v bool) {
		if v {
			b |= 1 << n
		}
	}
	set(0, e.State_0)
	set(1, e.State_1)
	set(2, e.State_2)
	set(3, e.State_3)
	set(4, e.State_4)
	set(5, e.State_5)
	set(6, e.State_6)
	set(7, e.State_7)
	set(8, e.State_8)
	set(9, e.State_9)
	set(10, e.State_10)
	set(11, e.State_11)
	set(12, e.State_12)
	set(13, e.State_13)
	set(14, e.State_14)
	return b
}

func packMaskAssertLower(m Mask) uint16 {
	b := packDiscreteEventMask(m.Discrete.Assert)
	set := func(n uint, v bool) {
		if v {
			b |= 1 << n
		}
	}
	set(14, m.Threshold.LNR.StatusReturned)
	set(13, m.Threshold.LCR.StatusReturned)
	set(12, m.Threshold.LNC.StatusReturned)
	set(11, m.Threshold.UNR.High_Assert)
	set(10, m.Threshold.UNR.Low_Assert)
	set(9, m.Threshold.UCR.High_Assert)
	set(8, m.Threshold.UCR.Low_Assert)
	set(7, m.Threshold.UNC.High_Assert)
	set(6, m.Threshold.UNC.Low_Assert)
	set(5, m.Threshold.LNR.High_Assert)
	set(4, m.Threshold.LNR.Low_Assert)
	set(3, m.Threshold.LCR.High_Assert)
	set(2, m.Threshold.LCR.Low_Assert)
	set(1, m.Threshold.LNC.High_Assert)
	set(0, m.Threshold.LNC.Low_Assert)
	return b
}

func packMaskDeassertUpper(m Mask) uint16 {
	b := packDiscreteEventMask(m.Discrete.Deassert)
	set := func(n uint, v bool) {
		if v {
			b |= 1 << n
		}
	}
	set(14, m.Threshold.UNR.StatusReturned)
	set(13, m.Threshold.UCR.StatusReturned)
	set(12, m.Threshold.UNC.StatusReturned)
	set(11, m.Threshold.UNR.High_Deassert)
	set(10, m.Threshold.UNR.Low_Deassert)
	set(9, m.Threshold.UCR.High_Deassert)
	set(8, m.Threshold.UCR.Low_Deassert)
	set(7, m.Threshold.UNC.High_Deassert)
	set(6, m.Threshold.UNC.Low_Deassert)
	set(5, m.Threshold.LNR.High_Deassert)
	set(4, m.Threshold.LNR.Low_Deassert)
	set(3, m.Threshold.LCR.High_Deassert)
	set(2, m.Threshold.LCR.Low_Deassert)
	set(1, m.Threshold.LNC.High_Deassert)
	set(0, m.Threshold.LNC.Low_Deassert)
	return b
}

func packMaskReading(m Mask) uint16 {
	b := packDiscreteEventMask(m.Discrete.Reading)
	set := func(n uint, v bool) {
		if v {
			b |= 1 << n
		}
	}
	set(13, m.Threshold.UNR.Settable)
	set(12, m.Threshold.UCR.Settable)
	set(11, m.Threshold.UNC.Settable)
	set(10, m.Threshold.LNR.Settable)
	set(9, m.Threshold.LCR.Settable)
	set(8, m.Threshold.LNC.Settable)

	set(5, m.Threshold.UNR.Readable)
	set(4, m.Threshold.UCR.Readable)
	set(3, m.Threshold.UNC.Readable)
	set(2, m.Threshold.LNR.Readable)
	set(1, m.Threshold.LCR.Readable)
	set(0, m.Threshold.LNC.Readable)
	return b
}

func packSignedNibble(v int8) uint8 {
	return uint8(TwoSComplementEncode(int32(v), 4)) & 0x0f
}

func pack10BitSigned(v int16) (lo, hi uint8) {
	u := uint16(TwoSComplementEncode(int32(v), 10))
	lo = byte(u & 0xff)
	hi = byte((u >> 8) & 0x03)
	return lo, hi
}

func packChannelInfo(c ChannelInfo) uint8 {
	var b uint8
	if c.TransmitSupported {
		b = SetBit7(b)
	}
	b |= (c.MessageReceiveLUN & 0x07) << 4
	b |= c.ChannelProtocol & 0x0f
	return b
}

func packIDField(tl TypeLength, raw []byte) []byte {
	if tl == 0 && len(raw) == 0 {
		return []byte{0xC0}
	}
	if tl == 0 && len(raw) > 0 {
		return packASCIITypeLengthField(string(raw))
	}
	out := make([]byte, 1+len(raw))
	out[0] = byte(tl)
	copy(out[1:], raw)
	return out
}

func packReadingFactorsBytes(rf ReadingFactors, linearization LinearizationFunc, sensorDirection uint8) (b23, b24, b25, b26, b27, b28, b29 uint8) {
	b23 = uint8(linearization)
	mLo, mHi := pack10BitSigned(rf.M)
	b24 = mLo
	b25 = mHi<<6 | (rf.Tolerance & 0x3f)

	bLo, bHi := pack10BitSigned(rf.B)
	b26 = bLo
	b27 = bHi<<6 | uint8(rf.Accuracy&0x3f)
	b28 = uint8((rf.Accuracy>>6)&0x0f)<<4 | (rf.Accuracy_Exp&0x03)<<2 | (sensorDirection & 0x03)
	b29 = packSignedNibble(rf.R_Exp)<<4 | packSignedNibble(rf.B_Exp)
	return b23, b24, b25, b26, b27, b28, b29
}

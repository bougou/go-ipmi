package types

import (
	"strings"
	"testing"
)

// TestSensorSpecificEventNames checks event names against IPMI v2.0
// Table 42-3, Sensor Type Codes.
func TestSensorSpecificEventNames(t *testing.T) {
	tests := []struct {
		name       string
		sensorType SensorType
		offset     uint8
		want       string
	}{
		{
			name:       "power unit power off",
			sensorType: SensorTypePowerUnit,
			offset:     0x00,
			want:       "Power Off / Power Down",
		},
		{
			name:       "entity presence disabled",
			sensorType: SensorTypeEntityPresence,
			offset:     0x02,
			want:       "Entity Disabled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := EventReadingTypeSensorSpecific.EventForOffset(tt.sensorType, tt.offset)
			if event == nil {
				t.Fatalf("no event for sensor type %#02x offset %#02x", uint8(tt.sensorType), tt.offset)
			}
			if event.EventName != tt.want {
				t.Errorf("EventName = %q, want %q", event.EventName, tt.want)
			}
		})
	}
}

func TestEventNamesHaveNoSurroundingSpace(t *testing.T) {
	check := func(table string, key any, offset uint8, name string) {
		if name != strings.TrimSpace(name) {
			t.Errorf("%s[%v][%#02x]: EventName %q has surrounding space", table, key, offset, name)
		}
	}
	for typ, events := range GenericEvents {
		for offset, event := range events {
			check("GenericEvents", typ, offset, event.EventName)
		}
	}
	for sensorType, events := range SensorSpecificEvents {
		for offset, event := range events {
			check("SensorSpecificEvents", sensorType, offset, event.EventName)
		}
	}
}

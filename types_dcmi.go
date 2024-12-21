package ipmi

// [DCMI specification v1.5]: 6.6.3 Set Power Limit
// Exception Actions, taken if the Power Limit is exceeded and cannot be controlled within the Correction Time Limit
type DCMIExceptionAction uint8

const (
	DCMIExceptionAction_NoAction          DCMIExceptionAction = 0x00
	DCMIExceptionAction_PowerOffAndLogSEL DCMIExceptionAction = 0x01
	DCMIExceptionAction_LogSEL            DCMIExceptionAction = 0x11
)

func (a DCMIExceptionAction) String() string {
	m := map[DCMIExceptionAction]string{
		0x00: "No Action",
		0x01: "Hard Power Off & Log Event to SEL",
		0x11: "Log Event to SEL",
	}
	s, ok := m[a]
	if ok {
		return s
	}
	return "unknown"
}

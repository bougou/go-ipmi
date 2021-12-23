package ipmi

type GetChassisStatusRequest struct {
	// no data
}

type GetChassisStatusResponse struct {
	CompletionCode

	// Current Power State
	// [7] - reserved
	// [6:5] - power restore policy[1]
	// 00b = chassis stays powered off after AC/mains returns
	// 01b = after AC returns, power is restored to the state that was in effect
	// when AC/mains was lost
	// 10b = chassis always powers up after AC/mains returns
	// 11b = unknown
	// [4] - power control fault
	// 1b = Controller attempted to turn system power on or off, but system
	// did not enter desired state.
	// [3] - power fault
	// 1b = fault detected in main power subsystem.
	// [2] - 1b = Interlock (chassis is presently shut down because a chassis
	// panel interlock switch is active). (IPMI 1.5)
	// [1] - Power overload
	// 1b = system shutdown because of power overload condition.
	// [0] - Power is on
	// 1b = system power is on
	// 0b = system power is off (soft-off S4/S5 or mechanical off)
	PowerRestorePolicy uint8
	PowerControlFault  bool
	InterLock          bool
	PowerOverload      bool
	PowerIsOn          bool
}

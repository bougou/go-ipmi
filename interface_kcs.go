package ipmi

// 9.2 KCS Interface-BMC Request Message Format
type KCSRequest struct {
	// The NetFn field occupies the most significant six bits of the first message byte.
	NetFn NetFn

	// The LUN field occupies the least significant two bits of the first message byte.
	LUN uint8

	Command uint8

	Data []byte
}

// 9.3 BMC-KCS Interface Response Message Format
type KCSResponse struct {
	// This is a return of the NetFn code that was passed in the Request Message. Except that an odd NetFn value is returned.
	NetFn NetFn

	// This is a return of the LUN that was passed in the Request Message.
	LUN uint8

	Command uint8

	// The Completion Code indicates whether the request completed successfully or not.
	CompletionCode

	Data []byte
}

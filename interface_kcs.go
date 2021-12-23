package ipmi

type KCSRequest struct {
	// The NetFn field occupies the most significant six bits of the first message byte.
	NetFn

	// The LUN field occupies the least significant two bits of the first message byte.
	LUN

	Command

	Data []byte
}

type KCSResponse struct {
	// This is a return of the NetFn code that was passed in the Request Message. Except that an odd NetFn value is returned.
	NetFn

	// This is a return of the LUN that was passed in the Request Message.
	LUN

	Command

	// The Completion Code indicates whether the request completed successfully or not.
	CompletionCode

	Data []byte
}

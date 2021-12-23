package ipmi

// 11.1 BT Interface-BMC Request Message Format
type BT struct {
	// This is not actually part of the message, but part of the framing for the BT Interface. This value
	// is the 1-based count of message bytes following the length byte. The minimum length byte
	// value for a command to the BMC would be 3 to cover the NetFn/LUN, Seq, and Cmd bytes
	Length uint8

	// Network Function code. This provides the first level of functional routing for messages received
	// by the BMC via the BT Interface. The NetFn field occupies the most significant six bits of the first message byte
	NetFn
	// Logical Unit Number. This is a sub-address that allows messages to be routed to different
	// 'logical units' that reside behind the same physical interface. The LUN field occupies the least significant two bits of the first message byte.
	LUN

	// Used for matching responses up with requests.
	Sequence uint8

	// Command code. This message byte specifies the operation that is to be executed under the
	// specified Network Function.
	Command

	// Zero or more bytes of data, as required by the given command. The general convention is to
	// pass data LS-byte first, but check the individual command specifications to be sure.
	Data []byte
}

type BTResponse struct {
	Length uint8

	NetFn
	LUN

	Sequence uint8

	Command

	CompletionCode

	Data []byte
}

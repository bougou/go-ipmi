package ipmi

type GeMessageRequest struct {
}

type GetMessageResponse struct {
	NetFn

	LUN

	Command

	CompletionCode

	Channel
}

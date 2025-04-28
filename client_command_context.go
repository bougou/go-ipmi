package ipmi

import "context"

// CommandContext represents the specific context used for IPMI command exchanges.
// It stores addressing information needed for certain IPMI commands, particularly
// sensor-related operations.
//
// The context contains:
// - responderAddr: The address of the responding device (e.g., BMC)
// - responderLUN: The Logical Unit Number of the responding device
// - requesterAddr: The address of the requesting device
// - requesterLUN: The Logical Unit Number of the requesting device
//
// This context is essential for commands that require specific addressing information,
// such as GetSensorReading and other sensor-related operations.
type CommandContext struct {
	responderAddr *uint8
	responderLUN  *uint8
	requesterAddr *uint8
	requesterLUN  *uint8
}

func (cmdCtx *CommandContext) WithResponderAddr(responderAddr uint8) *CommandContext {
	cmdCtx.responderAddr = &responderAddr
	return cmdCtx
}

func (cmdCtx *CommandContext) WithResponderLUN(responderLUN uint8) *CommandContext {
	cmdCtx.responderLUN = &responderLUN
	return cmdCtx
}

func (cmdCtx *CommandContext) WithRequesterAddr(requesterAddr uint8) *CommandContext {
	cmdCtx.requesterAddr = &requesterAddr
	return cmdCtx
}

func (cmdCtx *CommandContext) WithRequesterLUN(requesterLUN uint8) *CommandContext {
	cmdCtx.requesterLUN = &requesterLUN
	return cmdCtx
}

// commandContextKeyType is a custom type for the context key to avoid collisions
type commandContextKeyType string

// commandContextKey is the key used to store CommandContext in the context.Context
const commandContextKey commandContextKeyType = "CommandContext"

// WithCommandContext adds a CommandContext to the provided context.Context.
func WithCommandContext(ctx context.Context, commandContext *CommandContext) context.Context {
	return context.WithValue(ctx, commandContextKey, commandContext)
}

// GetCommandContext retrieves the CommandContext from the provided context.Context.
func GetCommandContext(ctx context.Context) *CommandContext {
	if ctx.Value(commandContextKey) == nil {
		return nil
	}
	return ctx.Value(commandContextKey).(*CommandContext)
}

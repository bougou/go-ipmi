package handlers

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// Handler processes a single IPMI command.
//
// reqData is the raw request body after the IPMI header has been stripped.
// Implementations must not hold references to reqData after returning.
//
// Returned values:
//   - respData: raw response body; may be nil on error.
//   - cc: IPMI completion code; [types.CodeOK] on success.
//   - err: non-nil only for transport-level or programming errors, not IPMI
//     completion-code errors.  When err != nil the server returns [types.CodeUnspecifiedError].
type Handler interface {
	Handle(ctx context.Context, hctx *HandlerContext, reqData []byte) (respData []byte, cc types.CompletionCode, err error)
}

// HandlerFunc adapts a plain function to [Handler].
type HandlerFunc func(ctx context.Context, hctx *HandlerContext, reqData []byte) ([]byte, types.CompletionCode, error)

func (f HandlerFunc) Handle(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	return f(ctx, hctx, data)
}

// Middleware wraps a [Handler] to add cross-cutting behaviour (logging, auth, metrics).
type Middleware func(Handler) Handler

// commandKey is the dispatch key: NetFn (high byte) | Cmd (low byte).
// We use uint16 to avoid allocating a struct as a map key on every lookup.
type commandKey uint16

func makeKey(netFn, cmd uint8) commandKey {
	// NetFn is even for requests; the registry stores the request NetFn.
	return commandKey(uint16(netFn)<<8 | uint16(cmd))
}

// Registry maps (NetFn, Cmd) pairs to [Handler] implementations.
type Registry struct {
	handlers   map[commandKey]Handler
	middleware []Middleware
}

// NewRegistry returns an empty [Registry].
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[commandKey]Handler),
	}
}

// Register adds or replaces the handler for (netFn, cmd).
// netFn should be the *request* NetFn (even value).
//
// The handler is wrapped with a privilege check (innermost) and then any
// registered middleware so that middleware can observe privilege rejections
// (e.g. audit logging). Both wrappings happen at registration time to avoid
// per-dispatch allocations.
func (r *Registry) Register(netFn, cmd uint8, h Handler) {
	checked := &dispatchingHandler{inner: h, netFn: netFn, cmd: cmd}
	r.handlers[makeKey(netFn, cmd)] = r.applyMiddleware(checked)
}

// RegisterFunc is a convenience wrapper around [Register] for plain functions.
func (r *Registry) RegisterFunc(netFn, cmd uint8, fn func(context.Context, *HandlerContext, []byte) ([]byte, types.CompletionCode, error)) {
	r.Register(netFn, cmd, HandlerFunc(fn))
}

// Use appends middleware.  Middleware is applied in the order it was added,
// so the first Use() call produces the outermost wrapper.
// Middleware registered after [Register] calls is NOT retroactively applied;
// call Use() before Register() or re-register the handlers afterwards.
func (r *Registry) Use(m Middleware) {
	r.middleware = append(r.middleware, m)
}

// Merge copies all handlers from other into r.  Handlers in other overwrite
// those in r when they share the same (netFn, cmd) key.
func (r *Registry) Merge(other *Registry) {
	for k, h := range other.handlers {
		r.handlers[k] = h
	}
}

// Dispatch looks up and calls the handler for (netFn, cmd).
// Privilege checking and middleware were applied at registration time
// ([Register]), so this is a simple lookup with no per-call wrapping.
// Returns [types.CodeInvalidCommand] when no handler is registered.
func (r *Registry) Dispatch(ctx context.Context, hctx *HandlerContext, netFn, cmd uint8, data []byte) ([]byte, types.CompletionCode, error) {
	h, ok := r.handlers[makeKey(netFn, cmd)]
	if !ok {
		return nil, types.CodeInvalidCommand, fmt.Errorf("no handler for netFn=0x%02x cmd=0x%02x", netFn, cmd)
	}
	return h.Handle(ctx, hctx, data)
}

// applyMiddleware wraps h with all currently registered middleware.
func (r *Registry) applyMiddleware(h Handler) Handler {
	// Apply in reverse so the first-registered middleware is outermost.
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}
	return h
}

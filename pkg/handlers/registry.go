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
	commands   map[commandKey]types.Command
	middleware []Middleware
}

// NewRegistry returns an empty [Registry].
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[commandKey]Handler),
		commands: make(map[commandKey]types.Command),
	}
}

// Register adds or replaces the handler for c, whose NetFn must be the
// *request* NetFn (even value).  c is remembered so that dispatching the
// command reports it through [HandlerContext.Command]; pass one of the
// [types] command-table entries to get its spec name, or a literal such as
// types.Command{ID: 0x01, NetFn: 0x2e, Name: "Acme Set Fan Curve"} for an OEM
// command with no table entry.
//
// The handler is wrapped with a privilege check (innermost) and then any
// registered middleware so that middleware can observe privilege rejections
// (e.g. audit logging). Both wrappings happen at registration time to avoid
// per-dispatch allocations.
func (r *Registry) Register(c types.Command, h Handler) {
	netFn, cmd := uint8(c.NetFn), c.ID
	checked := &dispatchingHandler{inner: h, netFn: netFn, cmd: cmd}
	key := makeKey(netFn, cmd)
	r.commands[key] = c
	r.handlers[key] = r.applyMiddleware(checked)
}

// RegisterFunc is a convenience wrapper around [Register] for plain functions.
func (r *Registry) RegisterFunc(c types.Command, fn func(context.Context, *HandlerContext, []byte) ([]byte, types.CompletionCode, error)) {
	r.Register(c, HandlerFunc(fn))
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
	for k, c := range other.commands {
		r.commands[k] = c
	}
}

// Dispatch identifies the request on hctx, then looks up and calls its handler.
// Privilege checking and middleware were applied at registration time
// ([Register]), so the found path is a simple lookup with no per-call wrapping.
//
// An unregistered command still runs the middleware chain before returning
// [types.CodeInvalidCommand]: an initiator probing for commands the BMC does not
// implement is exactly the kind of event middleware exists to observe, and
// short-circuiting here would make it invisible.  Having no registration to be
// wrapped at, that path builds its chain per call and therefore sees the
// middleware present now rather than the snapshot [Register] would have taken.
// Both differences are confined to a path that answers with an error.
func (r *Registry) Dispatch(ctx context.Context, hctx *HandlerContext, netFn, cmd uint8, data []byte) ([]byte, types.CompletionCode, error) {
	key := makeKey(netFn, cmd)
	if hctx != nil {
		hctx.Command = r.lookup(key, netFn, cmd)
	}

	h, ok := r.handlers[key]
	if !ok {
		return r.applyMiddleware(unsupportedHandler(netFn, cmd)).Handle(ctx, hctx, data)
	}
	return h.Handle(ctx, hctx, data)
}

// lookup returns the command registered under key, falling back to the raw
// NetFn/Cmd pair so callers always get the numbers even for commands the
// registry does not know by name.
func (r *Registry) lookup(key commandKey, netFn, cmd uint8) types.Command {
	if c, ok := r.commands[key]; ok {
		return c
	}
	return types.Command{ID: cmd, NetFn: types.NetFn(netFn)}
}

func unsupportedHandler(netFn, cmd uint8) Handler {
	return HandlerFunc(func(context.Context, *HandlerContext, []byte) ([]byte, types.CompletionCode, error) {
		return nil, types.CodeInvalidCommand, fmt.Errorf("no handler for netFn=0x%02x cmd=0x%02x", netFn, cmd)
	})
}

// applyMiddleware wraps h with all currently registered middleware.
func (r *Registry) applyMiddleware(h Handler) Handler {
	// Apply in reverse so the first-registered middleware is outermost.
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}
	return h
}

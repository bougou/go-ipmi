//go:build windows
// +build windows

package open

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Windows transports for the Open Interface.
//
// On Windows there is no OpenIPMI character device. Local (in-band) access to
// the BMC is provided by the Microsoft IPMI driver (ipmidrv.sys), which is
// surfaced through the Microsoft_IPMI WMI class in the root\wmi namespace.
// Two Backend implementations are provided and selectable at runtime through
// ResolveBackend (the client layer exposes the choice via its open-backend
// preference):
//
//   - COMBackend (wmi-com, this file): native COM/OLE dispatch via go-ole
//     (syscall-based, no cgo — CGO_ENABLED=0 cross-compiles work).
//     In-process; the per-request cost is dominated by the BMC/KCS round
//     trip (~50 ms measured on a live BMC), with negligible transport
//     overhead. Requires administrator rights.
//   - PowerShellBackend (wmi-ps, see wmi_ps_windows.go): drives the same WMI
//     provider through a PowerShell helper (PowerShellConn). Zero
//     dependencies beyond powershell.exe on PATH, but each request spawns a
//     fresh process — measured 0.4–0.6 s per request end-to-end. A
//     fallback/diagnostic transport, not for bulk work.
//
// The BackendAuto preference ("auto") tries COMBackend first and falls back
// to PowerShellBackend if COM initialisation fails (e.g. on hosts where COM
// activation itself is blocked — endpoint-protection rules against
// SWbemLocator/DCOM, locked-down DCOM security settings, or broken COM
// registration). Note the root\wmi namespace ACL applies to both transports
// equally, so a permission failure there is not a fallback trigger. Pin
// BackendCOM or BackendPowerShell to force a single backend.
//
// The COM transport (WindowsConn + COMBackend)
//
// The standard contract is the Microsoft_IPMI class in the ROOT\WMI
// namespace, whose RequestResponse method tunnels a single IPMI request to
// the BMC and returns the response. We talk to it through COM/OLE using
// go-ole.
//
// Preconditions:
//   - Windows Server 2003 R2 or later (Vista+ on client SKUs).
//   - The "Microsoft Generic IPMI Compliant Device" must be present in Device
//     Manager (loaded automatically when the platform exposes a KCS/BT/SSIF
//     interface to ACPI).
//   - The process needs administrator rights (WMI ACLs on Microsoft_IPMI).

// WindowsConn represents an established WMI connection to the Microsoft_IPMI
// provider. All COM dispatch is funneled through a single dedicated worker
// goroutine (started in OpenWindowsConn) that owns an OS thread and a COM
// apartment for the lifetime of the connection. This amortizes LockOSThread
// + CoInitializeEx/CoUninitialize across every request the connection will
// ever issue — for a 14k-entry SEL stream that's ~14k cycles avoided, saving
// hundreds of milliseconds per call versus spawning a fresh apartment per
// request.
type WindowsConn struct {
	unknown  *ole.IUnknown // SWbemLocator IUnknown (owns the COM apartment)
	locator  *ole.IDispatch
	service  *ole.IDispatch // SWbemServices for ROOT\WMI
	instance *ole.IDispatch // Microsoft_IPMI singleton instance

	// reqInClass is the class object representing the InParameters
	// definition of the RequestResponse method. We hold it so we can spawn
	// fresh inParam instances for every call without repeated reflection.
	reqInClass *ole.IDispatch

	// reqCh feeds the worker; workerDone is closed when the worker exits
	// (normally on Close, or if CoInitializeEx fails on the worker thread).
	// Callers select on workerDone to detect mid-stream shutdown.
	reqCh      chan *workReq
	workerDone chan struct{}
}

// workReq is a single queued IPMI request awaiting COM dispatch.
// Fields needed after enqueue are owned copies so the caller may return
// (and the GC reclaim the original Request / payload) on timeout without
// racing the worker.
type workReq struct {
	netfn         uint8
	lun           uint8
	responderAddr uint8
	cmd           uint8
	reqData       []byte
	reply         chan result
}

// result is the worker's reply for a single workReq.
type result struct {
	data []byte
	err  error
}

// OpenWindowsConn establishes a WMI connection to ROOT\WMI, resolves the
// Microsoft_IPMI singleton instance, and caches the RequestResponse in-param
// class for later spawning. The caller must invoke Close when done.
func OpenWindowsConn() (*WindowsConn, error) {
	// CoInitializeEx is idempotent per-goroutine; an "already initialised"
	// error with a different threading model is acceptable.
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		oleErr, ok := err.(*ole.OleError)
		// RPC_E_CHANGED_MODE (0x80010106) — apartment already up under a
		// different model. Tolerated.
		if !ok || uintptr(oleErr.Code()) != 0x80010106 {
			return nil, fmt.Errorf("CoInitializeEx failed: %w", err)
		}
	}

	conn := &WindowsConn{}
	success := false
	defer func() {
		if !success {
			conn.Close()
		}
	}()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return nil, fmt.Errorf("create SWbemLocator failed: %w", err)
	}
	conn.unknown = unknown

	locator, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("query IDispatch on SWbemLocator failed: %w", err)
	}
	conn.locator = locator

	serviceRaw, err := oleutil.CallMethod(locator, "ConnectServer", nil, `root\wmi`)
	if err != nil {
		return nil, fmt.Errorf("ConnectServer(root\\wmi) failed: %w", err)
	}
	conn.service = serviceRaw.ToIDispatch()

	// Pull the singleton Microsoft_IPMI instance. WMI exposes a "=@"
	// singleton selector for this class.
	instRaw, err := oleutil.CallMethod(conn.service, "Get", "Microsoft_IPMI=@")
	if err != nil {
		// Fall back to InstancesOf if the singleton selector is rejected
		// by some driver versions.
		inst, ferr := conn.firstInstance()
		if ferr != nil {
			return nil, fmt.Errorf("locate Microsoft_IPMI instance failed (Get: %v; InstancesOf: %v)", err, ferr)
		}
		conn.instance = inst
	} else {
		conn.instance = instRaw.ToIDispatch()
	}

	// Cache the RequestResponse in-param class for spawning.
	methodsRaw, err := oleutil.GetProperty(conn.instance, "Methods_")
	if err != nil {
		return nil, fmt.Errorf("read Methods_ failed: %w", err)
	}
	methods := methodsRaw.ToIDispatch()
	defer methods.Release()

	methodRaw, err := oleutil.CallMethod(methods, "Item", "RequestResponse")
	if err != nil {
		return nil, fmt.Errorf("locate RequestResponse method failed: %w", err)
	}
	method := methodRaw.ToIDispatch()
	defer method.Release()

	inParamsRaw, err := oleutil.GetProperty(method, "InParameters")
	if err != nil {
		return nil, fmt.Errorf("read RequestResponse InParameters failed: %w", err)
	}
	conn.reqInClass = inParamsRaw.ToIDispatch()

	// Start the dedicated COM worker. It owns its OS thread + COM apartment
	// for the lifetime of the connection. If startup fails the deferred
	// Close above will signal the worker to drain and exit.
	conn.reqCh = make(chan *workReq)
	conn.workerDone = make(chan struct{})
	go conn.worker()

	success = true
	return conn, nil
}

// firstInstance walks InstancesOf("Microsoft_IPMI") and returns the first hit.
// Only used as a fallback when the singleton Get selector is unavailable.
func (c *WindowsConn) firstInstance() (*ole.IDispatch, error) {
	instancesRaw, err := oleutil.CallMethod(c.service, "InstancesOf", "Microsoft_IPMI")
	if err != nil {
		return nil, err
	}
	instances := instancesRaw.ToIDispatch()
	defer instances.Release()

	enumRaw, err := instances.GetProperty("_NewEnum")
	if err != nil {
		return nil, fmt.Errorf("_NewEnum failed: %w", err)
	}
	enumUnk := enumRaw.ToIUnknown()
	defer enumUnk.Release()

	enum, err := enumUnk.IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return nil, fmt.Errorf("IEnumVARIANT cast failed: %w", err)
	}
	defer enum.Release()

	itemRaw, length, err := enum.Next(1)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, fmt.Errorf("no Microsoft_IPMI instance found; is the BMC visible to Windows?")
	}
	return itemRaw.ToIDispatch(), nil
}

// Close releases all COM resources held by the connection. Safe to call more
// than once and on a partially-constructed connection.
func (c *WindowsConn) Close() error {
	if c == nil {
		return nil
	}
	// Signal the worker to drain and exit. Skip if never started.
	if c.reqCh != nil {
		close(c.reqCh)
		<-c.workerDone
		c.reqCh = nil
		c.workerDone = nil
	}
	if c.reqInClass != nil {
		c.reqInClass.Release()
		c.reqInClass = nil
	}
	if c.instance != nil {
		c.instance.Release()
		c.instance = nil
	}
	if c.service != nil {
		c.service.Release()
		c.service = nil
	}
	if c.locator != nil {
		c.locator.Release()
		c.locator = nil
	}
	if c.unknown != nil {
		c.unknown.Release()
		c.unknown = nil
	}
	// We intentionally skip ole.CoUninitialize on the caller's thread — other
	// COM consumers in the same goroutine may still need the apartment, and
	// uniniting on every Close has historically destabilised go-ole. The
	// worker CoUninitializes on its own thread as part of its shutdown.
	return nil
}

// SendCommand sends a single IPMI request via Microsoft_IPMI::RequestResponse
// and returns the response payload in the exact same shape as the Linux
// backend: a byte slice whose first byte is the IPMI completion code,
// followed by the response data.
//
// MsgID is ignored — the WMI provider does not surface a sequence number.
//
// ctx cancellation and timeout both abandon the wait: the caller's select
// returns while the worker continues the in-flight WMI dispatch. When that
// dispatch completes the worker writes to the buffered reply channel (size 1),
// where the result is garbage-collected with the workReq. The next request
// queues behind the still-running call and is served when the worker returns
// to the for-loop.
//
// Request fields and payload are copied into workReq before enqueue so an
// abandoned caller's Request / payload can be reclaimed safely.
//
// Concurrent SendCommand callers serialize naturally on the worker: only
// one WMI dispatch is ever in flight, matching the per-call mutex behavior
// of the previous implementation without paying LockOSThread + CoInit per
// call.
func (c *WindowsConn) SendCommand(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	if c == nil || c.instance == nil {
		return nil, fmt.Errorf("WindowsConn is not initialised")
	}
	if req == nil {
		return nil, fmt.Errorf("nil open request")
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// Copy: the worker may outlive the caller's req/Data on timeout.
	reqData := append([]byte(nil), req.Data...)

	work := &workReq{
		netfn:         req.NetFn,
		lun:           req.LUN,
		responderAddr: req.EffectiveTarget(),
		cmd:           req.Cmd,
		reqData:       reqData,
		reply:         make(chan result, 1), // buffered: abandoned replies must not block the worker
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case c.reqCh <- work:
	case <-c.workerDone:
		return nil, fmt.Errorf("COM worker has exited")
	case <-ctx.Done():
		return nil, fmt.Errorf("WMI request canceled: %w", ctx.Err())
	case <-timer.C:
		return nil, fmt.Errorf("WMI request enqueue timed out after %s", timeout)
	}

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(timeout)

	select {
	case r := <-work.reply:
		return r.data, r.err
	case <-c.workerDone:
		return nil, fmt.Errorf("COM worker exited mid-request")
	case <-ctx.Done():
		return nil, fmt.Errorf("WMI request canceled: %w", ctx.Err())
	case <-timer.C:
		// Worker still busy. Caller gives up; the worker will deliver to the
		// buffered reply channel when the WMI call returns.
		return nil, fmt.Errorf("WMI Microsoft_IPMI::RequestResponse timed out after %s", timeout)
	}
}

// worker is the long-lived COM dispatch loop. It owns one OS thread and one
// MTA apartment for the lifetime of the connection. Per MSDN, CoInitializeEx
// returns S_OK on first init and S_FALSE if the thread was already
// initialised with the same concurrency model — both are success and both
// must be balanced by CoUninitialize. RPC_E_CHANGED_MODE (0x80010106) is
// the only real failure. go-ole wraps any non-S_OK HRESULT as an OleError,
// so we have to unpack the code.
func (c *WindowsConn) worker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		oleErr, ok := err.(*ole.OleError)
		if !ok || uintptr(oleErr.Code()) != 1 {
			// Real init failure. Drain the queue with the error so callers
			// don't hang waiting for a worker that will never service a
			// request. The reqCh sender side is closed by Close.
			for work := range c.reqCh {
				work.reply <- result{err: fmt.Errorf("COM worker CoInitializeEx: %w", err)}
			}
			close(c.workerDone)
			return
		}
		// S_FALSE (code 1) — thread was already in the MTA, OK to proceed.
	}
	defer ole.CoUninitialize()
	defer close(c.workerDone)

	for work := range c.reqCh {
		func() {
			defer func() {
				if r := recover(); r != nil {
					work.reply <- result{err: fmt.Errorf("worker panic: %v", r)}
				}
			}()
			data, err := c.invokeRequestResponse(
				work.netfn,
				work.lun,
				work.responderAddr,
				work.cmd,
				work.reqData,
			)
			work.reply <- result{data: data, err: err}
		}()
	}
}

// invokeRequestResponse performs a single WMI method invocation. Must be
// called from the worker goroutine so that the COM apartment is live for
// the duration of the dispatch.
func (c *WindowsConn) invokeRequestResponse(netfn, lun, responder, cmd uint8, data []byte) (resp []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			resp = nil
			err = fmt.Errorf("invokeRequestResponse panic: %v", r)
		}
	}()

	inRaw, err := oleutil.CallMethod(c.reqInClass, "SpawnInstance_")
	if err != nil {
		return nil, fmt.Errorf("SpawnInstance_ failed: %w", err)
	}
	in := inRaw.ToIDispatch()
	defer in.Release()

	if _, err := oleutil.PutProperty(in, "Command", int32(cmd)); err != nil {
		return nil, fmt.Errorf("set Command: %w", err)
	}
	if _, err := oleutil.PutProperty(in, "Lun", int32(lun)); err != nil {
		return nil, fmt.Errorf("set Lun: %w", err)
	}
	if _, err := oleutil.PutProperty(in, "NetworkFunction", int32(netfn)); err != nil {
		return nil, fmt.Errorf("set NetworkFunction: %w", err)
	}
	if _, err := oleutil.PutProperty(in, "ResponderAddress", int32(responder)); err != nil {
		return nil, fmt.Errorf("set ResponderAddress: %w", err)
	}

	if data == nil {
		data = []byte{}
	}
	if _, err := oleutil.PutProperty(in, "RequestData", data); err != nil {
		return nil, fmt.Errorf("set RequestData: %w", err)
	}
	if _, err := oleutil.PutProperty(in, "RequestDataSize", uint32(len(data))); err != nil {
		return nil, fmt.Errorf("set RequestDataSize: %w", err)
	}

	outRaw, err := oleutil.CallMethod(c.instance, "ExecMethod_", "RequestResponse", in)
	if err != nil {
		return nil, fmt.Errorf("ExecMethod_ RequestResponse failed: %w", err)
	}
	defer outRaw.Clear()

	out := outRaw.ToIDispatch()
	if out == nil {
		return nil, fmt.Errorf("ExecMethod_ returned no output object")
	}
	// Do NOT defer out.Release(). out points to the same IDispatch held by
	// outRaw; outRaw.Clear() above releases it. Releasing out as well would
	// double-free and corrupt the heap.

	// The Microsoft_IPMI provider returns the FULL IPMI response — including
	// the leading completion code byte — in ResponseData. (ResponseDataSize
	// for a Get Device ID response is 16, which is 1 cc + 15 payload bytes,
	// not 15.) The CompletionCode property mirrors ResponseData[0]; reading
	// it separately and prepending shifts every downstream field by one and
	// silently corrupts the parse. Trust ResponseData and return it as-is.
	respVar, err := oleutil.GetProperty(out, "ResponseData")
	if err != nil {
		return nil, fmt.Errorf("read ResponseData: %w", err)
	}
	defer respVar.Clear()

	respBytes, err := variantUI1ArrayToBytes(respVar)
	if err != nil {
		return nil, fmt.Errorf("decode ResponseData: %w", err)
	}

	if szVar, err := oleutil.GetProperty(out, "ResponseDataSize"); err == nil {
		defer szVar.Clear()
		want := int(uint32(szVar.Val))
		if want >= 0 && want <= len(respBytes) {
			respBytes = respBytes[:want]
		}
	}

	if len(respBytes) == 0 {
		return nil, fmt.Errorf("ResponseData empty (no completion code)")
	}
	return respBytes, nil
}

// variantUI1ArrayToBytes extracts a byte slice from a VARIANT that is expected
// to contain a SAFEARRAY of VT_UI1. Returns an empty slice if the VARIANT is
// VT_NULL/VT_EMPTY.
func variantUI1ArrayToBytes(v *ole.VARIANT) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	switch v.VT {
	case ole.VT_NULL, ole.VT_EMPTY:
		return nil, nil
	}
	if v.VT&ole.VT_ARRAY == 0 {
		return nil, fmt.Errorf("VARIANT is not an array (VT=0x%x)", v.VT)
	}

	sac := v.ToArray()
	if sac == nil {
		return nil, fmt.Errorf("VARIANT.ToArray() returned nil")
	}
	// NOTE: do NOT defer sac.Release(). The SAFEARRAY is owned by the VARIANT;
	// the caller will VariantClear the owning VARIANT (respVar.Clear()), which
	// destroys the SAFEARRAY. Calling SafeArrayDestroy here as well would
	// double-free and corrupt the heap.

	// ToValueArray gives []interface{} where each element is the VT_UI1
	// value as a uint8 (boxed in interface{}).
	values := sac.ToValueArray()
	out := make([]byte, len(values))
	for i, e := range values {
		switch x := e.(type) {
		case uint8:
			out[i] = x
		case int8:
			out[i] = uint8(x)
		case int32:
			out[i] = uint8(x)
		case uint32:
			out[i] = uint8(x)
		case int64:
			out[i] = uint8(x)
		default:
			return nil, fmt.Errorf("unexpected SAFEARRAY element type %T at index %d", e, i)
		}
	}
	return out, nil
}

// COMBackend implements Backend by talking to the Microsoft_IPMI WMI
// provider via native COM (in-process, go-ole). It wraps WindowsConn.
type COMBackend struct {
	conn *WindowsConn
}

var _ Backend = (*COMBackend)(nil)

// Connect opens the WMI connection. The devnum is ignored: Windows exposes a
// single Microsoft_IPMI instance regardless of how many BMCs are present.
func (b *COMBackend) Connect(ctx context.Context, devnum int32) error {
	conn, err := OpenWindowsConn()
	if err != nil {
		return fmt.Errorf("open WMI Microsoft_IPMI provider failed, err: %w", err)
	}
	b.conn = conn
	return nil
}

func (b *COMBackend) Close(ctx context.Context) error {
	if b.conn == nil {
		return nil
	}
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("close WMI connection failed, err: %w", err)
	}
	b.conn = nil
	return nil
}

func (b *COMBackend) Send(ctx context.Context, req *Request, timeout time.Duration) ([]byte, error) {
	if b.conn == nil {
		return nil, fmt.Errorf("wmi-com backend not connected")
	}
	return b.conn.SendCommand(ctx, req, timeout)
}

// ConnectCOMBackend constructs a COMBackend and connects it, returning the
// ready-to-use Backend. Intended as a factory callback for ResolveBackend.
func ConnectCOMBackend(ctx context.Context) (Backend, error) {
	b := &COMBackend{}
	if err := b.Connect(ctx, 0); err != nil {
		return nil, err
	}
	return b, nil
}

package serial

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type port struct {
	f  *os.File
	fd syscall.Handle
	rl sync.Mutex
	wl sync.Mutex
	ro *syscall.Overlapped
	wo *syscall.Overlapped
}
type dcb struct {
	DCBlength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XonLim, XoffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

type rwTimeouts struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

func open(address string, baudRate uint, dataBits uint8, stopBits StopBits, parity Parity, timeout time.Duration) (*port, error) {
	if len(strings.TrimSpace(address)) > 0 && address[0] != '\\' {
		address = "\\\\.\\" + address
	}
	name, err := syscall.UTF16PtrFromString(address)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(
		name,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL|syscall.FILE_FLAG_OVERLAPPED,
		0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(h), address)
	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	if err = setCommState(h, baudRate, dataBits, parity, stopBits); err != nil {
		return nil, err
	}
	if err = setupComm(h, 64, 64); err != nil {
		return nil, err
	}
	if err = setCommTimeouts(h, timeout); err != nil {
		return nil, err
	}
	if err = setCommMask(h); err != nil {
		return nil, err
	}

	ro, err := newOverlapped()
	if err != nil {
		return nil, err
	}
	wo, err := newOverlapped()
	if err != nil {
		return nil, err
	}
	return &port{
		f:  f,
		fd: h,
		ro: ro,
		wo: wo,
	}, nil
}

func (p *port) close() error {
	return p.f.Close()
}

func (p *port) write(buf []byte) (int, error) {
	p.wl.Lock()
	defer p.wl.Unlock()

	if err := resetEvent(p.wo.HEvent); err != nil {
		return 0, err
	}
	var n uint32
	err := syscall.WriteFile(p.fd, buf, &n, p.wo)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(n), err
	}
	return getOverlappedResult(p.fd, p.wo)
}

func (p *port) read(buf []byte) (int, error) {
	if p == nil || p.f == nil {
		return 0, fmt.Errorf("invalid port on read")
	}
	p.rl.Lock()
	defer p.rl.Unlock()

	if err := resetEvent(p.ro.HEvent); err != nil {
		return 0, err
	}
	var done uint32
	err := syscall.ReadFile(p.fd, buf, &done, p.ro)
	if err != nil && err != syscall.ERROR_IO_PENDING {
		return int(done), err
	}
	return getOverlappedResult(p.fd, p.ro)
}

// Discards data written to the port but not transmitted,
// or data received but not read
func (p *port) flush() error {
	return purgeComm(p.fd)
}

var (
	nSetCommState,
	nSetCommTimeouts,
	nSetCommMask,
	nSetupComm,
	nGetOverlappedResult,
	nCreateEvent,
	nResetEvent,
	nPurgeComm uintptr
	// nFlushFileBuffers uintptr
)

func init() {
	k32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		panic("LoadLibrary " + err.Error())
	}
	defer syscall.FreeLibrary(k32)

	nSetCommState = getProcAddr(k32, "SetCommState")
	nSetCommTimeouts = getProcAddr(k32, "SetCommTimeouts")
	nSetCommMask = getProcAddr(k32, "SetCommMask")
	nSetupComm = getProcAddr(k32, "SetupComm")
	nGetOverlappedResult = getProcAddr(k32, "GetOverlappedResult")
	nCreateEvent = getProcAddr(k32, "CreateEventW")
	nResetEvent = getProcAddr(k32, "ResetEvent")
	nPurgeComm = getProcAddr(k32, "PurgeComm")
	// nFlushFileBuffers = getProcAddr(k32, "FlushFileBuffers")
}

func getProcAddr(lib syscall.Handle, name string) uintptr {
	addr, err := syscall.GetProcAddress(lib, name)
	if err != nil {
		panic(name + " " + err.Error())
	}
	return addr
}

func setCommState(h syscall.Handle, baud uint, databits byte, parity Parity, stopbits StopBits) error {
	var params dcb
	params.DCBlength = uint32(unsafe.Sizeof(params))

	params.flags[0] = 0x01  // fBinary
	params.flags[0] |= 0x10 // Assert DSR

	params.BaudRate = uint32(baud)

	params.ByteSize = databits

	switch parity {
	case PARITY_NONE:
		params.Parity = 0
	case PARITY_ODD:
		params.Parity = 1
	case PARITY_EVEN:
		params.Parity = 2
	case PARITY_MARK:
		params.Parity = 3
	case PARITY_SPACE:
		params.Parity = 4
	default:
		return ErrBadParity
	}

	switch stopbits {
	case STOP_ONE:
		params.StopBits = 0
	case STOP_ONE_HALF:
		params.StopBits = 1
	case STOP_TWO:
		params.StopBits = 2
	default:
		return ErrBadStopBits
	}

	r, _, err := syscall.SyscallN(nSetCommState, uintptr(h), uintptr(unsafe.Pointer(&params)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setCommTimeouts(h syscall.Handle, readTimeout time.Duration) error {
	var timeouts rwTimeouts
	const MAXDWORD = 1<<32 - 1

	// blocking read by default
	var timeoutMs int64 = MAXDWORD - 1

	if readTimeout > 0 {
		// non-blocking read
		timeoutMs = readTimeout.Nanoseconds() / 1e6
		if timeoutMs < 1 {
			timeoutMs = 1
		} else if timeoutMs > MAXDWORD-1 {
			timeoutMs = MAXDWORD - 1
		}
	}

	/* From http://msdn.microsoft.com/en-us/library/aa363190(v=VS.85).aspx
		 For blocking I/O see below:
		 Remarks:
		 If an application sets ReadIntervalTimeout and
		 ReadTotalTimeoutMultiplier to MAXDWORD and sets
		 ReadTotalTimeoutConstant to a value greater than zero and
		 less than MAXDWORD, one of the following occurs when the
		 ReadFile function is called:
		 If there are any bytes in the input buffer, ReadFile returns
		       immediately with the bytes in the buffer.
		 If there are no bytes in the input buffer, ReadFile waits
	               until a byte arrives and then returns immediately.
		 If no bytes arrive within the time specified by
		       ReadTotalTimeoutConstant, ReadFile times out.
	*/

	timeouts.ReadIntervalTimeout = MAXDWORD
	timeouts.ReadTotalTimeoutMultiplier = MAXDWORD
	timeouts.ReadTotalTimeoutConstant = uint32(timeoutMs)
	r, _, err := syscall.SyscallN(nSetCommTimeouts, uintptr(h), uintptr(unsafe.Pointer(&timeouts)), 0)
	if r == 0 {
		return err
	}
	return nil
}

func setupComm(h syscall.Handle, in, out int) error {
	r, _, err := syscall.SyscallN(nSetupComm, uintptr(h), uintptr(in), uintptr(out))
	if r == 0 {
		return err
	}
	return nil
}

func setCommMask(h syscall.Handle) error {
	const EV_RXCHAR = 0x0001
	r, _, err := syscall.SyscallN(nSetCommMask, uintptr(h), EV_RXCHAR, 0)
	if r == 0 {
		return err
	}
	return nil
}

func resetEvent(h syscall.Handle) error {
	r, _, err := syscall.SyscallN(nResetEvent, uintptr(h), 0, 0)
	if r == 0 {
		return err
	}
	return nil
}

func purgeComm(h syscall.Handle) error {
	const PURGE_TXABORT = 0x0001
	const PURGE_RXABORT = 0x0002
	const PURGE_TXCLEAR = 0x0004
	const PURGE_RXCLEAR = 0x0008
	r, _, err := syscall.SyscallN(nPurgeComm, uintptr(h), PURGE_TXABORT|PURGE_RXABORT|PURGE_TXCLEAR|PURGE_RXCLEAR, 0)
	if r == 0 {
		return err
	}
	return nil
}

func newOverlapped() (*syscall.Overlapped, error) {
	var overlapped syscall.Overlapped
	r, _, err := syscall.SyscallN(nCreateEvent, 0, 1, 0, 0, 0, 0)
	if r == 0 {
		return nil, err
	}
	overlapped.HEvent = syscall.Handle(r)
	return &overlapped, nil
}

func getOverlappedResult(h syscall.Handle, overlapped *syscall.Overlapped) (int, error) {
	var n int
	r, _, err := syscall.SyscallN(nGetOverlappedResult,
		uintptr(h),
		uintptr(unsafe.Pointer(overlapped)),
		uintptr(unsafe.Pointer(&n)), 1, 0, 0)
	if r == 0 {
		return n, err
	}

	return n, nil
}

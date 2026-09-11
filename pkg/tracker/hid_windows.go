//go:build windows

package tracker

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

var (
	modCfgmgr32 = syscall.NewLazyDLL("cfgmgr32.dll")
	modHid      = syscall.NewLazyDLL("hid.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateEventW        = modKernel32.NewProc("CreateEventW")
	procGetOverlappedResult = modKernel32.NewProc("GetOverlappedResult")

	procCMGetDeviceInterfaceListSizeW = modCfgmgr32.NewProc("CM_Get_Device_Interface_List_SizeW")
	procCMGetDeviceInterfaceListW     = modCfgmgr32.NewProc("CM_Get_Device_Interface_ListW")

	procHidDGetAttributes         = modHid.NewProc("HidD_GetAttributes")
	procHidDGetManufacturerString = modHid.NewProc("HidD_GetManufacturerString")
	procHidDGetProductString      = modHid.NewProc("HidD_GetProductString")
	procHidDGetSerialNumberString = modHid.NewProc("HidD_GetSerialNumberString")
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var guidDevInterfaceHID = guid{
	0x4D1E55B2, 0xF16F, 0x11CF,
	[8]byte{0x88, 0xCB, 0x00, 0x11, 0x11, 0x00, 0x00, 0x30},
}

type hiddAttributes struct {
	Size          uint32
	VendorID      uint16
	ProductID     uint16
	VersionNumber uint16
}

const (
	crSuccess     = 0x00
	crBufferSmall = 0x1A

	cmGetDeviceInterfaceListPresent = 0x0

	hidStringBufLen = 256 // uint16 単位 (512 バイト)
)

func EnumerateHID() ([]HIDDevice, error) {
	paths, err := hidInterfacePaths()
	if err != nil {
		return nil, err
	}
	devs := make([]HIDDevice, 0, len(paths))
	for _, p := range paths {
		devs = append(devs, describeHID(p))
	}
	return devs, nil
}

func hidInterfacePaths() ([]string, error) {
	for {
		var size uint32
		r, _, _ := procCMGetDeviceInterfaceListSizeW.Call(
			uintptr(unsafe.Pointer(&size)),
			uintptr(unsafe.Pointer(&guidDevInterfaceHID)),
			0,
			cmGetDeviceInterfaceListPresent,
		)
		if r != crSuccess {
			return nil, fmt.Errorf("tracker: CM_Get_Device_Interface_List_SizeW failed: CR=0x%X", r)
		}
		if size == 0 {
			return nil, nil
		}
		buf := make([]uint16, size)
		r, _, _ = procCMGetDeviceInterfaceListW.Call(
			uintptr(unsafe.Pointer(&guidDevInterfaceHID)),
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(size),
			cmGetDeviceInterfaceListPresent,
		)
		if r == crBufferSmall {
			continue
		}
		if r != crSuccess {
			return nil, fmt.Errorf("tracker: CM_Get_Device_Interface_ListW failed: CR=0x%X", r)
		}
		return splitMultiSZ(buf), nil
	}
}

func splitMultiSZ(buf []uint16) []string {
	var out []string
	start := 0
	for i, c := range buf {
		if c != 0 {
			continue
		}
		if i == start {
			break // 二重 NUL (終端)
		}
		out = append(out, string(utf16.Decode(buf[start:i])))
		start = i + 1
	}
	return out
}

func describeHID(path string) HIDDevice {
	d := HIDDevice{Path: path}
	if vid, pid, ok := parseVIDPID(path); ok {
		d.VendorID, d.ProductID = vid, pid
	}

	p16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return d
	}

	h, err := syscall.CreateFile(
		p16,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return d
	}
	defer syscall.CloseHandle(h)

	attr := hiddAttributes{Size: uint32(unsafe.Sizeof(hiddAttributes{}))}
	if r, _, _ := procHidDGetAttributes.Call(uintptr(h), uintptr(unsafe.Pointer(&attr))); r != 0 {
		d.VendorID, d.ProductID = attr.VendorID, attr.ProductID
	}
	d.Manufacturer = hidString(procHidDGetManufacturerString, h)
	d.Product = hidString(procHidDGetProductString, h)
	d.SerialNumber = hidString(procHidDGetSerialNumberString, h)
	return d
}

func hidString(proc *syscall.LazyProc, h syscall.Handle) string {
	var buf [hidStringBufLen]uint16
	r, _, _ := proc.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)*2))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:])
}

func readInputReport(ctx context.Context, path string) (bool, error) {
	p16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	h, err := syscall.CreateFile(
		p16,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return false, fmt.Errorf("tracker: open %s: %w", path, err)
	}
	defer syscall.CloseHandle(h)

	ev, err := createEvent()
	if err != nil {
		return false, err
	}
	defer syscall.CloseHandle(ev)

	buf := make([]byte, 4096)
	ov := syscall.Overlapped{HEvent: ev}
	var n uint32
	err = syscall.ReadFile(h, buf, &n, &ov)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, syscall.ERROR_IO_PENDING) {
		return false, fmt.Errorf("tracker: read %s: %w", path, err)
	}

	const slice = 50 * time.Millisecond
	for {
		if ctx.Err() != nil {
			break
		}
		wait := slice
		if dl, ok := ctx.Deadline(); ok {
			if rem := time.Until(dl); rem < wait {
				wait = rem
			}
		}
		if wait <= 0 {
			break
		}
		ev, werr := syscall.WaitForSingleObject(ev, uint32(wait/time.Millisecond))
		switch {
		case werr != nil:
			syscall.CancelIoEx(h, &ov)
			return false, werr
		case ev == syscall.WAIT_OBJECT_0:
			if err := getOverlappedResult(h, &ov, &n, false); err != nil {
				return false, fmt.Errorf("tracker: read %s: %w", path, err)
			}
			return n > 0, nil
		case ev == uint32(syscall.WAIT_TIMEOUT):
		default:
			syscall.CancelIoEx(h, &ov)
			return false, fmt.Errorf("tracker: WaitForSingleObject returned %d", ev)
		}
	}

	syscall.CancelIoEx(h, &ov)
	_ = getOverlappedResult(h, &ov, &n, true)
	return false, nil
}

func createEvent() (syscall.Handle, error) {
	r, _, e := procCreateEventW.Call(0, 1, 0, 0)
	if r == 0 {
		if e != nil && e != syscall.Errno(0) {
			return 0, fmt.Errorf("tracker: CreateEventW: %w", e)
		}
		return 0, errors.New("tracker: CreateEventW failed")
	}
	return syscall.Handle(r), nil
}

func getOverlappedResult(h syscall.Handle, ov *syscall.Overlapped, n *uint32, wait bool) error {
	var w uintptr
	if wait {
		w = 1
	}
	r, _, e := procGetOverlappedResult.Call(uintptr(h), uintptr(unsafe.Pointer(ov)), uintptr(unsafe.Pointer(n)), w)
	if r == 0 {
		if e != nil && e != syscall.Errno(0) {
			return e
		}
		return errors.New("GetOverlappedResult failed")
	}
	return nil
}

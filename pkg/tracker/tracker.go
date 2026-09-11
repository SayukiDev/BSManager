package tracker

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var ErrUnsupported = errors.New("tracker: HID access is not supported on this platform")
var DefaultProbeTimeout = 500 * time.Millisecond

// ベンダー ID。
const (
	VendorValve uint16 = 0x28DE // Valve Corporation
	VendorHTC   uint16 = 0x0BB4 // HTC Corporation
)

type Kind int

const (
	KindWired Kind = iota + 1
	KindWireless
)

// String は Kind の表示名を返す。
func (k Kind) String() string {
	switch k {
	case KindWired:
		return "wired"
	case KindWireless:
		return "wireless"
	default:
		return "unknown"
	}
}

type HIDDevice struct {
	Path         string
	VendorID     uint16
	ProductID    uint16
	Manufacturer string
	Product      string
	SerialNumber string
}
type Device struct {
	HIDDevice
	Kind Kind
	Name string
}

// KnownDevice は既知の VID/PID。
type KnownDevice struct {
	VendorID  uint16
	ProductID uint16
	Name      string
	Dongle    bool
}

var KnownDevices = []KnownDevice{
	{VendorValve, 0x2022, "HTC Vive Tracker (2017)", false},
	{VendorValve, 0x2300, "HTC Vive Tracker (2018)", false},
	{VendorValve, 0x2301, "HTC Vive Tracker 3.0", false},
	{VendorValve, 0x2101, "Valve Watchman Dongle", true},
	{VendorValve, 0x2102, "Valve VR Radio", true},
}

type Status struct {
	Wired    []Device
	Wireless []Device
}

func (s Status) Any() bool { return len(s.Wired)+len(s.Wireless) > 0 }

func (s Status) Count() int { return len(s.Wired) + len(s.Wireless) }

func IsConnected() (bool, error) {
	st, err := Detect(context.Background())
	if err != nil {
		return false, err
	}
	return st.Any(), nil
}

func Detect(ctx context.Context) (Status, error) {
	devs, err := EnumerateHID()
	if err != nil {
		return Status{}, err
	}
	wired, dongles := classify(devs)

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultProbeTimeout)
		defer cancel()
	}
	wireless, err := probeDongles(ctx, dongles)
	if err != nil {
		return Status{}, err
	}
	return Status{Wired: wired, Wireless: wireless}, nil
}

func classify(devs []HIDDevice) (wired, dongles []Device) {
	type key struct {
		vid, pid uint16
		serial   string
	}
	seen := map[key]bool{}
	for _, d := range devs {
		k, ok := lookup(d)
		if !ok {
			continue
		}
		if d.SerialNumber != "" {
			id := key{d.VendorID, d.ProductID, d.SerialNumber}
			if seen[id] {
				continue
			}
			seen[id] = true
		}
		if k.Dongle {
			dongles = append(dongles, Device{HIDDevice: d, Kind: KindWireless, Name: k.Name})
		} else {
			wired = append(wired, Device{HIDDevice: d, Kind: KindWired, Name: k.Name})
		}
	}
	return wired, dongles
}

func lookup(d HIDDevice) (KnownDevice, bool) {
	for _, k := range KnownDevices {
		if d.VendorID == k.VendorID && d.ProductID == k.ProductID {
			return k, true
		}
	}
	return KnownDevice{}, false
}

var probeInputReport = readInputReport

func probeDongles(ctx context.Context, dongles []Device) ([]Device, error) {
	if len(dongles) == 0 {
		return nil, nil
	}
	active := make([]bool, len(dongles))
	errs := make([]error, len(dongles))
	var wg sync.WaitGroup
	for i, d := range dongles {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			active[i], errs[i] = probeInputReport(ctx, path)
		}(i, d.Path)
	}
	wg.Wait()

	var out []Device
	for i, d := range dongles {
		if errs[i] != nil {
			if errors.Is(errs[i], ErrUnsupported) {
				return nil, errs[i]
			}
			continue
		}
		if active[i] {
			out = append(out, d)
		}
	}
	return out, nil
}

var vidPidRe = regexp.MustCompile(`(?i)vid_([0-9a-f]{4})&pid_([0-9a-f]{4})`)

func parseVIDPID(path string) (vid, pid uint16, ok bool) {
	m := vidPidRe.FindStringSubmatch(path)
	if m == nil {
		return 0, 0, false
	}
	v, err1 := strconv.ParseUint(m[1], 16, 16)
	p, err2 := strconv.ParseUint(m[2], 16, 16)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return uint16(v), uint16(p), true
}

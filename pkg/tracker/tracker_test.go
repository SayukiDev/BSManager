package tracker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParseVIDPID(t *testing.T) {
	tests := []struct {
		path     string
		vid, pid uint16
		ok       bool
	}{
		{`\?\hid#vid_28de&pid_2101&mi_00#7&1a2b3c4d&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}`, 0x28DE, 0x2101, true},
		{`\?\HID#VID_0BB4&PID_0306#6&abc#{...}`, 0x0BB4, 0x0306, true},
		{`\?\hid#converteddevice&col01#5&...`, 0, 0, false},
		{``, 0, 0, false},
	}
	for _, tc := range tests {
		vid, pid, ok := parseVIDPID(tc.path)
		if vid != tc.vid || pid != tc.pid || ok != tc.ok {
			t.Errorf("parseVIDPID(%q) = %04X/%04X/%v, want %04X/%04X/%v",
				tc.path, vid, pid, ok, tc.vid, tc.pid, tc.ok)
		}
	}
}

func TestClassify(t *testing.T) {
	devs := []HIDDevice{
		{Path: "a", VendorID: VendorValve, ProductID: 0x2300, SerialNumber: "LHR-AAA"},
		{Path: "b", VendorID: VendorValve, ProductID: 0x2300, SerialNumber: "LHR-AAA"}, // 同一デバイスの別インターフェース
		{Path: "c", VendorID: VendorValve, ProductID: 0x2300, SerialNumber: "LHR-BBB"},
		{Path: "d", VendorID: VendorValve, ProductID: 0x2102, SerialNumber: "D1"},
		{Path: "e", VendorID: VendorValve, ProductID: 0x2101}, // シリアルなしはまとめない
		{Path: "f", VendorID: VendorValve, ProductID: 0x2101},
		{Path: "g", VendorID: VendorValve, ProductID: 0x2012, Product: "Controller"}, // Vive コントローラーは対象外
		{Path: "h", VendorID: 0x046D, ProductID: 0xC52B, Product: "USB Receiver"},
	}
	wired, dongles := classify(devs)
	if len(wired) != 2 {
		t.Errorf("wired = %d, want 2", len(wired))
	}
	if len(dongles) != 3 {
		t.Errorf("dongles = %d, want 3", len(dongles))
	}
	for _, d := range wired {
		if d.Kind != KindWired {
			t.Errorf("wired device %s has kind %v", d.Path, d.Kind)
		}
	}
	for _, d := range dongles {
		if d.Kind != KindWireless {
			t.Errorf("dongle %s has kind %v", d.Path, d.Kind)
		}
	}
}

func TestProbeDongles(t *testing.T) {
	orig := probeInputReport
	defer func() { probeInputReport = orig }()

	probeInputReport = func(ctx context.Context, path string) (bool, error) {
		switch path {
		case "active":
			return true, nil
		case "broken":
			return false, errors.New("access denied")
		default:
			<-ctx.Done()
			return false, nil
		}
	}
	dongles := []Device{
		{HIDDevice: HIDDevice{Path: "idle"}, Kind: KindWireless},
		{HIDDevice: HIDDevice{Path: "active"}, Kind: KindWireless},
		{HIDDevice: HIDDevice{Path: "broken"}, Kind: KindWireless},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	got, err := probeDongles(ctx, dongles)
	if err != nil {
		t.Fatalf("probeDongles: %v", err)
	}
	if len(got) != 1 || got[0].Path != "active" {
		t.Errorf("probeDongles = %+v, want only 'active'", got)
	}

	probeInputReport = func(ctx context.Context, path string) (bool, error) { return false, ErrUnsupported }
	if _, err := probeDongles(context.Background(), dongles); !errors.Is(err, ErrUnsupported) {
		t.Errorf("expected ErrUnsupported, got %v", err)
	}

	if got, err := probeDongles(context.Background(), nil); err != nil || got != nil {
		t.Errorf("empty input: got %v, %v", got, err)
	}
}

func TestStatus(t *testing.T) {
	if (Status{}).Any() || (Status{}).Count() != 0 {
		t.Error("empty status should be empty")
	}
	st := Status{Wired: []Device{{}}, Wireless: []Device{{}, {}}}
	if !st.Any() || st.Count() != 3 {
		t.Errorf("Count = %d, want 3", st.Count())
	}
}

// TestDetectReal は実機で列挙と判定がエラーなく完走することを確認し、結果をログに出す。
func TestDetectReal(t *testing.T) {
	devs, err := EnumerateHID()
	if errors.Is(err, ErrUnsupported) {
		t.Skip("HID not supported on this platform")
	}
	if err != nil {
		t.Fatalf("EnumerateHID: %v", err)
	}
	for _, d := range devs {
		if d.VendorID == VendorValve || d.VendorID == VendorHTC {
			t.Logf("hid %04X:%04X %q %q %q", d.VendorID, d.ProductID, d.Manufacturer, d.Product, d.SerialNumber)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	st, err := Detect(ctx)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	t.Logf("Detect took %v", time.Since(start))
	for _, d := range st.Wired {
		t.Logf("wired:    %s serial=%s", d.Name, d.SerialNumber)
	}
	for _, d := range st.Wireless {
		t.Logf("wireless: via %s dongle serial=%s", d.Name, d.SerialNumber)
	}
	t.Logf("wired=%d wireless=%d any=%v", len(st.Wired), len(st.Wireless), st.Any())
}

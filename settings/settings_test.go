package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s := NewSettings()
	if err := s.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.FilePath() != path {
		t.Fatalf("FilePath = %q, want %q", s.FilePath(), path)
	}
	if s.CheckInterval != DefaultCheckInterval || s.ShutdownWaiting != DefaultShutdownWaiting {
		t.Fatalf("defaults not kept: %+v", s.Content)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.json")
	s := NewSettings()
	if err := s.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	s.SetContent(Content{
		Setuped:  true,
		LogLevel: "debug",
		Devices: []BaseStation{
			{Enable: true, Name: " lhb-1 ", Addr: "aa:bb:cc:dd:ee:ff"},
			{Enable: false, Name: "dup", Addr: " AA:BB:CC:DD:EE:FF "},
			{Enable: false, Name: "lhb-2", Addr: "11:22:33:44:55:66"},
			{Enable: true, Addr: "   "},
		},
		CheckInterval: 0,
		FollowSteamVR: false,
	})
	if err := s.Save(""); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("tmp file should be removed")
	}

	s2 := NewSettings()
	if err := s2.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	c := s2.GetContent()
	if !c.Setuped || c.LogLevel != "debug" || c.FollowSteamVR {
		t.Fatalf("unexpected content: %+v", c)
	}
	if c.CheckInterval != DefaultCheckInterval {
		t.Fatalf("CheckInterval should be normalized, got %d", c.CheckInterval)
	}
	want := []BaseStation{
		{Enable: true, Name: "lhb-1", Addr: "AA:BB:CC:DD:EE:FF"},
		{Enable: false, Name: "lhb-2", Addr: "11:22:33:44:55:66"},
	}
	if len(c.Devices) != len(want) {
		t.Fatalf("Devices = %v, want %v", c.Devices, want)
	}
	for i := range want {
		if c.Devices[i] != want[i] {
			t.Fatalf("Devices = %v, want %v", c.Devices, want)
		}
	}
	enabled := c.EnabledAddrs()
	if len(enabled) != 1 || enabled[0] != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("EnabledAddrs = %v", enabled)
	}
}

func TestGetContentIsCopy(t *testing.T) {
	s := NewSettings()
	s.SetContent(Content{Devices: []BaseStation{{Enable: true, Addr: "AA:BB:CC:DD:EE:FF"}}})
	c := s.GetContent()
	c.Devices[0].Addr = "changed"
	if s.GetContent().Devices[0].Addr != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("GetContent must return a copy")
	}
}

func TestSaveWithoutPath(t *testing.T) {
	s := NewSettings()
	if err := s.Save(""); err == nil {
		t.Fatalf("Save without path should fail")
	}
}

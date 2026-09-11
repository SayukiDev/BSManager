package manager

import (
	"BSManager/settings"
	"runtime"
	"testing"
	"time"
)

func TestCloseBeforeStartIsNoop(t *testing.T) {
	m := NewManager(settings.NewSettings())
	m.Close()
	m.Close()
	if m.IsRunning() {
		t.Fatalf("should not be running")
	}
}

func TestStartRejectsInvalidInterval(t *testing.T) {
	s := settings.NewSettings()
	s.CheckInterval = 0
	m := NewManager(s)
	if err := m.Start(); err != ErrInvalidInterval {
		t.Fatalf("Start = %v, want ErrInvalidInterval", err)
	}
}

func TestRestartDoesNotLeakGoroutines(t *testing.T) {
	s := settings.NewSettings()
	s.CheckInterval = 1
	s.FollowSteamVR = false
	m := NewManager(s)

	before := runtime.NumGoroutine()
	for range 5 {
		if err := m.Start(); err != nil {
			t.Fatalf("Start: %v", err)
		}
		if !m.IsRunning() {
			t.Fatalf("should be running")
		}
		if err := m.Start(); err != nil {
			t.Fatalf("Start twice: %v", err)
		}
		m.Close()
		if m.IsRunning() {
			t.Fatalf("should be stopped")
		}
	}
	if err := m.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	m.Close()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > before {
		t.Fatalf("goroutine leak: before=%d after=%d", before, after)
	}
}

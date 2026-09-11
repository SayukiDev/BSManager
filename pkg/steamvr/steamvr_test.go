package steamvr

import "testing"

func TestMatchAny(t *testing.T) {
	tests := []struct {
		name    string
		running []string
		want    bool
	}{
		{"empty", nil, false},
		{"unrelated", []string{"explorer.exe", "steam.exe"}, false},
		{"vrserver", []string{"explorer.exe", "vrserver.exe"}, true},
		{"vrmonitor case-insensitive", []string{"VRMonitor.EXE"}, true},
		{"partial name does not match", []string{"vrserver.exe.bak", "myvrserver.exe"}, false},
	}
	for _, tc := range tests {
		if got := matchAny(tc.running); got != tc.want {
			t.Errorf("%s: matchAny = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestIsRunningReal は実機で判定が完走することを確認し、結果をログに出す。
func TestIsRunningReal(t *testing.T) {
	names, err := listProcessNames()
	if err != nil {
		t.Skipf("listProcessNames: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("no processes listed")
	}
	t.Logf("processes=%d SteamVR running=%v", len(names), IsRunning())
}

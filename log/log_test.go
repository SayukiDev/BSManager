package log

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestSetLogFileAndReadTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "test.log")
	t.Cleanup(Close)
	if err := SetLogFile(path); err != nil {
		t.Fatalf("SetLogFile: %v", err)
	}
	if LogFilePath() != path {
		t.Fatalf("LogFilePath = %q, want %q", LogFilePath(), path)
	}

	SetLogLevel("debug")
	Debug("debug line")
	for i := range 10 {
		Info("line", zap.Int("i", i))
	}
	_ = logger.Sync()

	lines, err := ReadTail(3)
	if err != nil {
		t.Fatalf("ReadTail: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3", len(lines))
	}
	var last struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		I     int    `json:"i"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &last); err != nil {
		t.Fatalf("line is not JSON: %q: %v", lines[2], err)
	}
	if last.Msg != "line" || last.I != 9 {
		t.Fatalf("unexpected last line: %+v", last)
	}

	all, err := ReadTail(0)
	if err != nil {
		t.Fatalf("ReadTail(0): %v", err)
	}
	if len(all) != 11 {
		t.Fatalf("len(all) = %d, want 11", len(all))
	}
	if !strings.Contains(all[0], "debug line") {
		t.Fatalf("debug level should be written after SetLogLevel(debug): %q", all[0])
	}

	SetLogLevel("error")
	Info("should not appear")
	_ = logger.Sync()
	after, _ := ReadTail(0)
	if len(after) != 11 {
		t.Fatalf("info should be filtered at error level, got %d lines", len(after))
	}
}

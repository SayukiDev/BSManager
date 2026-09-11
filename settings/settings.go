package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

const (
	AppDirName = "BSManager"
	FileName   = "settings.json"

	DefaultCheckInterval   = 10
	DefaultShutdownWaiting = 60
	DefaultMinTrackerCount = 2
	DefaultLogLevel        = "info"
)

var validLogLevels = []string{"debug", "info", "warn", "error", "fatal"}

type Content struct {
	Setuped               bool          `json:"setuped"`
	LogLevel              string        `json:"logLevel"`
	Devices               []BaseStation `json:"devices"`
	CheckInterval         int           `json:"checkInterval"`
	FollowSteamVR         bool          `json:"followSteamVR"`
	CheckTrackerConnected bool          `json:"checkTrackerConnected"`
	MinTrackerCount       int           `json:"minTrackerCount"`
	ShutdownWaiting       int           `json:"shutdownWaiting"`
}

func (c *Content) Normalize() {
	if c.CheckInterval < 1 {
		c.CheckInterval = DefaultCheckInterval
	}
	if c.ShutdownWaiting < 0 {
		c.ShutdownWaiting = 0
	}
	if c.MinTrackerCount < 0 {
		c.MinTrackerCount = 0
	}
	if !slices.Contains(validLogLevels, c.LogLevel) {
		c.LogLevel = DefaultLogLevel
	}
	c.Devices = NormalizeDevices(c.Devices)
}

type BaseStation struct {
	Enable bool   `json:"enable"`
	Name   string `json:"name"`
	Addr   string `json:"addr"`
}

func (c Content) Clone() Content {
	c.Devices = slices.Clone(c.Devices)
	if c.Devices == nil {
		c.Devices = []BaseStation{}
	}
	return c
}

type Settings struct {
	lock     sync.RWMutex
	filepath string
	Content
}

func NewSettings() *Settings {
	return &Settings{
		Content: Content{
			Setuped:               false,
			LogLevel:              DefaultLogLevel,
			Devices:               []BaseStation{},
			CheckInterval:         DefaultCheckInterval,
			FollowSteamVR:         true,
			CheckTrackerConnected: true,
			MinTrackerCount:       DefaultMinTrackerCount,
			ShutdownWaiting:       DefaultShutdownWaiting,
		},
	}
}

func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppDirName), nil
}

func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, FileName), nil
}

func (s *Settings) FilePath() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.filepath
}

func (s *Settings) Load(path string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.filepath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	c := s.Content.Clone()
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	c.Normalize()
	s.Content = c
	return nil
}

func (s *Settings) Save(path string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if path == "" {
		path = s.filepath
	}
	if path == "" {
		return errors.New("settings: save path is not set")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Content, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	s.filepath = path
	return nil
}

func (s *Settings) GetContent() Content {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Content.Clone()
}

func (s *Settings) SetContent(c Content) {
	c = c.Clone()
	c.Normalize()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Content = c
}

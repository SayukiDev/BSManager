package manager

import (
	"BSManager/settings"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrInvalidInterval = errors.New("manager: check interval must be >= 1")

type Manager struct {
	mu               sync.Mutex
	wg               sync.WaitGroup
	closeC           chan struct{}
	sets             *settings.Settings
	trackerCount     int
	steamVRActive    bool
	powerOn          atomic.Bool
	inited           atomic.Bool
	onStatusChanged  func(bool)
	onStatusChanging func(bool)
}

func NewManager(sets *settings.Settings) *Manager {
	return &Manager{
		sets:             sets,
		onStatusChanged:  func(bool) {},
		onStatusChanging: func(bool) {},
	}
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closeC != nil {
		return nil
	}
	if m.sets.CheckInterval < 1 {
		return ErrInvalidInterval
	}
	if m.sets.FollowSteamVR && !m.inited.Load() {
		m.inited.Store(true)
		s, err := m.checkStatus()
		if err != nil {
			return err
		}
		err = m.SetStatus(s)
		if err != nil {
			return err
		}
	}
	closeC := make(chan struct{})
	m.closeC = closeC
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.startHandle(closeC)
	}()
	return nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	closeC := m.closeC
	m.closeC = nil
	m.mu.Unlock()
	if closeC == nil {
		return
	}
	close(closeC)
	m.wg.Wait()
}

func (m *Manager) Restart() error {
	m.Close()
	return m.Start()
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeC != nil
}

func (m *Manager) IsPowerOn() bool {
	return m.powerOn.Load()
}

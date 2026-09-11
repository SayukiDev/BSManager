package manager

import (
	"BSManager/log"
	"BSManager/pkg/steamvr"
	"BSManager/pkg/tracker"
	"context"
	"time"

	"go.uber.org/zap"
)

func (m *Manager) SetOnStatusChanged(f func(bool)) {
	m.onStatusChanged = f
}

func (m *Manager) SetOnStatusChanging(f func(bool)) {
	m.onStatusChanging = f
}

func (m *Manager) checkStatus() (bool, error) {
	if !m.sets.FollowSteamVR {
		return false, nil
	}
	if !steamvr.IsRunning() {
		return false, nil
	}
	m.steamVRActive = true
	if m.sets.CheckTrackerConnected {
		ts, err := tracker.Detect(context.Background())
		if err != nil {
			return false, err
		}
		m.trackerCount = ts.Count()
		if ts.Count() < m.sets.MinTrackerCount {
			return false, nil
		}
	}
	return true, nil
}

func (m *Manager) startHandle(closeC <-chan struct{}) {
	shutdownTime := time.Time{}
	ticker := time.NewTicker(time.Second * time.Duration(m.sets.CheckInterval))
	defer ticker.Stop()
	for {
		select {
		case <-closeC:
			return
		case <-ticker.C:
		}
		isActive, err := m.checkStatus()
		if err != nil {
			log.Error("check status error", zap.Error(err))
			continue
		}
		if isActive {
			if !shutdownTime.IsZero() {
				shutdownTime = time.Time{}
			}
			if m.powerOn.Load() {
				continue
			}
			err = m.SetStatus(true)
			if err != nil {
				log.Error("sync power error", zap.Error(err))
			}
			continue
		}
		if isActive == m.powerOn.Load() {
			continue
		}
		if m.sets.ShutdownWaiting > 0 {
			if shutdownTime.IsZero() {
				shutdownTime = time.Now()
				continue
			}
			if shutdownTime.
				Add(time.Second * time.Duration(m.sets.ShutdownWaiting)).
				After(time.Now()) {
				continue
			}
			shutdownTime = time.Time{}
		}
		err = m.SetStatus(false)
		if err != nil {
			log.Error("sync power error", zap.Error(err))
		}
	}
}

package manager

import bs "github.com/SayukiDev/BasestationApiGo"

func (m *Manager) syncBSPower() error {
	var err error
	if m.powerOn.Load() {
		err = bs.SetPower(bs.PwrBooting)
		if err != nil {
			return err
		}
	} else {
		err = bs.SetPower(bs.PwrSleep)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetTrackerCount() int {
	return m.trackerCount
}

func (m *Manager) SetStatus(isActive bool) error {
	m.powerOn.Store(isActive)
	m.onStatusChanging(isActive)
	err := m.syncBSPower()
	if err != nil {
		return err
	}
	m.onStatusChanged(isActive)
	return nil
}

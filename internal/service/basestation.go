package service

import (
	"BSManager/log"
	"errors"
	"strings"

	bs "github.com/SayukiDev/BasestationApiGo"
	"go.uber.org/zap"
)

var ErrScanInProgress = errors.New("scan is already in progress")

type BaseStation struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	RSSI int    `json:"rssi"`
}

func (s *Service) SetOnScanUpdate(f func()) {
	if f == nil {
		f = func() {}
	}
	s.onScanUpdate = f
}

func (s *Service) SetOnScanStopped(f func(error)) {
	if f == nil {
		f = func(error) {}
	}
	s.onScanStopped = f
}

func (s *Service) StartScan() error {
	if !s.scanning.CompareAndSwap(false, true) {
		return ErrScanInProgress
	}
	s.scanning.Store(true)
	log.Info("realtime scan start")
	bs.SetOnScanDeviceUpdate(func(_ bs.DeviceInfo) {
		s.onScanUpdate()
	})
	bs.SetOnScanDeviceFailed(func(err error) {
		if !s.scanning.CompareAndSwap(true, false) {
			return
		}
		bs.SetOnScanDeviceUpdate(func(bs.DeviceInfo) {})
		log.Error("realtime scan failed", zap.Error(err))
		s.onScanStopped(err)
	})
	err := bs.Scanning()
	if err != nil {
		log.Error("realtime scan error", zap.Error(err))
	} else {
		log.Info("realtime scan finished")
	}
	if err != nil {
		s.onScanStopped(err)
		s.scanning.Store(false)
		return err
	}
	return nil
}

func (s *Service) StopScan() error {
	if !s.scanning.CompareAndSwap(true, false) {
		return nil
	}
	err := bs.StopScanning()
	bs.SetOnScanDeviceUpdate(func(bs.DeviceInfo) {})
	s.onScanStopped(err)
	return err
}

func (s *Service) GetScanedBaseStations() []BaseStation {
	temp := bs.GetBaseStation()
	bss := make([]BaseStation, 0, len(temp))
	for _, b := range temp {
		bss = append(bss, BaseStation{
			Name: b.Name,
			Addr: b.Addr,
			RSSI: int(b.RSSi),
		})
	}
	return bss
}

func (s *Service) IsScanning() bool {
	return s.scanning.Load()
}

func (s *Service) GetPowerState() (map[string]string, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	states, err := bs.GetPowerState()
	if err != nil {
		if len(states) == 0 {
			return nil, err
		}
		log.Warn("get power state partial error", zap.Error(err))
	}
	return states, nil
}

func (s *Service) SetPower(on bool) error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	if err := s.M.SetStatus(on); err != nil {
		log.Error("set power error", zap.Bool("on", on), zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) SetPowerSome(on bool, addr ...string) error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	if on {
		if err := bs.SetPowerSome(bs.PwrBooting, addr...); err != nil {
			log.Error("set power error", zap.Bool("on", on), zap.Error(err))
			return err
		}
		return nil
	}
	if err := bs.SetPowerSome(bs.PwrSleep, addr...); err != nil {
		log.Error("set power error", zap.Bool("on", on), zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) Identify(addr string) error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	return bs.Identify(strings.ToUpper(strings.TrimSpace(addr)))
}

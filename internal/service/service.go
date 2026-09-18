package service

import (
	"BSManager/internal/manager"
	"BSManager/log"
	"BSManager/settings"
	"sync"
	"sync/atomic"
	"time"

	bs "github.com/SayukiDev/BasestationApiGo"
	"go.uber.org/zap"
)

type Service struct {
	M    *manager.Manager
	Sets *settings.Settings

	opMu sync.Mutex

	scanning      atomic.Bool
	onScanUpdate  func()
	onScanStopped func(error)
}

func NewService(sets *settings.Settings) *Service {
	bs.SetDeviceControl(sets.GetContent().EnabledAddrs())
	return &Service{
		M:             manager.NewManager(sets),
		Sets:          sets,
		onScanUpdate:  func() {},
		onScanStopped: func(error) {},
	}
}

func (s *Service) SetOnStatusChanged(f func(bool)) {
	s.M.SetOnStatusChanged(f)
}

func (s *Service) SetOnStatusChanging(f func(bool)) {
	s.M.SetOnStatusChanging(f)
}

func (s *Service) Start() error {
	c := s.Sets.GetContent()
	if !c.Setuped || len(c.EnabledAddrs()) == 0 {
		log.Info("setup is not completed, manager is not started")
		return nil
	}
	// Scan 10 sec to avoid trouble
	go func() {
		err := bs.ScanningWithTimeout(10 * time.Second)
		if err != nil {
			log.Error("scan failed", zap.Error(err))
		}
	}()
	return s.M.Start()
}

func (s *Service) Stop() {
	s.M.Close()
}

func (s *Service) Restart() error {
	s.Stop()
	return s.Start()
}

func (s *Service) Shutdown(timeout time.Duration) {
	s.Stop()
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := bs.StopScanning(); err != nil {
			log.Warn("stop scanning error", zap.Error(err))
		}
		if err := bs.Disconnect(); err != nil {
			log.Warn("disconnect error", zap.Error(err))
		}
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		log.Warn("ble cleanup timed out")
	}
}

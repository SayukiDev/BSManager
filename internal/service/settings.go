package service

import (
	"BSManager/log"
	"BSManager/settings"
	"errors"
	"fmt"
	"slices"

	bs "github.com/SayukiDev/BasestationApiGo"
	"go.uber.org/zap"
	"tinygo.org/x/bluetooth"
)

var ErrNoDevice = errors.New("no device specified")

func (s *Service) GetSettings() settings.Content {
	return s.Sets.GetContent()
}

func (s *Service) SaveSettings(c settings.Content) error {
	c = c.Clone()
	c.Normalize()
	if err := validateDevices(c.Devices); err != nil {
		return err
	}

	s.opMu.Lock()
	defer s.opMu.Unlock()

	prev := s.Sets.GetContent()
	s.Stop()

	if !sameAddrs(prev.EnabledAddrs(), c.EnabledAddrs()) {
		if err := bs.Disconnect(); err != nil {
			log.Warn("disconnect before device change error", zap.Error(err))
		}
		bs.SetDeviceControl(c.EnabledAddrs())
	}

	s.Sets.SetContent(c)
	if err := s.Sets.Save(""); err != nil {
		log.Error("save settings error", zap.Error(err))
		return err
	}
	log.SetLogLevel(c.LogLevel)
	log.Info("settings saved", zap.Any("settings", c))

	if err := s.Start(); err != nil {
		log.Error("restart manager error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) SetDevices(devices []settings.BaseStation) error {
	devices = settings.NormalizeDevices(devices)
	if len(devices) == 0 {
		return ErrNoDevice
	}
	c := s.Sets.GetContent()
	c.Devices = devices
	c.Setuped = true
	return s.SaveSettings(c)
}

func validateDevices(devices []settings.BaseStation) error {
	for _, d := range devices {
		if _, err := bluetooth.ParseMAC(d.Addr); err != nil {
			return fmt.Errorf("invalid address %q: %w", d.Addr, err)
		}
	}
	return nil
}

func sameAddrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := slices.Clone(a)
	y := slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

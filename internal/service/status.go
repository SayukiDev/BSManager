package service

import (
	"BSManager/log"
	"BSManager/pkg/steamvr"
	"BSManager/settings"
	"time"
)

const trackerDetectTimeout = time.Second

type Status struct {
	Setuped        bool     `json:"setuped"`
	PowerOn        bool     `json:"powerOn"`
	FollowSteamVR  bool     `json:"followSteamVR"`
	SteamVRRunning bool     `json:"steamVRRunning"`
	ManagerRunning bool     `json:"managerRunning"`
	TrackerCount   int      `json:"trackerCount"`
	Devices        []settings.BaseStation `json:"devices"`
}

func (s *Service) GetStatus() Status {
	c := s.Sets.GetContent()
	st := Status{
		Setuped:        c.Setuped,
		PowerOn:        s.M.IsPowerOn(),
		FollowSteamVR:  c.FollowSteamVR,
		SteamVRRunning: steamvr.IsRunning(),
		ManagerRunning: s.M.IsRunning(),
		Devices:        c.Devices,
	}
	if c.CheckTrackerConnected {
		st.TrackerCount = s.M.GetTrackerCount()
	}
	return st
}

func (s *Service) GetLogPath() string {
	return log.LogFilePath()
}

func (s *Service) ReadLogs(lines int) ([]string, error) {
	return log.ReadTail(lines)
}

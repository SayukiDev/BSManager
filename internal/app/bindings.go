package app

import (
	"BSManager/internal/service"
	"BSManager/log"
	"BSManager/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetSettings() settings.Content {
	return a.svc.GetSettings()
}

func (a *App) SaveSettings(c settings.Content) error {
	if err := a.svc.SaveSettings(c); err != nil {
		return err
	}
	a.emit(EventSettingsChanged, a.svc.GetSettings())
	return nil
}

func (a *App) SetDevices(devices []settings.BaseStation) error {
	if err := a.svc.SetDevices(devices); err != nil {
		return err
	}
	a.emit(EventSettingsChanged, a.svc.GetSettings())
	return nil
}

func (a *App) SetPowerSome(on bool, addrs []string) error {
	return a.svc.SetPowerSome(on, addrs...)
}

func (a *App) StopScan() error {
	return a.svc.StopScan()
}

func (a *App) GetPowerState() (map[string]string, error) {
	return a.svc.GetPowerState()
}

func (a *App) SetPower(on bool) error {
	return a.svc.SetPower(on)
}

func (a *App) Identify(addr string) error {
	return a.svc.Identify(addr)
}

func (a *App) GetStatus() service.Status {
	return a.svc.GetStatus()
}

func (a *App) GetLogPath() string {
	return a.svc.GetLogPath()
}

func (a *App) ReadLogs(lines int) ([]string, error) {
	return a.svc.ReadLogs(lines)
}

func (a *App) StartScan() error {
	return a.svc.StartScan()
}

func (a *App) IsScanning() bool {
	return a.svc.IsScanning()
}

func (a *App) GetScanedBaseStations() []service.BaseStation {
	return a.svc.GetScanedBaseStations()
}

func (a *App) HideWindow() {
	a.ctxMu.RLock()
	ctx := a.ctx
	a.ctxMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.WindowHide(ctx)
	log.Info("window hidden, running in background")
}

func (a *App) ShowWindow() {
	a.ctxMu.RLock()
	ctx := a.ctx
	a.ctxMu.RUnlock()
	if ctx == nil {
		return
	}
	if runtime.WindowIsMinimised(ctx) {
		runtime.WindowUnminimise(ctx)
	}
	runtime.WindowShow(ctx)
}

func (a *App) Quit() {
	a.ctxMu.RLock()
	ctx := a.ctx
	a.ctxMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.Quit(ctx)
}

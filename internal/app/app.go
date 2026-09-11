package app

import (
	"BSManager/internal/service"
	"BSManager/log"
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

const shutdownTimeout = 3 * time.Second

type App struct {
	ctx   context.Context
	ctxMu sync.RWMutex
	svc   *service.Service

	trayIcon []byte
	trayEnd  func()
}

func NewApp(svc *service.Service, trayIcon []byte) *App {
	return &App{
		svc:      svc,
		trayIcon: trayIcon,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctxMu.Lock()
	a.ctx = ctx
	a.ctxMu.Unlock()

	a.svc.SetOnStatusChanging(func(on bool) {
		a.emit(EventPowerChanging, on)
	})
	a.svc.SetOnStatusChanged(func(on bool) {
		a.emit(EventPowerChanged, on)
	})
	a.svc.SetOnScanUpdate(func() {
		a.emit(EventScanUpdate)
	})
	a.svc.SetOnScanStopped(func(err error) {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		a.emit(EventScanStopped, msg)
	})
	if a.svc.Sets.Setuped {
		err := a.StartSVC()
		if err != nil {
			log.Error("start svc failed", zap.Error(err))
			a.emit(EventError, "start: "+err.Error())
		}
	}
	a.emit(EventReady)
	a.startTray()
}

func (a *App) StartSVC() error {
	if err := a.svc.Start(); err != nil {
		log.Error("start error", zap.Error(err))
		a.emit(EventError, "start: "+err.Error())
	}
	return nil
}

func (a *App) Shutdown(_ context.Context) {
	log.Info("shutdown")
	a.stopTray()
	a.svc.Shutdown(shutdownTimeout)
	log.Close()
}

func (a *App) OnSecondInstanceLaunch(_ options.SecondInstanceData) {
	a.ShowWindow()
}

func (a *App) emit(name string, data ...any) {
	a.ctxMu.RLock()
	ctx := a.ctx
	a.ctxMu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, name, data...)
}

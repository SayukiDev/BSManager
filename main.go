package main

import (
	"BSManager/internal/app"
	"BSManager/internal/service"
	"BSManager/log"
	"BSManager/settings"
	"embed"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIcon []byte

func main() {
	dir, err := settings.DefaultDir()
	if err != nil {
		log.ErrorE("config dir unavailable", zap.Error(err))
	}
	if err := log.SetCrashLog(filepath.Join(dir, log.CrashLogFileName)); err != nil {
		log.Warn("set crash log failed", zap.Error(err))
	}

	if err := log.SetLogFile(filepath.Join(dir, log.LogFileName)); err != nil {
		log.Error("open log file failed", zap.Error(err))
	}
	sets := settings.NewSettings()
	settingsPath := filepath.Join(dir, settings.FileName)
	if err := sets.Load(settingsPath); err != nil {
		log.Error("load settings failed", zap.String("path", settingsPath), zap.Error(err))
	}
	c := sets.GetContent()
	log.SetLogLevel(c.LogLevel)
	svc := service.NewService(sets)
	a := app.NewApp(svc, trayIcon)
	err = wails.Run(&options.App{
		Title:     "BSManager",
		Width:     1200,
		Height:    860,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  a.Startup,
		OnShutdown: a.Shutdown,
		Frameless:  true,
		Bind: []any{
			a,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "b7c9e5d2-4f61-4a8e-9c3b-2d7f0e1a6b58",
			OnSecondInstanceLaunch: a.OnSecondInstanceLaunch,
		},
	})
	if err != nil {
		log.Error("wails run error", zap.Error(err))
	}
}

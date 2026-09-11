package app

import (
	"BSManager/log"

	"github.com/energye/systray"
)

const trayTooltip = "BSManager"

func (a *App) startTray() {
	if len(a.trayIcon) == 0 {
		log.Warn("tray icon is empty, tray disabled")
		return
	}
	start, end := systray.RunWithExternalLoop(a.onTrayReady, nil)
	a.trayEnd = end
	start()
}

func (a *App) stopTray() {
	if a.trayEnd == nil {
		return
	}
	a.trayEnd()
	a.trayEnd = nil
}

func (a *App) onTrayReady() {
	systray.SetIcon(a.trayIcon)
	systray.SetTooltip(trayTooltip)
	systray.SetOnDClick(func(systray.IMenu) { a.ShowWindow() })
	systray.SetOnRClick(func(menu systray.IMenu) {
		if err := menu.ShowMenu(); err != nil {
			log.Warn("show tray menu error: " + err.Error())
		}
	})
	open := systray.AddMenuItem("BSManager を開く", "ウィンドウを表示します")
	open.Click(a.ShowWindow)
	systray.AddSeparator()
	quit := systray.AddMenuItem("終了", "バックグラウンド動作を止めて終了します")
	quit.Click(a.Quit)
}

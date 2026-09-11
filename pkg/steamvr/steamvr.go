// Package steamvr は SteamVR が実行中かをプロセスの有無で判定する。
package steamvr

import "strings"

var ProcessNames = []string{
	"vrserver.exe",
	"vrmonitor.exe",
}

func IsRunning() bool {
	names, err := listProcessNames()
	if err != nil {
		return false
	}
	return matchAny(names)
}

func matchAny(running []string) bool {
	for _, r := range running {
		for _, want := range ProcessNames {
			if strings.EqualFold(r, want) {
				return true
			}
		}
	}
	return false
}

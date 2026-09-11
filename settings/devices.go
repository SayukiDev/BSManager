package settings

import "strings"

func NormalizeDevices(devices []BaseStation) []BaseStation {
	out := make([]BaseStation, 0, len(devices))
	seen := make(map[string]struct{}, len(devices))
	for _, d := range devices {
		d.Addr = NormalizeAddr(d.Addr)
		if d.Addr == "" {
			continue
		}
		if _, ok := seen[d.Addr]; ok {
			continue
		}
		seen[d.Addr] = struct{}{}
		d.Name = strings.TrimSpace(d.Name)
		out = append(out, d)
	}
	return out
}

func NormalizeAddr(addr string) string {
	return strings.ToUpper(strings.TrimSpace(addr))
}

func (c Content) EnabledAddrs() []string {
	addrs := make([]string, 0, len(c.Devices))
	for _, d := range c.Devices {
		if d.Enable {
			addrs = append(addrs, d.Addr)
		}
	}
	return addrs
}

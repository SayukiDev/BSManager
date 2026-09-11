//go:build !windows

package steamvr

import "errors"

var errUnsupported = errors.New("steamvr: process enumeration is not supported on this platform")

func listProcessNames() ([]string, error) {
	return nil, errUnsupported
}

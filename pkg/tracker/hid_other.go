//go:build !windows

package tracker

import "context"

func EnumerateHID() ([]HIDDevice, error) {
	return nil, ErrUnsupported
}

func readInputReport(ctx context.Context, path string) (bool, error) {
	return false, ErrUnsupported
}

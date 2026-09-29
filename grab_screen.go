//go:build windows || linux

package main

import (
	"errors"
	"image"

	"github.com/kbinani/screenshot"
)

// grabScreen takes every screen at once, as the Print Screen key does.
func grabScreen() (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, errors.New("Aucun écran à capturer.")
	}
	all := screenshot.GetDisplayBounds(0)
	for i := 1; i < n; i++ {
		all = all.Union(screenshot.GetDisplayBounds(i))
	}
	return screenshot.CaptureRect(all)
}

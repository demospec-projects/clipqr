//go:build !linux

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// workArea: Windows and macOS already leave out the taskbar, the Dock and the
// menu bar, in the same units as SetPosition.
func workArea(screen *application.Screen) (work, whole application.Rect, scale float64) {
	return screen.WorkArea, screen.Bounds, 1
}

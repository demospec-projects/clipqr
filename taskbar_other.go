//go:build !linux || gtk3

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// hideFromTaskbar: Windows keeps the panel out of the taskbar with
// HiddenOnTaskbar, and macOS shows ClipQR in the menu bar only.
func hideFromTaskbar(*application.WebviewWindow) {}

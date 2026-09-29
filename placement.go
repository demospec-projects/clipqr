package main

import (
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	panelWidth  = 400
	panelHeight = 640
	panelMargin = 8
)

// placeNearClock puts the panel in the corner where the clock sits: bottom
// right just above the taskbar, or top right under a top bar (always on macOS,
// whose clock is in the menu bar). The panel can still be dragged afterwards.
func placeNearClock(app *application.App, panel *application.WebviewWindow) {
	screen := app.Screen.GetPrimary()
	if screen == nil {
		return
	}
	work, whole, scale := workArea(screen)
	width := int(float64(panelWidth) * scale)
	height := int(float64(panelHeight) * scale)
	margin := int(float64(panelMargin) * scale)

	x := work.X + work.Width - width - margin
	y := work.Y + work.Height - height - margin
	barOnTop := work.Y > whole.Y && work.Y+work.Height == whole.Y+whole.Height
	if runtime.GOOS == "darwin" || barOnTop {
		y = work.Y + margin
	}
	panel.SetPosition(x, y)
}

// placeAgainAfterWindowManager: an X11 window manager places a window itself
// when it appears (Cinnamon centers it), after the move asked before showing.
// The panel is put back above the clock once the window manager is done.
func placeAgainAfterWindowManager(app *application.App, panel *application.WebviewWindow) {
	for _, wait := range []time.Duration{50 * time.Millisecond, 150 * time.Millisecond, 300 * time.Millisecond} {
		time.Sleep(wait)
		placeNearClock(app, panel)
	}
}

func intersect(a, b application.Rect) application.Rect {
	x1, y1 := max(a.X, b.X), max(a.Y, b.Y)
	x2, y2 := min(a.X+a.Width, b.X+b.Width), min(a.Y+a.Height, b.Y+b.Height)
	if x2 <= x1 || y2 <= y1 {
		return a
	}
	return application.Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

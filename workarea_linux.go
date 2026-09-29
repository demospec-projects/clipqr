package main

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// workArea returns, in X11 pixels, the screen without its panels. Wails
// (GTK 4) reports the whole screen as work area, so the panels are read from
// what the window manager publishes (_NET_WORKAREA). Under Wayland there is
// nothing to read, and the compositor places the window itself anyway.
func workArea(screen *application.Screen) (work, whole application.Rect, scale float64) {
	whole = screen.PhysicalBounds
	work = whole
	scale = float64(screen.ScaleFactor)
	if scale <= 0 {
		scale = 1
	}
	conn, err := xgb.NewConn()
	if err != nil {
		return
	}
	defer conn.Close()
	name := "_NET_WORKAREA"
	atom, err := xproto.InternAtom(conn, true, uint16(len(name)), name).Reply()
	if err != nil || atom.Atom == xproto.AtomNone {
		return
	}
	root := xproto.Setup(conn).DefaultScreen(conn).Root
	prop, err := xproto.GetProperty(conn, false, root, atom.Atom, xproto.AtomCardinal, 0, 4).Reply()
	if err != nil || prop.Format != 32 || prop.ValueLen < 4 {
		return
	}
	value := func(i int) int { return int(xgb.Get32(prop.Value[i*4:])) }
	work = intersect(whole, application.Rect{X: value(0), Y: value(1), Width: value(2), Height: value(3)})
	return
}

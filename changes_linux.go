package main

import (
	"context"
	"os"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xfixes"
	"github.com/jezek/xgb/xproto"
)

// clipboardChanges signals each new copy. Under X11 the library would read
// the whole clipboard every second to notice a change, a screenshot included;
// XFixes instead tells when an application takes the clipboard, and only then
// is it read. Under Wayland the library is told by the compositor itself.
func clipboardChanges(ctx context.Context) <-chan struct{} {
	if os.Getenv("XDG_SESSION_TYPE") != "wayland" {
		if ch, err := xfixesChanges(ctx); err == nil {
			return ch
		}
	}
	return watchChanges(ctx)
}

func xfixesChanges(ctx context.Context) (<-chan struct{}, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}
	if err = xfixes.Init(conn); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err = xfixes.QueryVersion(conn, 5, 0).Reply(); err != nil {
		conn.Close()
		return nil, err
	}
	name := "CLIPBOARD"
	atom, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		conn.Close()
		return nil, err
	}
	root := xproto.Setup(conn).DefaultScreen(conn).Root
	if err = xfixes.SelectSelectionInputChecked(conn, root, atom.Atom, xfixes.SelectionEventMaskSetSelectionOwner).Check(); err != nil {
		conn.Close()
		return nil, err
	}
	ch := make(chan struct{}, 1)
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go func() {
		defer close(ch)
		for {
			ev, err := conn.WaitForEvent()
			if ev == nil && err == nil {
				return // connection closed
			}
			if _, ok := ev.(xfixes.SelectionNotifyEvent); ok {
				select {
				case ch <- struct{}{}:
				default: // a change is already waiting to be read
				}
			}
		}
	}()
	return ch, nil
}

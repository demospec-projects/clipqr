package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const keysymPrint = 0xff61

// Settings in which desktops reserve Print for their own screenshot tool.
var desktopPrintBindings = [][2]string{
	{"org.cinnamon.desktop.keybindings.media-keys", "screenshot"},
	{"org.gnome.shell.keybindings", "show-screenshot-ui"},
	{"org.gnome.settings-daemon.plugins.media-keys", "screenshot"},
}

type desktopBinding struct {
	Schema, Key, Value string
}

// printKey takes Print Screen away from the desktop while capture mode is on,
// and gives it back when it is turned off or ClipQR quits.
type printKey struct {
	mu        sync.Mutex
	conn      *xgb.Conn
	stateFile string
}

func newPrintKey(dataDir string) *printKey {
	return &printKey{stateFile: filepath.Join(dataDir, "impr-ecran-bureau.json")}
}

func printKeyAvailable() bool {
	return os.Getenv("XDG_SESSION_TYPE") != "wayland"
}

func (p *printKey) start(onPress func()) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		return nil
	}
	if !printKeyAvailable() {
		return errors.New("Sous Wayland, une application ne peut pas prendre la touche Impr. écran.")
	}
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	root := xproto.Setup(conn).DefaultScreen(conn).Root
	keycode, ok := keycodeOf(conn, keysymPrint)
	if !ok {
		conn.Close()
		return errors.New("Ce clavier n’a pas de touche Impr. écran.")
	}
	p.releaseDesktopBindings()
	// The desktop lets go of the key a moment after its setting changes.
	for try := 0; ; try++ {
		if err = grab(conn, root, keycode); err == nil {
			break
		}
		if try == 20 {
			conn.Close()
			p.restoreDesktopBindings()
			return errors.New("La touche Impr. écran est déjà prise par une autre application.")
		}
		time.Sleep(100 * time.Millisecond)
	}
	p.conn = conn
	go func() {
		for {
			ev, err := conn.WaitForEvent()
			if ev == nil && err == nil {
				return // connection closed by stop
			}
			if key, ok := ev.(xproto.KeyPressEvent); ok && key.Detail == keycode {
				onPress()
			}
		}
	}()
	return nil
}

func (p *printKey) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.restoreDesktopBindings()
}

// grab takes the key alone, whatever the state of Caps Lock and Num Lock.
func grab(conn *xgb.Conn, root xproto.Window, keycode xproto.Keycode) error {
	for _, mods := range []uint16{0, xproto.ModMaskLock, xproto.ModMask2, xproto.ModMaskLock | xproto.ModMask2} {
		if err := xproto.GrabKeyChecked(conn, true, root, mods, keycode, xproto.GrabModeAsync, xproto.GrabModeAsync).Check(); err != nil {
			xproto.UngrabKey(conn, keycode, root, xproto.ModMaskAny)
			return err
		}
	}
	return nil
}

func keycodeOf(conn *xgb.Conn, keysym xproto.Keysym) (xproto.Keycode, bool) {
	setup := xproto.Setup(conn)
	count := byte(setup.MaxKeycode - setup.MinKeycode + 1)
	mapping, err := xproto.GetKeyboardMapping(conn, setup.MinKeycode, count).Reply()
	if err != nil {
		return 0, false
	}
	per := int(mapping.KeysymsPerKeycode)
	for i := 0; i < int(count); i++ {
		for j := 0; j < per; j++ {
			if mapping.Keysyms[i*per+j] == keysym {
				return setup.MinKeycode + xproto.Keycode(i), true
			}
		}
	}
	return 0, false
}

// releaseDesktopBindings empties the desktop settings bound to Print alone,
// after writing them down so that even a crash cannot lose them.
func (p *printKey) releaseDesktopBindings() {
	var released []desktopBinding
	for _, b := range desktopPrintBindings {
		out, err := exec.Command("gsettings", "get", b[0], b[1]).Output()
		value := strings.TrimSpace(string(out))
		if err == nil && strings.Contains(value, "'Print'") {
			released = append(released, desktopBinding{b[0], b[1], value})
		}
	}
	if len(released) == 0 {
		return
	}
	data, _ := json.Marshal(released)
	if os.MkdirAll(filepath.Dir(p.stateFile), 0700) != nil || os.WriteFile(p.stateFile, data, 0600) != nil {
		return // not written down: the desktop keeps its key
	}
	for _, b := range released {
		exec.Command("gsettings", "set", b.Schema, b.Key, "[]").Run()
	}
}

// restoreDesktopBindings gives back what was released, also at the launch
// following a crash.
func (p *printKey) restoreDesktopBindings() {
	data, err := os.ReadFile(p.stateFile)
	if err != nil {
		return
	}
	var released []desktopBinding
	if json.Unmarshal(data, &released) == nil {
		for _, b := range released {
			exec.Command("gsettings", "set", b.Schema, b.Key, b.Value).Run()
		}
	}
	os.Remove(p.stateFile)
}

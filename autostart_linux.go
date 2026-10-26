package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

// The desktop launches at the opening of the session each program described
// in ~/.config/autostart (freedesktop.org specification): Cinnamon, GNOME,
// KDE Plasma, Xfce, MATE…
func autostartFile() string {
	return filepath.Join(xdg.ConfigHome, "autostart", "clipqr.desktop")
}

func autostartAvailable() bool { return true }

// autostartEnabled: the file is there, and not turned off from the desktop's
// startup applications, which keep it with Hidden or X-GNOME-Autostart-enabled.
func autostartEnabled() bool {
	keys, ok := readAutostartFile()
	return ok && keys["Hidden"] != "true" && keys["X-GNOME-Autostart-enabled"] != "false"
}

func autostartProgram() string {
	keys, _ := readAutostartFile()
	return keys["TryExec"]
}

func setAutostart(on bool) error {
	path := autostartFile()
	if !on {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// An AppImage is launched through its file, not through its mounted copy.
	program := os.Getenv("APPIMAGE")
	if program == "" {
		var err error
		if program, err = os.Executable(); err != nil {
			return err
		}
	}
	// TryExec: once ClipQR is uninstalled, the desktop ignores the entry.
	entry := strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=ClipQR",
		"Comment=Historique du presse-papiers et QR codes",
		"TryExec=" + program,
		"Exec=" + desktopQuote(program) + " -" + hiddenFlag,
		"Icon=clipqr",
		"Terminal=false",
		"X-GNOME-Autostart-enabled=true",
	}, "\n") + "\n"
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(entry), 0644)
}

func readAutostartFile() (map[string]string, bool) {
	data, err := os.ReadFile(autostartFile())
	if err != nil {
		return nil, false
	}
	keys := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			keys[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return keys, true
}

// desktopQuote writes a path for the Exec key: in quotes, with the characters
// the specification reserves escaped (a space in the path is then harmless).
func desktopQuote(path string) string {
	return `"` + strings.NewReplacer(`\`, `\\\\`, `"`, `\\"`, "`", "\\\\`", `$`, `\\$`, `%`, `%%`).Replace(path) + `"`
}

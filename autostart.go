package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// hiddenFlag: launched at the opening of the session, ClipQR waits near the
// clock without opening its panel.
const hiddenFlag = "hidden"

// settings keeps what the user answered once.
type settings struct {
	// AutostartAsked: the first launch asked whether ClipQR starts with the session.
	AutostartAsked bool `json:"autostartAsked"`
}

func loadSettings(path string) settings {
	var st settings
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &st)
	}
	return st
}

func (st settings) save(path string) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SetAutostart launches ClipQR, or no longer, at each opening of the session.
// Called by the question of the first launch, the panel and the tray menu;
// each follows the others.
func (s *ClipService) SetAutostart(on bool) error {
	err := setAutostart(on)
	s.mu.Lock()
	s.autostart = autostartEnabled()
	s.settings.AutostartAsked = true
	// Not saved, the question only comes back at the next launch.
	s.settings.save(s.settingsPath)
	enabled := s.autostart
	s.publish()
	s.mu.Unlock()
	// Outside the lock: the tray menu is updated on the interface thread.
	if s.autostartChanged != nil {
		s.autostartChanged(enabled)
	}
	if err != nil {
		return fmt.Errorf("Lancement à l’ouverture de session non modifié : %v", err)
	}
	return nil
}

// repairAutostart: an entry whose program was moved or deleted follows
// ClipQR to where it runs now.
func repairAutostart() {
	if program := autostartProgram(); program != "" && autostartEnabled() {
		if _, err := os.Stat(program); err != nil {
			setAutostart(true)
		}
	}
}

package main

import "errors"

// On macOS, the launch at the opening of the session is left to the user
// (System Settings → General → Login Items).
func autostartAvailable() bool { return false }

func autostartEnabled() bool { return false }

func autostartProgram() string { return "" }

func setAutostart(bool) error {
	return errors.New("à régler dans Réglages Système → Général → Ouverture.")
}

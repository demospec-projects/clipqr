package main

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	// The Startup apps of the Task Manager and of the Settings turn an entry
	// off here, without removing it from runKey.
	approvedKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
	runValue    = "ClipQR"
)

func autostartAvailable() bool { return true }

func autostartEnabled() bool {
	if autostartCommand() == "" {
		return false
	}
	approved, err := registry.OpenKey(registry.CURRENT_USER, approvedKey, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer approved.Close()
	state, _, err := approved.GetBinaryValue(runValue)
	// An odd first byte: turned off by the user.
	return err != nil || len(state) == 0 || state[0]&1 == 0
}

func autostartCommand() string {
	run, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer run.Close()
	command, _, _ := run.GetStringValue(runValue)
	return command
}

// autostartProgram: the command starts with the program, in quotes.
func autostartProgram() string {
	command := autostartCommand()
	if rest, ok := strings.CutPrefix(command, `"`); ok {
		program, _, _ := strings.Cut(rest, `"`)
		return program
	}
	program, _, _ := strings.Cut(command, " ")
	return program
}

func setAutostart(on bool) error {
	run, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer run.Close()
	// Turned on again from ClipQR, the entry is no longer held off by Windows.
	if approved, err := registry.OpenKey(registry.CURRENT_USER, approvedKey, registry.SET_VALUE); err == nil {
		approved.DeleteValue(runValue)
		approved.Close()
	}
	if !on {
		if err := run.DeleteValue(runValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
		return nil
	}
	program, err := os.Executable()
	if err != nil {
		return err
	}
	return run.SetStringValue(runValue, `"`+program+`" -`+hiddenFlag)
}

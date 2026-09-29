package main

import "errors"

// A Mac keyboard has no Print Screen key: macOS takes screenshots with ⇧⌘3,
// which ClipQR collects like any copied image when ⌃ is added (⌃⇧⌘3).
type printKey struct{}

func newPrintKey(string) *printKey { return &printKey{} }

func printKeyAvailable() bool { return false }

func (p *printKey) start(func()) error {
	return errors.New("Les claviers Mac n’ont pas de touche Impr. écran : ⌃⇧⌘3 copie une capture, que ClipQR garde.")
}

func (p *printKey) stop() {}

func (p *printKey) restoreDesktopBindings() {}

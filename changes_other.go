//go:build !linux

package main

import "context"

// clipboardChanges signals each new copy. Windows reports changes as they
// happen; macOS compares a change counter once a second, reading nothing
// until the clipboard really changed.
func clipboardChanges(ctx context.Context) <-chan struct{} {
	return watchChanges(ctx)
}

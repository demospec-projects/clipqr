package main

import (
	"context"

	"golang.design/x/clipboard"
)

// watchChanges turns the library's own watch into the same signal.
func watchChanges(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		for range clipboard.Watch(ctx) {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}()
	return ch
}

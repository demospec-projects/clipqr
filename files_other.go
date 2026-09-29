//go:build !linux

package main

import (
	"context"

	"golang.design/x/clipboard"
)

// writeFiles puts the files on the clipboard the way Explorer and Finder do.
func writeFiles(ctx context.Context, paths []string) error {
	_, err := clipboard.WriteFiles(ctx, paths)
	return err
}

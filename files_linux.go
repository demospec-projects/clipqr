package main

import (
	"context"
	"net/url"
	"strings"

	"golang.design/x/clipboard"
)

// Nemo and Nautilus paste files only from their own format; the list of
// addresses serves the other file managers, and the paths the text editors.
var gnomeCopiedFiles = clipboard.Register("x-special/gnome-copied-files")

func writeFiles(ctx context.Context, paths []string) error {
	uris := make([]string, len(paths))
	for i, p := range paths {
		uris[i] = (&url.URL{Scheme: "file", Path: p}).String()
	}
	_, err := clipboard.WriteAll(ctx,
		clipboard.Item{Format: gnomeCopiedFiles, Bytes: []byte("copy\n" + strings.Join(uris, "\n"))},
		clipboard.Item{Format: clipboard.FmtFiles, Bytes: []byte(strings.Join(uris, "\r\n") + "\r\n")},
		clipboard.Item{Format: clipboard.FmtText, Bytes: []byte(strings.Join(paths, "\n"))},
	)
	return err
}

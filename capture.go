package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/demospec-projects/clipqr/internal/history"
	"github.com/demospec-projects/clipqr/internal/imagestore"
	"github.com/demospec-projects/clipqr/internal/screenshot"
	"golang.design/x/clipboard"
)

// SetCaptureMode gives Print Screen to ClipQR, or back to the system. Called
// by the panel's camera button and by the tray menu; each follows the other.
func (s *ClipService) SetCaptureMode(on bool) error {
	s.mu.Lock()
	unchanged := s.captureMode == on
	s.mu.Unlock()
	if unchanged {
		return nil
	}
	if on {
		if err := s.printKey.start(func() { go s.takeScreenshot() }); err != nil {
			if s.captureModeChanged != nil {
				s.captureModeChanged(false)
			}
			return err
		}
	} else {
		s.printKey.stop()
	}
	s.mu.Lock()
	s.captureMode = on
	s.publish()
	s.mu.Unlock()
	if s.captureModeChanged != nil {
		s.captureModeChanged(on)
	}
	return nil
}

// takeScreenshot saves the whole screen as IE_XXXX_yy_mm_dd_HHhMM.png in the
// screenshots folder, puts it on the clipboard like Print Screen always did,
// and adds it to the list. Pressing again while it is saved does nothing.
func (s *ClipService) takeScreenshot() {
	if !s.capturing.CompareAndSwap(false, true) {
		return
	}
	defer s.capturing.Store(false)

	var img *image.RGBA
	var err error
	grab := func() { img, err = grabScreen() }
	if s.withPanelHidden != nil {
		s.withPanelHidden(grab)
	} else {
		grab()
	}
	if err != nil {
		s.fail(fmt.Sprintf("Capture impossible : %v", err))
		return
	}
	now := time.Now()
	var buf bytes.Buffer
	if err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&buf, img); err != nil {
		s.fail(fmt.Sprintf("Capture impossible : %v", err))
		return
	}
	path := filepath.Join(s.capturesDir, screenshot.FileName(now, nil))
	if err = os.MkdirAll(s.capturesDir, 0755); err == nil {
		err = os.WriteFile(path, buf.Bytes(), 0644)
	}
	if err != nil {
		s.fail(fmt.Sprintf("Capture non enregistrée : %v", err))
		return
	}
	key := imagestore.Fingerprint(img)
	s.images.MakeThumb(key, img)

	s.mu.Lock()
	defer s.mu.Unlock()
	entry := history.Entry{Image: path, ImageKey: key, Width: img.Rect.Dx(), Height: img.Rect.Dy(), Capture: true, CopiedAt: now}
	if _, err = clipboard.Write(context.Background(), clipboard.FmtImage, buf.Bytes()); err == nil {
		s.written, s.writtenAt = entry.Key(), now
		s.current = entry.Key()
	}
	if s.data.AddEntry(entry) {
		s.persist()
	}
	s.publish()
}

var thumbPath = regexp.MustCompile(`^/miniature/([0-9a-f]{24})$`)

// serveThumbnails hands the interface the thumbnails of the images in the list.
func (s *ClipService) serveThumbnails(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := thumbPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			next.ServeHTTP(w, r)
			return
		}
		data, err := os.ReadFile(s.images.ThumbPath(m[1]))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// The address is the fingerprint of the pixels: it never changes.
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=31536000, immutable")
		w.Write(data)
	})
}

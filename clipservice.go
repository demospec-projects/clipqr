package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/demospec-projects/clipqr/internal/history"
	"github.com/demospec-projects/clipqr/internal/imagestore"
	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/clipboard"
)

const stateEvent = "clipqr:state"

// ownImageDelay: an image ClipQR puts on the clipboard can come back
// re-encoded by the system, so any image seen this soon after is its own.
const ownImageDelay = 3 * time.Second

func init() {
	application.RegisterEvent[State](stateEvent)
}

// ClipService watches the clipboard, keeps the history and answers the interface.
type ClipService struct {
	mu          sync.Mutex
	path        string
	images      imagestore.Store
	capturesDir string
	data        history.History
	paused      bool
	current     string // key of the entry on the clipboard
	// written is the key of what ClipQR itself just put on the clipboard: when
	// the watcher sees it come back, the list must not move under the mouse.
	written   string
	writtenAt time.Time
	errText   string
	cancel    context.CancelFunc

	printKey    *printKey
	captureMode bool
	capturing   atomic.Bool

	settingsPath string
	settings     settings
	autostart    bool

	emit               func(State)
	pausedChanged      func(bool)
	captureModeChanged func(bool)
	autostartChanged   func(bool)
	// withPanelHidden runs a screen grab with the panel out of the picture.
	withPanelHidden func(grab func())
}

func newClipService(dataDir, capturesDir string, paused bool) *ClipService {
	s := &ClipService{
		path:        filepath.Join(dataDir, "history.json"),
		images:      imagestore.Store{Dir: filepath.Join(dataDir, "images")},
		capturesDir: capturesDir,
		paused:      paused,
		printKey:    newPrintKey(dataDir),

		settingsPath: filepath.Join(dataDir, "settings.json"),
	}
	var err error
	if s.data, err = history.Load(s.path); err != nil {
		s.errText = fmt.Sprintf("Impossible de charger l’historique : %v", err)
	}
	s.settings = loadSettings(s.settingsPath)
	repairAutostart()
	s.autostart = autostartEnabled()
	return s
}

func (s *ClipService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	// After a crash in capture mode, the desktop gets its Print key back.
	s.printKey.restoreDesktopBindings()
	if err := clipboard.Init(); err != nil {
		s.mu.Lock()
		s.errText = fmt.Sprintf("Presse-papiers inaccessible : %v", err)
		s.mu.Unlock()
		return nil
	}
	watchCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.readClipboard(watchCtx)
	go func() {
		for range clipboardChanges(watchCtx) {
			// An application often puts several formats in turn: let it finish.
			time.Sleep(150 * time.Millisecond)
			s.readClipboard(watchCtx)
		}
	}()
	return nil
}

func (s *ClipService) ServiceShutdown() error {
	s.printKey.stop()
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// readClipboard records what was just copied: files first, then text, then
// an image — a spreadsheet offers its cells both as text and as a picture,
// and the text is what the user copied.
func (s *ClipService) readClipboard(ctx context.Context) {
	s.mu.Lock()
	paused := s.paused
	s.mu.Unlock()
	if paused {
		s.mu.Lock()
		s.current = ""
		s.publish()
		s.mu.Unlock()
		return
	}
	formats, _ := clipboard.Formats(ctx)
	has := func(f clipboard.Format) bool {
		for _, g := range formats {
			if g == f {
				return true
			}
		}
		return false
	}
	switch {
	case has(clipboard.FmtFiles):
		if paths, err := clipboard.ReadFiles(ctx); err == nil && len(paths) > 0 {
			s.record(history.Entry{Files: paths})
		}
	case has(clipboard.FmtText) || len(formats) == 0:
		if text, err := clipboard.Read(ctx, clipboard.FmtText); err == nil {
			s.record(history.Entry{Text: string(text)})
		}
	case has(clipboard.FmtImage):
		s.recordImage(ctx)
	}
}

func (s *ClipService) recordImage(ctx context.Context) {
	s.mu.Lock()
	own := strings.HasPrefix(s.written, "image:") && time.Since(s.writtenAt) < ownImageDelay
	s.mu.Unlock()
	if own {
		return
	}
	data, err := clipboard.Read(ctx, clipboard.FmtImage)
	if err != nil || len(data) > imagestore.MaxBytes {
		return
	}
	path, key, img, err := s.images.Keep(data)
	if err != nil {
		return
	}
	if _, err := os.Stat(s.images.ThumbPath(key)); err != nil {
		s.images.MakeThumb(key, img)
	}
	s.record(history.Entry{Image: path, ImageKey: key, Width: img.Bounds().Dx(), Height: img.Bounds().Dy()})
}

func (s *ClipService) record(entry history.Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := entry.Key()
	if key == s.written {
		return
	}
	s.written = ""
	entry.CopiedAt = time.Now()
	if s.data.AddEntry(entry) {
		s.persist()
	}
	s.current = key
	s.publish()
}

// State returns everything the interface shows.
func (s *ClipService) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state()
}

// Copy puts the entry back on the clipboard, and the panel stays open for the next one.
func (s *ClipService) Copy(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.find(id)
	if !ok {
		return errors.New("Cet élément n’est plus dans l’historique.")
	}
	ctx := context.Background()
	var err error
	switch {
	case len(entry.Files) > 0:
		existing := existingFiles(entry.Files)
		if len(existing) == 0 {
			return errors.New("Fichiers introuvables : ils ont été déplacés ou supprimés.")
		}
		err = writeFiles(ctx, existing)
	case entry.ImageKey != "":
		data, readErr := os.ReadFile(entry.Image)
		if readErr != nil {
			return errors.New("Image introuvable : le fichier a été déplacé ou supprimé.")
		}
		_, err = clipboard.Write(ctx, clipboard.FmtImage, data)
	default:
		_, err = clipboard.Write(ctx, clipboard.FmtText, []byte(entry.Text))
	}
	if err != nil {
		return fmt.Errorf("Copie impossible : %w", err)
	}
	s.written, s.writtenAt = entry.Key(), time.Now()
	s.current = entry.Key()
	s.publish()
	return nil
}

func (s *ClipService) SetPinned(id string, pinned bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.find(id); ok && s.data.SetPinned(entry.Key(), pinned) {
		s.persist()
		s.publish()
	}
}

// SetSecret hides an entry (a password) from whoever looks over the shoulder;
// it is still copied in clear.
func (s *ClipService) SetSecret(id string, secret bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.find(id); ok && s.data.SetSecret(entry.Key(), secret) {
		s.persist()
		s.publish()
	}
}

// Delete forgets one entry, even pinned. Like Clear, it leaves the clipboard
// and the saved screenshots alone; a copied image's own copy goes with it.
func (s *ClipService) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.find(id)
	if !ok || !s.data.Remove(entry.Key()) {
		return
	}
	// Copied again from elsewhere, it must come back to the list.
	if s.written == entry.Key() {
		s.written = ""
	}
	s.persist()
	s.publish()
}

// Clear forgets the unpinned entries; the clipboard itself is left alone, and
// so are the screenshots saved in the screenshots folder.
func (s *ClipService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.ClearUnpinned() > 0 {
		s.persist()
	}
	s.publish()
}

// SetPaused is called by the panel and by the tray menu; each follows the other.
func (s *ClipService) SetPaused(paused bool) {
	s.mu.Lock()
	changed := s.paused != paused
	s.paused = paused
	if changed {
		s.publish()
	}
	s.mu.Unlock()
	// Outside the lock: the tray menu is updated on the interface thread.
	if changed && s.pausedChanged != nil {
		s.pausedChanged(paused)
	}
}

func (s *ClipService) DismissError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errText = ""
	s.publish()
}

// QRCode returns a text as a PNG image, ready for an <img> tag.
func (s *ClipService) QRCode(id string) (string, error) {
	s.mu.Lock()
	entry, ok := s.find(id)
	s.mu.Unlock()
	if !ok || entry.Text == "" {
		return "", errors.New("Seul un texte se met dans un QR code.")
	}
	png, err := qrcode.Encode(entry.Text, qrcode.Medium, 512)
	if err != nil {
		return "", errors.New("Ce texte dépasse la capacité d’un QR code. Copiez un extrait plus court.")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// Open hands a link to the browser, or an email address to the mail client.
func (s *ClipService) Open(id string) error {
	s.mu.Lock()
	entry, ok := s.find(id)
	s.mu.Unlock()
	target := ""
	if ok && !entry.Secret {
		target = history.OpenTarget(entry.Text)
	}
	if target == "" {
		return errors.New("Ce texte ne s’ouvre pas.")
	}
	return application.Get().Browser.OpenURL(target)
}

func (s *ClipService) fail(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errText = message
	s.publish()
}

// persist saves the history, then deletes the copies of images it no longer holds.
func (s *ClipService) persist() {
	if err := s.data.Save(s.path); err != nil {
		s.errText = fmt.Sprintf("Impossible d’enregistrer l’historique : %v", err)
		return
	}
	s.images.Prune(s.data.ImageKeys())
}

func (s *ClipService) publish() {
	if s.emit != nil {
		s.emit(s.state())
	}
}

func (s *ClipService) find(id string) (history.Entry, bool) {
	for _, e := range s.data.Entries {
		if entryID(e) == id {
			return e, true
		}
	}
	return history.Entry{}, false
}

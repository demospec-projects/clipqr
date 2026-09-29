package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/demospec-projects/clipqr/internal/history"
	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/clipboard"
)

const stateEvent = "clipqr:state"

// displayRunes bounds what the interface receives for a long text: it shows
// and searches this much, and acts on the full text through the entry id.
const displayRunes = 2000

func init() {
	application.RegisterEvent[State](stateEvent)
}

type EntryView struct {
	ID        string       `json:"id"`
	Text      string       `json:"text"`
	Truncated bool         `json:"truncated"`
	CopiedAt  *time.Time   `json:"copiedAt"`
	Pinned    bool         `json:"pinned"`
	Kind      history.Kind `json:"kind"`
}

type State struct {
	Entries []EntryView `json:"entries"`
	Paused  bool        `json:"paused"`
	// Current is the id of the entry now on the clipboard, ready for Ctrl+V.
	Current string `json:"current"`
	Error   string `json:"error"`
}

// ClipService watches the clipboard, keeps the history and answers the interface.
type ClipService struct {
	mu      sync.Mutex
	path    string
	data    history.History
	paused  bool
	current string
	// written is the text ClipQR itself just put on the clipboard: when the
	// watcher sees it come back, the list must not move under the user's mouse.
	written string
	errText string
	cancel  context.CancelFunc

	emit          func(State)
	pausedChanged func(bool)
}

func newClipService(path string, paused bool) *ClipService {
	s := &ClipService{path: path, paused: paused}
	var err error
	if s.data, err = history.Load(path); err != nil {
		s.errText = fmt.Sprintf("Impossible de charger l’historique : %v", err)
	}
	return s
}

func (s *ClipService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	if err := clipboard.Init(); err != nil {
		s.mu.Lock()
		s.errText = fmt.Sprintf("Presse-papiers inaccessible : %v", err)
		s.mu.Unlock()
		return nil
	}
	watchCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	if text, err := clipboard.Read(watchCtx, clipboard.FmtText); err == nil {
		s.capture(string(text))
	}
	go func() {
		for data := range clipboard.Watch(watchCtx, clipboard.FmtText) {
			s.capture(string(data.Bytes))
		}
	}()
	return nil
}

func (s *ClipService) ServiceShutdown() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *ClipService) capture(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if text == s.written {
		return
	}
	s.written = ""
	if s.paused {
		s.current = ""
		s.publish()
		return
	}
	if s.data.Add(text, time.Now()) {
		s.persist()
	}
	s.current = text
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
		return errors.New("Ce texte n’est plus dans l’historique.")
	}
	if _, err := clipboard.Write(context.Background(), clipboard.FmtText, []byte(entry.Text)); err != nil {
		return fmt.Errorf("Copie impossible : %w", err)
	}
	s.written = entry.Text
	s.current = entry.Text
	s.publish()
	return nil
}

func (s *ClipService) SetPinned(id string, pinned bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.find(id); ok && s.data.SetPinned(entry.Text, pinned) {
		s.persist()
		s.publish()
	}
}

// Clear forgets the unpinned texts; the clipboard itself is left alone.
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

// QRCode returns the entry as a PNG image, ready for an <img> tag.
func (s *ClipService) QRCode(id string) (string, error) {
	s.mu.Lock()
	entry, ok := s.find(id)
	s.mu.Unlock()
	if !ok {
		return "", errors.New("Ce texte n’est plus dans l’historique.")
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
	if ok {
		target = history.OpenTarget(entry.Text)
	}
	if target == "" {
		return errors.New("Ce texte ne s’ouvre pas.")
	}
	return application.Get().Browser.OpenURL(target)
}

func (s *ClipService) persist() {
	if err := s.data.Save(s.path); err != nil {
		s.errText = fmt.Sprintf("Impossible d’enregistrer l’historique : %v", err)
	}
}

func (s *ClipService) publish() {
	if s.emit != nil {
		s.emit(s.state())
	}
}

func (s *ClipService) state() State {
	st := State{Entries: []EntryView{}, Paused: s.paused, Error: s.errText}
	for _, e := range s.data.Entries {
		view := EntryView{ID: entryID(e.Text), Text: e.Text, Pinned: e.Pinned, Kind: history.KindOf(e.Text)}
		if utf8.RuneCountInString(e.Text) > displayRunes {
			view.Text = string([]rune(e.Text)[:displayRunes])
			view.Truncated = true
		}
		if !e.CopiedAt.IsZero() {
			at := e.CopiedAt
			view.CopiedAt = &at
		}
		if e.Text == s.current {
			st.Current = view.ID
		}
		st.Entries = append(st.Entries, view)
	}
	return st
}

func (s *ClipService) find(id string) (history.Entry, bool) {
	for _, e := range s.data.Entries {
		if entryID(e.Text) == id {
			return e, true
		}
	}
	return history.Entry{}, false
}

func entryID(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:8])
}

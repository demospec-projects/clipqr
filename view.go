package main

import (
	"fmt"
	"hash/maphash"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/demospec-projects/clipqr/internal/history"
)

// displayRunes bounds what the interface receives for a long text: it shows
// and searches this much, and acts on the full text through the entry id.
const displayRunes = 2000

type EntryView struct {
	ID        string       `json:"id"`
	Text      string       `json:"text"`
	Truncated bool         `json:"truncated"`
	CopiedAt  *time.Time   `json:"copiedAt"`
	Pinned    bool         `json:"pinned"`
	Kind      history.Kind `json:"kind"`
	// Secret: nothing of the content is sent to the interface, only its mark.
	Secret bool `json:"secret"`
	// Files: the copied paths; Missing counts those no longer found.
	Files   []string `json:"files"`
	Missing int      `json:"missing"`
	Folder  bool     `json:"folder"`
	// Place is the folder the files come from, home shortened to ~.
	Place string `json:"place"`
	// Thumb is the address of the image's thumbnail; Name, a screenshot's file name.
	Thumb   string `json:"thumb"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Capture bool   `json:"capture"`
	Name    string `json:"name"`
}

type State struct {
	Entries []EntryView `json:"entries"`
	Paused  bool        `json:"paused"`
	// Current is the id of the entry now on the clipboard, ready for Ctrl+V.
	Current string `json:"current"`
	Error   string `json:"error"`
	// CaptureMode: Print Screen belongs to ClipQR, which saves each screenshot
	// in CapturesDir. CaptureAvailable is false where the key cannot be taken.
	CaptureMode      bool   `json:"captureMode"`
	CaptureAvailable bool   `json:"captureAvailable"`
	CapturesDir      string `json:"capturesDir"`
	// Autostart: ClipQR is launched at each opening of the session. The panel
	// asks once (AskAutostart), then the choice is changed from the panel or
	// the tray menu.
	Autostart          bool `json:"autostart"`
	AutostartAvailable bool `json:"autostartAvailable"`
	AskAutostart       bool `json:"askAutostart"`
}

// state builds what the interface shows. Called with the lock held.
func (s *ClipService) state() State {
	st := State{
		Entries: []EntryView{}, Paused: s.paused, Error: s.errText,
		CaptureMode: s.captureMode, CaptureAvailable: printKeyAvailable(), CapturesDir: displayPath(s.capturesDir),
		Autostart: s.autostart, AutostartAvailable: autostartAvailable(),
		AskAutostart: autostartAvailable() && !s.autostart && !s.settings.AutostartAsked,
	}
	for _, e := range s.data.Entries {
		view := EntryView{ID: entryID(e), Pinned: e.Pinned, Kind: history.KindOfEntry(e), Secret: e.Secret}
		if !e.CopiedAt.IsZero() {
			at := e.CopiedAt
			view.CopiedAt = &at
		}
		switch view.Kind {
		case history.KindFiles:
			view.Missing = len(e.Files) - len(existingFiles(e.Files))
			info, err := os.Stat(e.Files[0])
			view.Folder = len(e.Files) == 1 && err == nil && info.IsDir()
			if !e.Secret {
				view.Files = e.Files
				view.Place = displayPath(filepath.Dir(e.Files[0]))
			}
		case history.KindImage:
			view.Width, view.Height, view.Capture = e.Width, e.Height, e.Capture
			if _, err := os.Stat(e.Image); err != nil {
				view.Missing = 1
			}
			if !e.Secret {
				view.Thumb = "/miniature/" + e.ImageKey
				if e.Capture {
					view.Name = strings.TrimSuffix(filepath.Base(e.Image), ".png")
				}
			}
		default:
			if e.Secret {
				view.Kind = history.KindText
			} else {
				view.Text = cutRunes(e.Text, displayRunes)
				view.Truncated = len(view.Text) < len(e.Text)
			}
		}
		if e.HasKey(s.current) {
			st.Current = view.ID
		}
		st.Entries = append(st.Entries, view)
	}
	return st
}

// idSeed makes the ids of this session; they are never written down.
var idSeed = maphash.MakeSeed()

// entryID names an entry for the interface, which only ever sees ids of the
// running session: a fast hash of the key, read in place.
func entryID(e history.Entry) string {
	var h maphash.Hash
	h.SetSeed(idSeed)
	prefix, body := e.KeyParts()
	h.WriteString(prefix)
	h.WriteString(body)
	return fmt.Sprintf("%016x", h.Sum64())
}

func existingFiles(paths []string) []string {
	var found []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}

// displayPath shortens the home folder to ~ for the interface.
func displayPath(path string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if rest, ok := strings.CutPrefix(path, home); ok {
			return "~" + rest
		}
	}
	return path
}

// cutRunes keeps the first n runes of a text, without copying it: the list
// may hold a hundred long texts, and the state is sent at every change.
func cutRunes(text string, n int) string {
	for i := range text {
		if n == 0 {
			return text[:i]
		}
		n--
	}
	return text
}

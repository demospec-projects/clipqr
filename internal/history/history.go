package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Limit is the number of unpinned texts kept; pinned texts do not count.
const Limit = 100
const MaxTextBytes = 256 * 1024

type Entry struct {
	Text     string    `json:"text"`
	CopiedAt time.Time `json:"copiedAt"`
	Pinned   bool      `json:"pinned,omitempty"`
}

// History lists texts from the most recently copied to the oldest.
type History struct {
	Entries []Entry `json:"entries"`
}

// Add puts text at the top, or moves it there if it is already known.
// It reports whether the history changed.
func (h *History) Add(text string, at time.Time) bool {
	return h.add(Entry{Text: text, CopiedAt: at})
}

func (h *History) add(entry Entry) bool {
	if strings.TrimSpace(entry.Text) == "" || len(entry.Text) > MaxTextBytes {
		return false
	}
	if len(h.Entries) > 0 && h.Entries[0].Text == entry.Text {
		return false
	}
	entries := []Entry{entry}
	for _, old := range h.Entries {
		if old.Text == entry.Text {
			// A text copied again keeps its pin.
			entries[0].Pinned = entries[0].Pinned || old.Pinned
			continue
		}
		entries = append(entries, old)
	}
	h.Entries = entries
	h.trim()
	return true
}

func (h *History) SetPinned(text string, pinned bool) bool {
	for i := range h.Entries {
		if h.Entries[i].Text == text {
			h.Entries[i].Pinned = pinned
			h.trim()
			return true
		}
	}
	return false
}

// ClearUnpinned forgets every text that is not pinned and returns how many were removed.
func (h *History) ClearUnpinned() int {
	kept := []Entry{}
	for _, e := range h.Entries {
		if e.Pinned {
			kept = append(kept, e)
		}
	}
	removed := len(h.Entries) - len(kept)
	h.Entries = kept
	return removed
}

// trim drops the oldest unpinned texts beyond Limit.
func (h *History) trim() {
	kept := []Entry{}
	unpinned := 0
	for _, e := range h.Entries {
		if !e.Pinned {
			if unpinned == Limit {
				continue
			}
			unpinned++
		}
		kept = append(kept, e)
	}
	h.Entries = kept
}

// Load reads the history file. It also reads the format of the first
// Windows version ({"items": [...]}), whose texts have no copy time.
func Load(path string) (History, error) {
	var h History
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return h, nil
	}
	if err != nil {
		return h, err
	}
	var saved struct {
		Entries []Entry  `json:"entries"`
		Items   []string `json:"items"`
	}
	if err = json.Unmarshal(data, &saved); err != nil {
		return h, err
	}
	for i := len(saved.Items) - 1; i >= 0; i-- {
		h.Add(saved.Items[i], time.Time{})
	}
	for i := len(saved.Entries) - 1; i >= 0; i-- {
		h.add(saved.Entries[i])
	}
	return h, nil
}

// Save writes the file in one step, so a crash never leaves it half written.
func (h *History) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "history-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

package history

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)

func texts(h History) []string {
	var out []string
	for _, e := range h.Entries {
		out = append(out, e.Text)
	}
	return out
}

func TestHistoryMoveAndBound(t *testing.T) {
	var h History
	for i := 0; i < 105; i++ {
		h.Add(fmt.Sprint(i), now)
	}
	if len(h.Entries) != 100 || h.Entries[0].Text != "104" || h.Entries[99].Text != "5" {
		t.Fatal("history bound/order", texts(h))
	}
	h.Add("50", now)
	if h.Entries[0].Text != "50" || len(h.Entries) != 100 {
		t.Fatal("move failed")
	}
	seen := map[string]bool{}
	for _, s := range texts(h) {
		if seen[s] {
			t.Fatal("duplicate", s)
		}
		seen[s] = true
	}
	if h.Add("50", now) || h.Add(" \n\t", now) {
		t.Fatal("unchanged or empty text accepted")
	}
}

func TestPinnedTextsSurviveLimitAndClear(t *testing.T) {
	var h History
	h.Add("épinglé", now)
	h.SetPinned("épinglé", true)
	for i := 0; i < 150; i++ {
		h.Add(fmt.Sprint(i), now)
	}
	if len(h.Entries) != 101 || h.Entries[100].Text != "épinglé" {
		t.Fatal("pinned text dropped by the limit", len(h.Entries))
	}
	h.Add("épinglé", now)
	if !h.Entries[0].Pinned {
		t.Fatal("copying a pinned text again lost its pin")
	}
	if removed := h.ClearUnpinned(); removed != 100 || len(h.Entries) != 1 {
		t.Fatal("clear should keep pinned texts only", removed, texts(h))
	}
	if h.SetPinned("inconnu", true) {
		t.Fatal("pinned an unknown text")
	}
}

func TestUnpinningTrimsToLimit(t *testing.T) {
	var h History
	for i := 0; i < 100; i++ {
		h.Add(fmt.Sprint(i), now)
	}
	h.SetPinned("0", true)
	h.Add("100", now)
	h.SetPinned("0", false)
	if len(h.Entries) != 100 || h.Entries[99].Text == "0" {
		t.Fatal("unpinned text should fall off the end", texts(h))
	}
}

func TestPersistencePreservesText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "history.json")
	h := History{}
	text := "  Français 🦊\r\n deuxième ligne\t "
	h.Add(text, now)
	h.SetPinned(text, true)
	if err := h.Save(p); err != nil {
		t.Fatal(err)
	}
	h.Add("autre", now.Add(time.Minute))
	if err := h.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 2 || got.Entries[1].Text != text || !got.Entries[1].Pinned || !got.Entries[0].CopiedAt.Equal(now.Add(time.Minute)) {
		t.Fatal("history was changed", got)
	}
}

func TestLoadsFirstWindowsFormat(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(p, []byte(`{"items":["récent","ancien","récent"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(texts(got)) != "[récent ancien]" || !got.Entries[0].CopiedAt.IsZero() {
		t.Fatal("old format not read", got)
	}
}

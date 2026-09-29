package history

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)

func key(text string) string { return Entry{Text: text}.Key() }

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
	h.SetPinned(key("épinglé"), true)
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
	if h.SetPinned(key("inconnu"), true) {
		t.Fatal("pinned an unknown text")
	}
}

func TestUnpinningTrimsToLimit(t *testing.T) {
	var h History
	for i := 0; i < 100; i++ {
		h.Add(fmt.Sprint(i), now)
	}
	h.SetPinned(key("0"), true)
	h.Add("100", now)
	h.SetPinned(key("0"), false)
	if len(h.Entries) != 100 || h.Entries[99].Text == "0" {
		t.Fatal("unpinned text should fall off the end", texts(h))
	}
}

func TestSecretTextStaysHidden(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.json")
	var h History
	h.Add("motdepasse", now)
	h.SetSecret(key("motdepasse"), true)
	h.Add("autre", now)
	h.Add("motdepasse", now)
	if !h.Entries[0].Secret || h.Entries[1].Secret {
		t.Fatal("copying a hidden text again showed it", h.Entries)
	}
	if err := h.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Entries[0].Secret {
		t.Fatal("hidden flag lost on reload")
	}
	if !got.SetSecret(key("motdepasse"), false) || got.Entries[0].Secret || got.SetSecret(key("inconnu"), true) {
		t.Fatal("SetSecret")
	}
}

func TestPersistencePreservesText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "history.json")
	h := History{}
	text := "  Français 🦊\r\n deuxième ligne\t "
	h.Add(text, now)
	h.SetPinned(key(text), true)
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

func TestFilesAndImagesAreEntriesToo(t *testing.T) {
	var h History
	files := Entry{Files: []string{"/tmp/a.pdf", "/tmp/b.png"}, CopiedAt: now}
	image := Entry{Image: "/tmp/i.png", ImageKey: "abc", Width: 10, Height: 5, CopiedAt: now}
	if !h.AddEntry(files) || !h.AddEntry(image) || h.AddEntry(Entry{Files: []string{}}) || h.AddEntry(Entry{ImageKey: "x"}) {
		t.Fatal("files and images accepted, empty ones refused")
	}
	h.Add("/tmp/a.pdf\n/tmp/b.png", now)
	if len(h.Entries) != 3 {
		t.Fatal("a text is never mistaken for the files it names", len(h.Entries))
	}
	h.AddEntry(Entry{Files: []string{"/tmp/a.pdf", "/tmp/b.png"}, CopiedAt: now})
	if h.Entries[0].Key() != files.Key() || len(h.Entries) != 3 {
		t.Fatal("the same files copied again move to the top")
	}
	if !h.ImageKeys()["abc"] || len(h.ImageKeys()) != 1 {
		t.Fatal("ImageKeys", h.ImageKeys())
	}
}

func TestScreenshotKeepsItsFileWhenCopiedAgain(t *testing.T) {
	var h History
	h.AddEntry(Entry{Image: "/ecrans/IE_AB12.png", ImageKey: "k", Capture: true, CopiedAt: now})
	h.Add("autre", now)
	h.AddEntry(Entry{Image: "/clipqr/images/k.png", ImageKey: "k", CopiedAt: now})
	if e := h.Entries[0]; !e.Capture || e.Image != "/ecrans/IE_AB12.png" {
		t.Fatal("a screenshot copied back must stay its saved file", e)
	}
}

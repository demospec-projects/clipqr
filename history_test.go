package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestHistoryMoveAndBound(t *testing.T) {
	var h history
	for i := 0; i < 105; i++ {
		h.add(fmt.Sprint(i))
	}
	if len(h.Items) != 100 || h.Items[0] != "104" || h.Items[99] != "5" {
		t.Fatal("history bound/order", h)
	}
	h.add("50")
	if h.Items[0] != "50" || len(h.Items) != 100 {
		t.Fatal("move failed")
	}
	seen := map[string]bool{}
	for _, s := range h.Items {
		if seen[s] {
			t.Fatal("duplicate", s)
		}
		seen[s] = true
	}
	if h.add("50") || h.add(" \n\t") {
		t.Fatal("unchanged or empty text accepted")
	}
}
func TestPersistencePreservesText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "history.json")
	h := history{}
	text := "  Français 🦊\r\n deuxième ligne\t "
	h.add(text)
	if err := h.save(p); err != nil {
		t.Fatal(err)
	}
	h.add("autre")
	if err := h.save(p); err != nil {
		t.Fatal(err)
	}
	got, err := loadHistory(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[1] != text {
		t.Fatal("text was changed", got)
	}
}

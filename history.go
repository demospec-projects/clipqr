package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const historyLimit = 100
const maxTextBytes = 256 * 1024

type history struct {
	Items []string `json:"items"`
}

func (h *history) add(text string) bool {
	if strings.TrimSpace(text) == "" || len(text) > maxTextBytes {
		return false
	}
	if len(h.Items) > 0 && h.Items[0] == text {
		return false
	}
	items := []string{text}
	for _, old := range h.Items {
		if old != text && len(items) < historyLimit {
			items = append(items, old)
		}
	}
	h.Items = items
	return true
}

func loadHistory(path string) (history, error) {
	var h history
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return h, nil
	}
	if err != nil {
		return h, err
	}
	var saved history
	if err = json.Unmarshal(data, &saved); err != nil {
		return h, err
	}
	for i := len(saved.Items) - 1; i >= 0; i-- {
		h.add(saved.Items[i])
	}
	return h, nil
}

func (h *history) save(path string) error {
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

func preview(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	r := []rune(text)
	if len(r) > 100 {
		return string(r[:100]) + "…"
	}
	return text
}

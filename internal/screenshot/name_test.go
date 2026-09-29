package screenshot

import (
	"bytes"
	"regexp"
	"testing"
	"time"
)

func TestFileName(t *testing.T) {
	at := time.Date(2026, 9, 28, 13, 32, 45, 0, time.Local)
	if got := FileName(at, bytes.NewReader([]byte{0, 1, 25, 35})); got != "IE_ABZ9_26_09_28_13h32.png" {
		t.Fatal(got)
	}
	pattern := regexp.MustCompile(`^IE_[A-Z0-9]{4}_26_09_28_13h32\.png$`)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		name := FileName(at, nil)
		if !pattern.MatchString(name) {
			t.Fatal("bad name", name)
		}
		seen[name] = true
	}
	if len(seen) < 45 {
		t.Fatal("random part repeats too often", len(seen))
	}
}

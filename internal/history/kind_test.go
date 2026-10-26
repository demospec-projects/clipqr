package history

import (
	"strings"
	"testing"
)

func TestKindOf(t *testing.T) {
	cases := map[string]Kind{
		"https://demospec.ca/chantiers?id=9001": KindLink,
		"  www.demospec.ca  ":                   KindLink,
		"http://localhost:9000/":                KindLink,
		"https://demospec":                      KindText,
		"voir https://demospec.ca":              KindText,
		"bmostafa@demospec.ca":                  KindEmail,
		"mailto:bmostafa@demospec.ca":           KindEmail,
		"514 555-1234":                          KindPhone,
		"+1 (514) 555-1234":                     KindPhone,
		"2026-09-29":                            KindText,
		"12.50":                                 KindText,
		"Bonjour,\nà demain":                    KindText,
	}
	for text, want := range cases {
		if got := KindOf(text); got != want {
			t.Errorf("KindOf(%q) = %s, want %s", text, got, want)
		}
	}
}

func TestOpenTarget(t *testing.T) {
	cases := map[string]string{
		"www.demospec.ca":      "https://www.demospec.ca",
		"bmostafa@demospec.ca": "mailto:bmostafa@demospec.ca",
		"514 555-1234":         "",
		"du texte":             "",
	}
	for text, want := range cases {
		if got := OpenTarget(text); got != want {
			t.Errorf("OpenTarget(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestLongTextIsJustText(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 1500)
	if KindOf(long) != KindLink || OpenTarget(long) != long {
		t.Fatal("a long link is still a link")
	}
	tooLong := "https://example.com/" + strings.Repeat("a", 5000)
	if KindOf(tooLong) != KindText || OpenTarget(tooLong) != "" {
		t.Fatal("a text too long for a link should be plain text")
	}
}

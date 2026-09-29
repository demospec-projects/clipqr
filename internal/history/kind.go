package history

import (
	"net/url"
	"regexp"
	"strings"
)

// Kind tells the interface which icon to show and whether the text can be opened.
type Kind string

const (
	KindText  Kind = "text"
	KindLink  Kind = "link"
	KindEmail Kind = "email"
	KindPhone Kind = "phone"
	KindFiles Kind = "files"
	KindImage Kind = "image"
)

// KindOfEntry: files and images first, then what the text looks like.
func KindOfEntry(e Entry) Kind {
	switch {
	case len(e.Files) > 0:
		return KindFiles
	case e.ImageKey != "":
		return KindImage
	}
	return KindOf(e.Text)
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
var phonePattern = regexp.MustCompile(`^\+?[0-9 ().-]+$`)

func KindOf(text string) Kind {
	t := strings.TrimSpace(text)
	switch {
	case linkTarget(t) != "":
		return KindLink
	case emailPattern.MatchString(strings.TrimPrefix(t, "mailto:")):
		return KindEmail
	case isPhone(t):
		return KindPhone
	}
	return KindText
}

// OpenTarget returns the address to hand to the system for a link or an
// email, or "" when the text is not something that opens.
func OpenTarget(text string) string {
	t := strings.TrimSpace(text)
	switch KindOf(t) {
	case KindLink:
		return linkTarget(t)
	case KindEmail:
		return "mailto:" + strings.TrimPrefix(t, "mailto:")
	}
	return ""
}

func linkTarget(t string) string {
	if strings.ContainsAny(t, " \t\r\n") {
		return ""
	}
	lower := strings.ToLower(t)
	if strings.HasPrefix(lower, "www.") {
		t = "https://" + t
		lower = "https://" + lower
	}
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return ""
	}
	u, err := url.Parse(t)
	if err != nil || !strings.Contains(u.Host, ".") && u.Hostname() != "localhost" {
		return ""
	}
	return t
}

// isPhone accepts 10 to 15 digits, so that dates and amounts are not taken for numbers.
func isPhone(t string) bool {
	if !phonePattern.MatchString(t) {
		return false
	}
	digits := 0
	for _, r := range t {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= 10 && digits <= 15
}

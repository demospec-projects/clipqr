// Package screenshot names the screenshots ClipQR takes with Print Screen.
package screenshot

import (
	"crypto/rand"
	"io"
	"time"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// FileName gives IE_XXXX_26_09_28_13h32.png: IE for « impression écran »,
// four random characters so that screenshots taken in the same minute never
// collide, then the date (yy_mm_dd) and the time.
func FileName(at time.Time, random io.Reader) string {
	if random == nil {
		random = rand.Reader
	}
	b := make([]byte, 4)
	if _, err := io.ReadFull(random, b); err != nil {
		b = []byte{byte(at.Nanosecond()), byte(at.Nanosecond() >> 8), byte(at.Nanosecond() >> 16), byte(at.Nanosecond() >> 24)}
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return "IE_" + string(b) + "_" + at.Format("06_01_02_15h04") + ".png"
}

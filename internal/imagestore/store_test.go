package imagestore

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func sample() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.NRGBA{uint8(x), uint8(y), 90, 255})
		}
	}
	return img
}

func TestFingerprintIgnoresEncoding(t *testing.T) {
	a := sample()
	b := image.NewRGBA(a.Rect)
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			b.Set(x, y, a.At(x, y))
		}
	}
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatal("same pixels, different fingerprints")
	}
	b.Set(3, 3, color.RGBA{0, 0, 0, 255})
	if Fingerprint(a) == Fingerprint(b) {
		t.Fatal("different pixels, same fingerprint")
	}
}

func TestKeepThumbAndPrune(t *testing.T) {
	s := Store{Dir: filepath.Join(t.TempDir(), "images")}
	var buf bytes.Buffer
	png.Encode(&buf, sample())
	path, key, img, err := s.Keep(buf.Bytes())
	if err != nil || img.Bounds().Dx() != 640 {
		t.Fatal(err)
	}
	if err := s.MakeThumb(key, img); err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open(s.ThumbPath(key))
	thumb, err := png.Decode(f)
	f.Close()
	if err != nil || thumb.Bounds().Dx() != 160 || thumb.Bounds().Dy() != 90 {
		t.Fatal("thumbnail keeps the proportions within 160×100", thumb.Bounds(), err)
	}
	s.Prune(map[string]bool{key: true})
	if _, err := os.Stat(path); err != nil {
		t.Fatal("kept image removed")
	}
	s.Prune(map[string]bool{})
	if _, err := os.Stat(path); err == nil {
		t.Fatal("image out of the history not removed")
	}
	if _, err := os.Stat(s.ThumbPath(key)); err == nil {
		t.Fatal("thumbnail out of the history not removed")
	}
}

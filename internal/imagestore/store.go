// Package imagestore keeps the images copied to the clipboard and their
// thumbnails, in ClipQR's own folder.
package imagestore

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// MaxBytes: a bigger image is not kept (a screenshot of three 4K screens is far below).
const MaxBytes = 48 << 20

const thumbWidth, thumbHeight = 160, 100

type Store struct {
	Dir string
}

// Fingerprint identifies an image by its pixels, whatever the encoding the
// clipboard gave it in (the same image comes back RGBA or NRGBA).
func Fingerprint(img image.Image) string {
	rgba := rgbaOf(img)
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, [2]int32{int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy())})
	for y := 0; y < rgba.Rect.Dy(); y++ {
		h.Write(rgba.Pix[y*rgba.Stride : y*rgba.Stride+rgba.Rect.Dx()*4])
	}
	return hex.EncodeToString(h.Sum(nil)[:12])
}

// Keep saves a copied image (PNG bytes) under its fingerprint.
func (s Store) Keep(data []byte) (path, key string, img image.Image, err error) {
	img, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		return "", "", nil, err
	}
	key = Fingerprint(img)
	path = filepath.Join(s.Dir, key+".png")
	if _, statErr := os.Stat(path); statErr != nil {
		err = writeFile(path, data)
	}
	return path, key, img, err
}

func (s Store) ThumbPath(key string) string {
	return filepath.Join(s.Dir, key+"-mini.png")
}

// MakeThumb writes the small version the interface shows in the list.
func (s Store) MakeThumb(key string, img image.Image) error {
	b := img.Bounds()
	scale := min(float64(thumbWidth)/float64(b.Dx()), float64(thumbHeight)/float64(b.Dy()), 1)
	w, h := max(1, int(float64(b.Dx())*scale)), max(1, int(float64(b.Dy())*scale))
	// A screenshot is twenty times too big for the filter alone (a tenth of
	// a second per 4K screen): averaging blocks first leaves it a short way
	// to go, as smooth and ten times faster.
	if factor := int(0.5 / scale); factor >= 2 && b.Dx() >= factor && b.Dy() >= factor {
		img = shrink(rgbaOf(img), factor)
		b = img.Bounds()
	}
	small := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(small, small.Rect, img, b, xdraw.Over, nil)
	var buf bytes.Buffer
	if err := png.Encode(&buf, small); err != nil {
		return err
	}
	return writeFile(s.ThumbPath(key), buf.Bytes())
}

// Prune deletes the images and thumbnails no longer in the history. Only
// ClipQR's folder is touched: saved screenshots are the user's files.
func (s Store) Prune(keep map[string]bool) {
	files, _ := os.ReadDir(s.Dir)
	for _, f := range files {
		key := strings.TrimSuffix(strings.TrimSuffix(f.Name(), ".png"), "-mini")
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".png") && !keep[key] {
			os.Remove(filepath.Join(s.Dir, f.Name()))
		}
	}
}

// rgbaOf gives the image as RGBA with its origin at (0, 0), the layout the
// fingerprint and the shrink read directly.
func rgbaOf(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok && rgba.Rect.Min == (image.Point{}) {
		return rgba
	}
	rgba := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(rgba, rgba.Rect, img, img.Bounds().Min, draw.Src)
	return rgba
}

// shrink averages each factor×factor block of pixels into one.
func shrink(img *image.RGBA, factor int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, img.Rect.Dx()/factor, img.Rect.Dy()/factor))
	n := uint32(factor * factor)
	for y := 0; y < out.Rect.Dy(); y++ {
		for x := 0; x < out.Rect.Dx(); x++ {
			var sum [4]uint32
			for dy := 0; dy < factor; dy++ {
				row := img.Pix[(y*factor+dy)*img.Stride+x*factor*4:]
				for dx := 0; dx < factor; dx++ {
					for c := range sum {
						sum[c] += uint32(row[dx*4+c])
					}
				}
			}
			o := y*out.Stride + x*4
			for c, v := range sum {
				out.Pix[o+c] = uint8(v / n)
			}
		}
	}
	return out
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

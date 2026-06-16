package png

import (
	"bytes"
	"image"
	"image/color"
	gopng "image/png"
	"testing"
)

var h = handler{}

func makeMinimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	if err := gopng.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestValidate_valid(t *testing.T) {
	data := makeMinimalPNG(t)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid PNG: %v", err)
	}
}

func TestValidate_badMagic(t *testing.T) {
	data := []byte("NOTPNG\x00\x00\x00\x00IEND\xae\x42\x60\x82")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestValidate_missingIEND(t *testing.T) {
	data := makeMinimalPNG(t)
	// Truncate last 12 bytes (IEND chunk).
	data = data[:len(data)-12]
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for missing IEND")
	}
}

func TestValidate_tooSmall(t *testing.T) {
	data := []byte{0x89, 0x50}
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for too-small file")
	}
}

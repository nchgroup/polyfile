package jpg

import (
	"bytes"
	"image"
	"image/color"
	gojpeg "image/jpeg"
	"testing"
)

var h = handler{}

func makeMinimalJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	var buf bytes.Buffer
	if err := gojpeg.Encode(&buf, img, &gojpeg.Options{Quality: 1}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestValidate_valid(t *testing.T) {
	data := makeMinimalJPEG(t)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid JPEG: %v", err)
	}
}

func TestValidate_badMagic(t *testing.T) {
	// Valid EOI at end but wrong magic.
	data := []byte("NOTJPEG\xff\xd9")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestValidate_missingEOI(t *testing.T) {
	data := makeMinimalJPEG(t)
	// Replace EOI bytes so they can't be found.
	data[len(data)-2] = 0x00
	data[len(data)-1] = 0x00
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for missing EOI")
	}
}

func TestValidate_tooSmall(t *testing.T) {
	data := []byte{0xff, 0xd8}
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for too-small file")
	}
}

func TestValidate_eoiBeforeEnd(t *testing.T) {
	// EOI followed by extra bytes (simulates polyglot with appended ZIP).
	data := makeMinimalJPEG(t)
	data = append(data, []byte("PK\x05\x06extra bytes")...)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid with appended data: %v", err)
	}
}

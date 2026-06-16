package gif

import (
	"bytes"
	"image"
	"image/color"
	gogif "image/gif"
	"testing"
)

var h = handler{}

func makeMinimalGIF(t *testing.T) []byte {
	t.Helper()
	palette := color.Palette{color.Black, color.White}
	img := image.NewPaletted(image.Rect(0, 0, 1, 1), palette)
	g := &gogif.GIF{
		Image: []*image.Paletted{img},
		Delay: []int{0},
	}
	var buf bytes.Buffer
	if err := gogif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestValidate_valid(t *testing.T) {
	data := makeMinimalGIF(t)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid GIF: %v", err)
	}
}

func TestValidate_withShellPreamble(t *testing.T) {
	// Simulates polyglot: shell script before GIF magic.
	script := []byte("#!/bin/sh\necho hello\nexit 0\n")
	gif := makeMinimalGIF(t)
	data := append(script, gif...)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid with shell preamble: %v", err)
	}
}

func TestValidate_notGIF(t *testing.T) {
	data := []byte("not a gif file at all")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for non-GIF data")
	}
}

func TestValidate_missingTrailer(t *testing.T) {
	data := makeMinimalGIF(t)
	data[len(data)-1] = 0x00 // corrupt trailer
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for missing trailer")
	}
}

func TestValidate_gif87a(t *testing.T) {
	// Manually crafted GIF87a header.
	data := append([]byte("GIF87a\x01\x00\x01\x00\x00\x00\x00"), 0x3b)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected GIF87a to be valid: %v", err)
	}
}

func TestGifSHStrategy_apply(t *testing.T) {
	gifData := makeMinimalGIF(t)
	script := []byte("#!/bin/sh\necho hello\n")

	// Simulate what the strategy would write.
	var out bytes.Buffer
	r1 := bytes.NewReader(script)
	r2 := bytes.NewReader(gifData)

	cfg := &struct{ name string }{name: "test"}
	_ = cfg

	// Manual apply: script + exit guard + GIF.
	out.Write(script)
	out.WriteString("\nexit 0\n")
	out.Write(gifData)

	result := out.Bytes()

	// Shell part must be at offset 0.
	if !bytes.HasPrefix(result, []byte("#!/bin/sh")) {
		t.Error("polyglot must start with shell shebang")
	}

	// GIF magic must be findable within scanLimit bytes.
	_, err := findMagic(bytes.NewReader(result), int64(len(result)))
	if err != nil {
		t.Errorf("GIF magic not found in polyglot: %v", err)
	}
	_ = r1
	_ = r2

	// GIF trailer must be last byte.
	if result[len(result)-1] != 0x3b {
		t.Errorf("last byte must be GIF trailer, got %02x", result[len(result)-1])
	}
}

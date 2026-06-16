package mp3

import (
	"bytes"
	"testing"
)

var h = handler{}

// minID3 builds a minimal ID3v2.3 header with empty tag body (size=0).
func minID3(extra ...byte) []byte {
	// ID3 + version 2.3.0 + no flags + syncsafe size 0
	hdr := []byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	return append(hdr, extra...)
}

func TestValidate_id3v2(t *testing.T) {
	data := minID3(0xff, 0xfb, 0x90, 0x00) // ID3v2 header + fake audio frame sync
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid ID3v2 MP3: %v", err)
	}
}

func TestValidate_syncWord(t *testing.T) {
	// Raw MP3 without ID3 tag, starts with sync word (MPEG1 Layer3).
	data := []byte{0xff, 0xfb, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid sync-word MP3: %v", err)
	}
}

func TestValidate_notMP3(t *testing.T) {
	data := []byte("NOTANMP3file00000000")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for non-MP3 data")
	}
}

func TestValidate_tooSmall(t *testing.T) {
	data := []byte("ID3")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for too-small file")
	}
}

func TestValidate_badVersion(t *testing.T) {
	data := []byte{'I', 'D', '3', 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for invalid ID3v2 version 0xFF")
	}
}

func TestValidate_badSyncsafe(t *testing.T) {
	// Size byte 6 has high bit set → not syncsafe.
	data := []byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00}
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for non-syncsafe size")
	}
}

func TestDecodeSyncsafe(t *testing.T) {
	// Known value: 0x00 0x00 0x02 0x01 = 257 bytes (2<<7 | 1)
	got := DecodeSyncsafe([]byte{0x00, 0x00, 0x02, 0x01})
	if got != 257 {
		t.Errorf("DecodeSyncsafe: got %d, want 257", got)
	}
}

func TestTagSize(t *testing.T) {
	// Tag body size = 100 bytes → total = 10 + 100 = 110.
	// Syncsafe encoding of 100 = 0x00 0x00 0x00 0x64.
	data := []byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64}
	data = append(data, make([]byte, 100)...) // tag body
	r := bytes.NewReader(data)
	sz, err := TagSize(r)
	if err != nil {
		t.Fatalf("TagSize: %v", err)
	}
	if sz != 110 {
		t.Errorf("TagSize: got %d, want 110", sz)
	}
}

func TestValidate_withAppendedZIP(t *testing.T) {
	// Simulate polyglot: ID3v2 header + fake audio + appended ZIP bytes.
	mp3 := minID3(0xff, 0xfb, 0x90, 0x00)
	zipSuffix := []byte("PK\x05\x06some zip data")
	data := append(mp3, zipSuffix...)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid with appended ZIP: %v", err)
	}
}

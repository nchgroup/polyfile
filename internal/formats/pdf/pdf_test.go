package pdf

import (
	"bytes"
	"testing"
)

var h = handler{}

func TestValidate_valid(t *testing.T) {
	data := []byte("%PDF-1.4\n%some content\n%%EOF\n")
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid PDF, got: %v", err)
	}
}

func TestValidate_badMagic(t *testing.T) {
	data := []byte("NOTPDF\n%%EOF\n")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestValidate_noEOF(t *testing.T) {
	data := []byte("%PDF-1.4\nno trailer here\n")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for missing EOF marker")
	}
}

func TestValidate_tooSmall(t *testing.T) {
	data := []byte("%PDF")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for too-small file")
	}
}

package sh

import (
	"bytes"
	"testing"
)

var h = handler{}

func TestValidate_valid(t *testing.T) {
	data := []byte("#!/bin/sh\necho hello\nexit 0\n")
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid shell script: %v", err)
	}
}

func TestValidate_missingShebang(t *testing.T) {
	data := []byte("echo hello\n")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for missing shebang")
	}
}

func TestValidate_tooSmall(t *testing.T) {
	data := []byte("#")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for too-small file")
	}
}

func TestValidate_gifDataAfterScript(t *testing.T) {
	// Simulates polyglot: shell + GIF binary appended.
	data := append([]byte("#!/bin/sh\necho ok\nexit 0\n"), []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x3b}...)
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid with appended GIF data: %v", err)
	}
}

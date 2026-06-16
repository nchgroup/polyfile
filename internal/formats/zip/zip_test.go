package zip

import (
	"archive/zip"
	"bytes"
	"testing"
)

var h = handler{}

func makeZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(content))
	}
	w.Close()
	return buf.Bytes()
}

func TestValidate_valid(t *testing.T) {
	data := makeZIP(t, map[string]string{"hello.txt": "world"})
	r := bytes.NewReader(data)
	if err := h.Validate(r, int64(len(data))); err != nil {
		t.Fatalf("expected valid ZIP: %v", err)
	}
}

func TestValidate_notZIP(t *testing.T) {
	data := []byte("not a zip file at all")
	r := bytes.NewReader(data)
	if h.Validate(r, int64(len(data))) == nil {
		t.Fatal("expected error for non-ZIP data")
	}
}

func TestFindEOCD_fields(t *testing.T) {
	data := makeZIP(t, map[string]string{"a.txt": "aaa", "b.txt": "bbb"})
	r := bytes.NewReader(data)
	e, err := FindEOCD(r, int64(len(data)))
	if err != nil {
		t.Fatalf("FindEOCD: %v", err)
	}
	if e.TotalEntries != 2 {
		t.Errorf("TotalEntries: got %d, want 2", e.TotalEntries)
	}
	if e.CentralDirOffset == 0 {
		t.Error("CentralDirOffset should not be zero")
	}
}

func TestWritePatchedZIP_roundtrip(t *testing.T) {
	original := makeZIP(t, map[string]string{"foo.txt": "hello polyfile"})
	delta := uint32(1024)

	var patched bytes.Buffer
	r := bytes.NewReader(original)
	if err := writePatchedZIP(&patched, r, int64(len(original)), delta); err != nil {
		t.Fatalf("writePatchedZIP: %v", err)
	}

	// EOCD central dir offset must be shifted by delta.
	pr := bytes.NewReader(patched.Bytes())
	e, err := FindEOCD(pr, int64(patched.Len()))
	if err != nil {
		t.Fatalf("FindEOCD on patched: %v", err)
	}

	orig, _ := FindEOCD(r, int64(len(original)))
	if e.CentralDirOffset != orig.CentralDirOffset+delta {
		t.Errorf("CentralDirOffset: got %d, want %d", e.CentralDirOffset, orig.CentralDirOffset+delta)
	}
}

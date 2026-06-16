package jpg

import (
	"bytes"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var magic = []byte{0xff, 0xd8, 0xff} // SOI + first marker byte

type handler struct{}

func (handler) ID() string           { return "jpg" }
func (handler) Extensions() []string { return []string{".jpg", ".jpeg"} }

func (handler) Constraints() core.Constraints {
	return core.Constraints{
		MagicOffset:              0,
		Magic:                    magic,
		RequiresTrailer:          true,
		CanPrependArbitraryBytes: false,
		CanAppendArbitraryBytes:  true,
	}
}

func (handler) Validate(r io.ReaderAt, size int64) error {
	if size < int64(len(magic))+2 {
		return fmt.Errorf("file too small")
	}
	buf := make([]byte, len(magic))
	if _, err := r.ReadAt(buf, 0); err != nil {
		return fmt.Errorf("reading magic: %w", err)
	}
	if !bytes.Equal(buf, magic) {
		return fmt.Errorf("JPEG magic mismatch: got %X", buf)
	}
	// Scan forward for EOI (\xff\xd9). Real JPEG parsers stop at EOI and
	// ignore any appended bytes, so EOI need not be at absolute EOF.
	return findEOI(r, size)
}

// findEOI scans r for the EOI marker \xff\xd9.
func findEOI(r io.ReaderAt, size int64) error {
	const chunk = 4096
	buf := make([]byte, chunk)
	var prev byte
	for off := int64(0); off < size; {
		n := min(int64(chunk), size-off)
		nr, err := r.ReadAt(buf[:n], off)
		for i := 0; i < nr; i++ {
			if prev == 0xff && buf[i] == 0xd9 {
				return nil
			}
			prev = buf[i]
		}
		if err != nil {
			break
		}
		off += int64(nr)
	}
	return fmt.Errorf("JPEG EOI marker (\\xff\\xd9) not found")
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
}

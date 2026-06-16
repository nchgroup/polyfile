package pdf

import (
	"bytes"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var magic = []byte("%PDF-")

type handler struct{}

func (handler) ID() string         { return "pdf" }
func (handler) Extensions() []string { return []string{".pdf"} }

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
	if size < int64(len(magic)) {
		return fmt.Errorf("file too small")
	}
	buf := make([]byte, len(magic))
	if _, err := r.ReadAt(buf, 0); err != nil {
		return fmt.Errorf("reading magic: %w", err)
	}
	if !bytes.Equal(buf, magic) {
		return fmt.Errorf("PDF magic mismatch: got %q", buf)
	}
	// Scan last 1024 bytes for %%EOF
	scan := min(int64(1024), size)
	tail := make([]byte, scan)
	if _, err := r.ReadAt(tail, size-scan); err != nil {
		return fmt.Errorf("reading tail: %w", err)
	}
	if !bytes.Contains(tail, []byte("%%EOF")) {
		return fmt.Errorf("%%EOF marker not found in last %d bytes", scan)
	}
	return nil
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
}

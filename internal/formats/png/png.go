package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var magic = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a} // \x89PNG\r\n\x1a\n

type handler struct{}

func (handler) ID() string           { return "png" }
func (handler) Extensions() []string { return []string{".png"} }

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
		return fmt.Errorf("PNG magic mismatch")
	}
	// Walk chunks to find IEND; stop on first match (tolerates appended bytes).
	return walkChunks(r, size)
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
}

// walkChunks reads PNG chunks starting at offset 8 until IEND is found.
// It tolerates arbitrary bytes appended after IEND (for polyglots).
func walkChunks(r io.ReaderAt, size int64) error {
	const hdrSize = 8 // 4-byte length + 4-byte type
	off := int64(len(magic))

	for off+hdrSize <= size {
		var hdr [hdrSize]byte
		if _, err := r.ReadAt(hdr[:], off); err != nil {
			return fmt.Errorf("reading chunk header at offset %d: %w", off, err)
		}
		dataLen := int64(binary.BigEndian.Uint32(hdr[:4]))
		chunkType := string(hdr[4:8])

		if chunkType == "IEND" {
			return nil
		}
		off += hdrSize + dataLen + 4 // header + data + CRC
	}
	return fmt.Errorf("PNG IEND chunk not found")
}

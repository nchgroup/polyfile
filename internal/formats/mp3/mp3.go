package mp3

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var id3Magic = []byte("ID3")

type handler struct{}

func (handler) ID() string           { return "mp3" }
func (handler) Extensions() []string { return []string{".mp3"} }

func (handler) Constraints() core.Constraints {
	return core.Constraints{
		MagicOffset:              0,
		Magic:                    id3Magic,
		RequiresTrailer:          false,
		CanPrependArbitraryBytes: false,
		CanAppendArbitraryBytes:  true, // players stop at last audio frame
	}
}

func (handler) Validate(r io.ReaderAt, size int64) error {
	if size < 10 {
		return fmt.Errorf("file too small")
	}
	var hdr [10]byte
	if _, err := r.ReadAt(hdr[:], 0); err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	// ID3v2 tag: "ID3" + version + flags + syncsafe size
	if bytes.Equal(hdr[:3], id3Magic) {
		return validateID3v2Header(hdr)
	}

	// Bare MP3: sync word 0xFF 0xE? or 0xFF 0xF? (11 sync bits set)
	if hdr[0] == 0xff && hdr[1]&0xe0 == 0xe0 {
		return nil
	}

	return fmt.Errorf("not an MP3: missing ID3v2 tag or MPEG sync word at offset 0")
}

// validateID3v2Header checks version and syncsafe size bytes per the ID3v2 spec.
func validateID3v2Header(hdr [10]byte) error {
	// hdr[3] = major version, hdr[4] = revision; both must not be 0xFF.
	if hdr[3] == 0xff || hdr[4] == 0xff {
		return fmt.Errorf("invalid ID3v2 version bytes: %02x %02x", hdr[3], hdr[4])
	}
	// hdr[6:10] is the syncsafe tag size; high bit of each byte must be 0.
	for i := 6; i < 10; i++ {
		if hdr[i]&0x80 != 0 {
			return fmt.Errorf("ID3v2 size byte %d has high bit set (not syncsafe)", i-6)
		}
	}
	return nil
}

// DecodeSyncsafe decodes a 4-byte syncsafe integer (ID3v2 size field).
func DecodeSyncsafe(b []byte) uint32 {
	_ = binary.BigEndian // keep import used elsewhere if added later
	return uint32(b[0])<<21 | uint32(b[1])<<14 | uint32(b[2])<<7 | uint32(b[3])
}

// TagSize returns the total byte length of the ID3v2 tag (header + content).
// r must start with a valid ID3v2 header. Returns an error for non-ID3v2 files.
func TagSize(r io.ReaderAt) (int64, error) {
	var hdr [10]byte
	if _, err := r.ReadAt(hdr[:], 0); err != nil {
		return 0, err
	}
	if !bytes.Equal(hdr[:3], id3Magic) {
		return 0, fmt.Errorf("no ID3v2 tag")
	}
	tagContent := int64(DecodeSyncsafe(hdr[6:10]))
	const headerLen = 10
	// If the extended header flag (bit 6 of flags) is set, there's also a footer.
	hasFooter := hdr[5]&0x10 != 0
	if hasFooter {
		return headerLen + tagContent + 10, nil
	}
	return headerLen + tagContent, nil
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
}

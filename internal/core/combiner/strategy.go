package combiner

import (
	"io"

	"polyfile/internal/core"
)

// SuffixAppend appends format B after format A.
// Used by PDF+ZIP, PNG+ZIP, JPEG+ZIP.
// Format A must tolerate trailing bytes; Format B must search for trailer from EOF.
type SuffixAppend struct{}

func (SuffixAppend) Name() string { return "suffix-append" }

func (SuffixAppend) Apply(dst io.Writer, payloads []io.ReaderAt, sizes []int64, cfg core.Config) error {
	// Copy payload A verbatim.
	if err := copyPayload(dst, payloads[0], sizes[0]); err != nil {
		return err
	}
	// Copy payload B verbatim.
	// Offset patching (e.g. ZIP central directory offsets) is handled by the
	// format-specific strategy wrapper that embeds SuffixAppend.
	return copyPayload(dst, payloads[1], sizes[1])
}

// PrefixPrepend prepends format B header before format A payload.
// Used by GIF+SH.
type PrefixPrepend struct{}

func (PrefixPrepend) Name() string { return "prefix-prepend" }

func (PrefixPrepend) Apply(dst io.Writer, payloads []io.ReaderAt, sizes []int64, cfg core.Config) error {
	// Format B first (header/comment), then format A payload.
	if err := copyPayload(dst, payloads[1], sizes[1]); err != nil {
		return err
	}
	return copyPayload(dst, payloads[0], sizes[0])
}

// copyPayload streams all bytes from r into dst.
func copyPayload(dst io.Writer, r io.ReaderAt, size int64) error {
	buf := make([]byte, 32*1024)
	var off int64
	for off < size {
		n := int64(len(buf))
		if size-off < n {
			n = size - off
		}
		nr, err := r.ReadAt(buf[:n], off)
		if nr > 0 {
			if _, ew := dst.Write(buf[:nr]); ew != nil {
				return ew
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		off += int64(nr)
	}
	return nil
}

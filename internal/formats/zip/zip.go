package zip

import (
	"encoding/binary"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

const (
	localSig  uint32 = 0x04034b50
	cdSig     uint32 = 0x02014b50
	eocdSig   uint32 = 0x06054b50
	eocdFixed        = 22
	maxComment       = 65535
)

// EOCD is the end-of-central-directory record.
type EOCD struct {
	DiskNumber       uint16
	StartDisk        uint16
	EntriesOnDisk    uint16
	TotalEntries     uint16
	CentralDirSize   uint32
	CentralDirOffset uint32
	CommentLen       uint16
	Comment          []byte
	FileOffset       int64 // byte offset of this record in the source file
}

type handler struct{}

func (handler) ID() string           { return "zip" }
func (handler) Extensions() []string { return []string{".zip", ".jar", ".apk"} }

func (handler) Constraints() core.Constraints {
	return core.Constraints{
		MagicOffset:              -1,
		Magic:                    []byte{0x50, 0x4b, 0x05, 0x06},
		RequiresTrailer:          true,
		CanPrependArbitraryBytes: true,
		CanAppendArbitraryBytes:  false,
	}
}

func (handler) Validate(r io.ReaderAt, size int64) error {
	_, err := FindEOCD(r, size)
	return err
}

// FindEOCD locates and parses the EOCD record in r.
func FindEOCD(r io.ReaderAt, size int64) (*EOCD, error) {
	if size < eocdFixed {
		return nil, fmt.Errorf("file too small to contain EOCD")
	}
	searchLen := min(int64(eocdFixed+maxComment), size)
	buf := make([]byte, searchLen)
	startOff := size - searchLen
	if _, err := r.ReadAt(buf, startOff); err != nil {
		return nil, fmt.Errorf("reading EOCD search area: %w", err)
	}
	// Scan right-to-left: first valid match is the real EOCD.
	for i := int(searchLen) - eocdFixed; i >= 0; i-- {
		if binary.LittleEndian.Uint32(buf[i:]) != eocdSig {
			continue
		}
		commentLen := int(binary.LittleEndian.Uint16(buf[i+20:]))
		// EOCD + comment must end exactly at EOF.
		if i+eocdFixed+commentLen != int(searchLen) {
			continue
		}
		return &EOCD{
			DiskNumber:       binary.LittleEndian.Uint16(buf[i+4:]),
			StartDisk:        binary.LittleEndian.Uint16(buf[i+6:]),
			EntriesOnDisk:    binary.LittleEndian.Uint16(buf[i+8:]),
			TotalEntries:     binary.LittleEndian.Uint16(buf[i+10:]),
			CentralDirSize:   binary.LittleEndian.Uint32(buf[i+12:]),
			CentralDirOffset: binary.LittleEndian.Uint32(buf[i+16:]),
			CommentLen:       uint16(commentLen),
			Comment:          append([]byte(nil), buf[i+eocdFixed:i+eocdFixed+commentLen]...),
			FileOffset:       startOff + int64(i),
		}, nil
	}
	return nil, fmt.Errorf("EOCD not found")
}

// --- Suffix-append strategy (reused by all X+ZIP pairs) ---

// suffixZIPStrategy writes the primary format verbatim then appends a ZIP
// with its central-directory offsets patched by +primarySize.
// It works for any pair where the primary format tolerates trailing bytes.
type suffixZIPStrategy struct{ name string }

func (s suffixZIPStrategy) Name() string { return s.name }

func (s suffixZIPStrategy) Apply(dst io.Writer, payloads []io.ReaderAt, sizes []int64, cfg core.Config) error {
	// Find which input is ZIP and which is primary.
	primaryIdx, zipIdx := 0, 1
	if cfg.Inputs[0].Handler.ID() == "zip" {
		primaryIdx, zipIdx = 1, 0
	}

	primarySize := sizes[primaryIdx]
	zipPayload := payloads[zipIdx]
	zipSize := sizes[zipIdx]

	if err := copyRange(dst, payloads[primaryIdx], 0, primarySize); err != nil {
		return fmt.Errorf("writing primary payload: %w", err)
	}
	return writePatchedZIP(dst, zipPayload, zipSize, uint32(primarySize))
}

func newSuffixZIPFactory(pairName string) registry.StrategyFactory {
	return func(a, b core.Handler) (core.Strategy, error) {
		return suffixZIPStrategy{name: pairName}, nil
	}
}

// writePatchedZIP writes the ZIP payload to dst, patching central-directory
// offsets by +delta so they remain valid when appended after delta bytes.
func writePatchedZIP(dst io.Writer, r io.ReaderAt, size int64, delta uint32) error {
	eocd, err := FindEOCD(r, size)
	if err != nil {
		return fmt.Errorf("ZIP: %w", err)
	}

	cdStart := int64(eocd.CentralDirOffset)
	cdEnd := cdStart + int64(eocd.CentralDirSize)

	// Local file headers + data: copy verbatim (no absolute offsets inside).
	if err := copyRange(dst, r, 0, cdStart); err != nil {
		return fmt.Errorf("writing local entries: %w", err)
	}

	// Central directory: patch each entry's local-header offset field.
	if err := writePatchedCD(dst, r, cdStart, cdEnd, delta); err != nil {
		return fmt.Errorf("writing central directory: %w", err)
	}

	// EOCD with patched central-directory offset.
	return writePatchedEOCD(dst, eocd, delta)
}

// writePatchedCD iterates central-directory entries from [cdStart, cdEnd)
// and rewrites each local-header offset += delta.
func writePatchedCD(dst io.Writer, r io.ReaderAt, cdStart, cdEnd int64, delta uint32) error {
	const hdrSize = 46
	off := cdStart
	for off < cdEnd {
		var hdr [hdrSize]byte
		if _, err := r.ReadAt(hdr[:], off); err != nil {
			return fmt.Errorf("reading CD entry at %d: %w", off, err)
		}
		if binary.LittleEndian.Uint32(hdr[:4]) != cdSig {
			return fmt.Errorf("invalid CD signature at offset %d", off)
		}
		fileNameLen := int(binary.LittleEndian.Uint16(hdr[28:]))
		extraLen := int(binary.LittleEndian.Uint16(hdr[30:]))
		commentLen := int(binary.LittleEndian.Uint16(hdr[32:]))

		// Patch local-header offset (bytes 42–45).
		orig := binary.LittleEndian.Uint32(hdr[42:])
		binary.LittleEndian.PutUint32(hdr[42:], orig+delta)

		if _, err := dst.Write(hdr[:]); err != nil {
			return err
		}

		varLen := fileNameLen + extraLen + commentLen
		if varLen > 0 {
			varBuf := make([]byte, varLen)
			if _, err := r.ReadAt(varBuf, off+hdrSize); err != nil {
				return fmt.Errorf("reading CD variable fields at %d: %w", off, err)
			}
			if _, err := dst.Write(varBuf); err != nil {
				return err
			}
		}
		off += hdrSize + int64(varLen)
	}
	return nil
}

func writePatchedEOCD(dst io.Writer, e *EOCD, delta uint32) error {
	var buf [eocdFixed]byte
	binary.LittleEndian.PutUint32(buf[0:], eocdSig)
	binary.LittleEndian.PutUint16(buf[4:], e.DiskNumber)
	binary.LittleEndian.PutUint16(buf[6:], e.StartDisk)
	binary.LittleEndian.PutUint16(buf[8:], e.EntriesOnDisk)
	binary.LittleEndian.PutUint16(buf[10:], e.TotalEntries)
	binary.LittleEndian.PutUint32(buf[12:], e.CentralDirSize)
	binary.LittleEndian.PutUint32(buf[16:], e.CentralDirOffset+delta)
	binary.LittleEndian.PutUint16(buf[20:], e.CommentLen)
	if _, err := dst.Write(buf[:]); err != nil {
		return err
	}
	if len(e.Comment) > 0 {
		_, err := dst.Write(e.Comment)
		return err
	}
	return nil
}

func copyRange(dst io.Writer, r io.ReaderAt, start, end int64) error {
	buf := make([]byte, 32*1024)
	for off := start; off < end; {
		n := min(int64(len(buf)), end-off)
		nr, err := r.ReadAt(buf[:n], off)
		if nr > 0 {
			if _, ew := dst.Write(buf[:nr]); ew != nil {
				return ew
			}
		}
		if err != nil && err != io.EOF {
			return err
		}
		off += int64(nr)
	}
	return nil
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
	registry.RegisterPairWithMeta("pdf", "zip", "suffix-append",
		"ZIP EOCD appended after %%EOF; CD offsets patched",
		newSuffixZIPFactory("pdf+zip/suffix-append"))
	registry.RegisterPairWithMeta("png", "zip", "suffix-append",
		"ZIP EOCD appended after IEND chunk; CD offsets patched",
		newSuffixZIPFactory("png+zip/suffix-append"))
	registry.RegisterPairWithMeta("jpg", "zip", "suffix-append",
		"ZIP EOCD appended after JPEG EOI; CD offsets patched",
		newSuffixZIPFactory("jpg+zip/suffix-append"))
	registry.RegisterPairWithMeta("mp3", "zip", "suffix-append",
		"ZIP EOCD appended after last audio frame; CD offsets patched",
		newSuffixZIPFactory("mp3+zip/suffix-append"))
}

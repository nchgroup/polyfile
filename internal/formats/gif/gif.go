package gif

import (
	"bytes"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var (
	magic87 = []byte("GIF87a")
	magic89 = []byte("GIF89a")
	trailer = byte(0x3b)
)

// scanLimit is how far into the file to search for GIF magic.
// Shell script preamble before the GIF may be up to this size.
const scanLimit = 4096

type handler struct{}

func (handler) ID() string           { return "gif" }
func (handler) Extensions() []string { return []string{".gif"} }

func (handler) Constraints() core.Constraints {
	return core.Constraints{
		MagicOffset:              -1, // scanned, not at fixed offset
		Magic:                    magic89,
		RequiresTrailer:          true,
		CanPrependArbitraryBytes: true, // shell script precedes GIF in polyglot
		CanAppendArbitraryBytes:  false,
	}
}

func (handler) Validate(r io.ReaderAt, size int64) error {
	if size < 7 {
		return fmt.Errorf("file too small")
	}
	if _, err := findMagic(r, size); err != nil {
		return err
	}
	// Trailer must be last byte.
	last := make([]byte, 1)
	if _, err := r.ReadAt(last, size-1); err != nil {
		return fmt.Errorf("reading trailer: %w", err)
	}
	if last[0] != trailer {
		return fmt.Errorf("GIF trailer (0x3b) not found at end of file")
	}
	return nil
}

// findMagic scans the first scanLimit bytes for GIF87a or GIF89a.
func findMagic(r io.ReaderAt, size int64) (int64, error) {
	scan := min(int64(scanLimit), size)
	buf := make([]byte, scan)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return 0, fmt.Errorf("reading scan area: %w", err)
	}
	for i := int64(0); i <= scan-6; i++ {
		if bytes.Equal(buf[i:i+6], magic87) || bytes.Equal(buf[i:i+6], magic89) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("GIF magic (GIF87a/GIF89a) not found in first %d bytes", scan)
}

// --- GIF+SH strategy ---

type gifSHStrategy struct{}

func (gifSHStrategy) Name() string { return "gif+sh/prefix-prepend" }

func (gifSHStrategy) Apply(dst io.Writer, payloads []io.ReaderAt, sizes []int64, cfg core.Config) error {
	shIdx, gifIdx := 0, 1
	if cfg.Inputs[0].Handler.ID() == "gif" {
		shIdx, gifIdx = 1, 0
	}

	// Write shell script.
	if err := copyRange(dst, payloads[shIdx], 0, sizes[shIdx]); err != nil {
		return fmt.Errorf("writing shell script: %w", err)
	}

	// Inject exit guard so the shell does not attempt to parse GIF binary.
	// If the user's script already calls exit, this line is unreachable (harmless).
	if _, err := fmt.Fprint(dst, "\nexit 0\n"); err != nil {
		return fmt.Errorf("writing exit guard: %w", err)
	}

	// Write GIF verbatim.
	if err := copyRange(dst, payloads[gifIdx], 0, sizes[gifIdx]); err != nil {
		return fmt.Errorf("writing GIF: %w", err)
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
	registry.RegisterPairWithMeta("gif", "sh", "prefix-prepend",
		"shell shebang + script prepended; exit guard prevents shell parsing GIF binary",
		func(a, b core.Handler) (core.Strategy, error) {
			return gifSHStrategy{}, nil
		})
}

package sh

import (
	"bytes"
	"fmt"
	"io"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

var shebang = []byte("#!")

type handler struct{}

func (handler) ID() string           { return "sh" }
func (handler) Extensions() []string { return []string{".sh", ".bash"} }

func (handler) Constraints() core.Constraints {
	return core.Constraints{
		MagicOffset:              0,
		Magic:                    shebang,
		RequiresTrailer:          false,
		CanPrependArbitraryBytes: false,
		CanAppendArbitraryBytes:  true,
	}
}

func (handler) Validate(r io.ReaderAt, size int64) error {
	if size < 2 {
		return fmt.Errorf("file too small")
	}
	buf := make([]byte, 2)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return fmt.Errorf("reading shebang: %w", err)
	}
	if !bytes.Equal(buf, shebang) {
		return fmt.Errorf("missing shebang (#!) at offset 0: got %q", buf)
	}
	return nil
}

func init() {
	registry.Register(handler{}, registry.SourceBuiltin, "")
}

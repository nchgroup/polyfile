package writer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Result holds metadata about the written output file.
type Result struct {
	Path   string
	Size   int64
	SHA256 string
}

// Write atomically writes content produced by fn to path.
// It writes to a temp file first, then renames on success.
func Write(path string, fn func(w io.Writer) error) (*Result, error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".polyfile-tmp-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpPath)
		}
	}()

	h := sha256.New()
	mw := io.MultiWriter(tmp, h)

	if err := fn(mw); err != nil {
		return nil, err
	}

	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("closing temp file: %w", err)
	}

	info, err := os.Stat(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("stat temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, fmt.Errorf("renaming to %s: %w", path, err)
	}

	success = true
	return &Result{
		Path:   path,
		Size:   info.Size(),
		SHA256: hex.EncodeToString(h.Sum(nil)),
	}, nil
}

package combiner

import (
	"fmt"
	"io"
	"os"

	"polyfile/internal/core"
	"polyfile/internal/core/writer"
)

// Combine orchestrates reading, strategy application, and writing.
func Combine(cfg core.Config) (*writer.Result, error) {
	if len(cfg.Inputs) < 2 {
		return nil, fmt.Errorf("at least two input files required")
	}

	files := make([]*os.File, len(cfg.Inputs))
	sizes := make([]int64, len(cfg.Inputs))
	payloads := make([]io.ReaderAt, len(cfg.Inputs))

	for i, inp := range cfg.Inputs {
		f, err := os.Open(inp.Path)
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", inp.Path, err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", inp.Path, err)
		}

		files[i] = f
		sizes[i] = info.Size()
		payloads[i] = f
	}

	result, err := writer.Write(cfg.Output, func(w io.Writer) error {
		return cfg.Strategy.Apply(w, payloads, sizes, cfg)
	})
	if err != nil {
		return nil, fmt.Errorf("writing output: %w", err)
	}

	if cfg.Verify {
		if err := verify(cfg); err != nil {
			return nil, fmt.Errorf("output verification failed: %w", err)
		}
	}

	return result, nil
}

func verify(cfg core.Config) error {
	f, err := os.Open(cfg.Output)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	for _, h := range cfg.Pair {
		if err := h.Validate(f, info.Size()); err != nil {
			return fmt.Errorf("format %s validation: %w", h.ID(), err)
		}
	}
	return nil
}

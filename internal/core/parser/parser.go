package parser

import (
	"fmt"
	"path/filepath"

	"polyfile/internal/core"
	"polyfile/internal/core/registry"
)

// Parse resolves the Config from a list of input paths and an output path.
// It looks up handlers by file extension and validates compatibility.
func Parse(inputs []string, output string, verify bool) (core.Config, error) {
	if len(inputs) != 2 {
		return core.Config{}, fmt.Errorf("exactly 2 input files required, got %d", len(inputs))
	}

	var handlers [2]core.Handler
	var inFiles [2]core.InputFile

	for i, path := range inputs {
		ext := filepath.Ext(path)
		h, err := registry.LookupByExt(ext)
		if err != nil {
			return core.Config{}, fmt.Errorf("input %s: %w", path, err)
		}
		handlers[i] = h
		inFiles[i] = core.InputFile{
			Path:    path,
			Handler: h,
		}
	}

	if err := registry.ValidateCompat(handlers[0], handlers[1]); err != nil {
		return core.Config{}, fmt.Errorf("incompatible formats: %w", err)
	}

	strat, err := registry.ResolveStrategy(handlers[0], handlers[1])
	if err != nil {
		return core.Config{}, err
	}

	cfg := core.Config{
		Inputs:   []core.InputFile{inFiles[0], inFiles[1]},
		Output:   output,
		Pair:     handlers,
		Strategy: strat,
		Verify:   verify,
	}
	return cfg, nil
}

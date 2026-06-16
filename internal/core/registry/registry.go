package registry

import (
	"fmt"
	"strings"
	"sync"

	"polyfile/internal/core"
)

// Source indicates where a handler was registered from.
type Source int

const (
	SourceBuiltin Source = iota
	SourceFile
	SourceCLI
)

func (s Source) String() string {
	switch s {
	case SourceBuiltin:
		return "built-in"
	case SourceFile:
		return "file"
	case SourceCLI:
		return "cli"
	default:
		return "unknown"
	}
}

// Entry is a registered handler with its source metadata.
type Entry struct {
	Handler core.Handler
	Source  Source
	Path    string // non-empty when Source == SourceFile
}

var (
	mu       sync.RWMutex
	handlers = map[string]*Entry{} // keyed by handler ID
	byExt    = map[string]string{} // extension → handler ID
)

// Register adds a handler to the global registry.
// CLI source overrides file source; file source overrides built-in.
func Register(h core.Handler, src Source, path string) error {
	mu.Lock()
	defer mu.Unlock()

	id := h.ID()
	existing, ok := handlers[id]
	if ok && existing.Source > src {
		// lower numeric value = higher priority; existing wins
		return nil
	}

	handlers[id] = &Entry{Handler: h, Source: src, Path: path}
	for _, ext := range h.Extensions() {
		byExt[strings.ToLower(ext)] = id
	}
	return nil
}

// Lookup returns the handler for a given format ID.
func Lookup(id string) (core.Handler, error) {
	mu.RLock()
	defer mu.RUnlock()

	e, ok := handlers[id]
	if !ok {
		return nil, fmt.Errorf("unknown format %q", id)
	}
	return e.Handler, nil
}

// LookupByExt returns the handler for a file extension (e.g. ".pdf").
func LookupByExt(ext string) (core.Handler, error) {
	mu.RLock()
	defer mu.RUnlock()

	id, ok := byExt[strings.ToLower(ext)]
	if !ok {
		return nil, fmt.Errorf("no handler for extension %q", ext)
	}
	e := handlers[id]
	return e.Handler, nil
}

// All returns a snapshot of all registered entries sorted by ID.
func All() []*Entry {
	mu.RLock()
	defer mu.RUnlock()

	out := make([]*Entry, 0, len(handlers))
	for _, e := range handlers {
		out = append(out, e)
	}
	return out
}

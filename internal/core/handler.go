package core

import "io"

// Handler knows how to read, validate, and embed a specific format.
type Handler interface {
	ID() string
	Extensions() []string
	Validate(r io.ReaderAt, size int64) error
	Constraints() Constraints
}

// Constraints describes where a format's required bytes must live.
type Constraints struct {
	MagicOffset              int64
	Magic                    []byte
	RequiresTrailer          bool
	CanPrependArbitraryBytes bool
	CanAppendArbitraryBytes  bool
	InternalOffsets          []OffsetField
}

// OffsetField is an absolute-offset field inside a format that must be patched on relocation.
type OffsetField struct {
	Name   string
	Offset int64
	Width  int // bytes
	Endian string // "le" | "be"
}

// Strategy performs the actual byte combination for a given format pair.
type Strategy interface {
	Name() string
	Apply(dst io.Writer, payloads []io.ReaderAt, sizes []int64, cfg Config) error
}

// InputFile is a resolved input with its open reader and detected handler.
type InputFile struct {
	Path    string
	Handler Handler
	Size    int64
	Reader  io.ReaderAt
}

// Config is the validated, parsed representation of the user's request.
type Config struct {
	Inputs   []InputFile
	Output   string
	Pair     [2]Handler
	Strategy Strategy
	Verify   bool
}

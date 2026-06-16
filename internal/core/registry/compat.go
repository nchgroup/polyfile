package registry

import (
	"fmt"
	"sort"
	"sync"

	"polyfile/internal/core"
)

// PairKey is a canonical, order-independent key for a format pair.
type PairKey struct {
	A, B string // sorted lexicographically so A <= B always
}

func NewPairKey(a, b string) PairKey {
	if a > b {
		a, b = b, a
	}
	return PairKey{A: a, B: b}
}

// PairMeta holds human-readable info about a registered format pair.
type PairMeta struct {
	A, B     string
	Strategy string
	Notes    string
}

// StrategyFactory creates a Strategy for a given pair.
type StrategyFactory func(a, b core.Handler) (core.Strategy, error)

var (
	pairMu    sync.RWMutex
	strategies = map[PairKey]StrategyFactory{}
	pairMetas  []PairMeta
)

// RegisterPair registers a strategy factory for a format pair.
func RegisterPair(a, b string, f StrategyFactory) {
	RegisterPairWithMeta(a, b, "", "", f)
}

// RegisterPairWithMeta registers a strategy factory with human-readable metadata.
func RegisterPairWithMeta(a, b, strategy, notes string, f StrategyFactory) {
	key := NewPairKey(a, b)
	pairMu.Lock()
	defer pairMu.Unlock()
	strategies[key] = f
	pairMetas = append(pairMetas, PairMeta{A: key.A, B: key.B, Strategy: strategy, Notes: notes})
}

// AllPairs returns a snapshot of registered pair metadata sorted by A+B.
func AllPairs() []PairMeta {
	pairMu.RLock()
	defer pairMu.RUnlock()
	out := append([]PairMeta(nil), pairMetas...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	return out
}

// ResolveStrategy returns the strategy for combining handlers a and b.
func ResolveStrategy(a, b core.Handler) (core.Strategy, error) {
	key := NewPairKey(a.ID(), b.ID())
	pairMu.RLock()
	f, ok := strategies[key]
	pairMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no strategy for pair %q + %q", a.ID(), b.ID())
	}
	return f(a, b)
}

// ValidateCompat checks whether two handlers can be combined.
func ValidateCompat(a, b core.Handler) error {
	ca := a.Constraints()
	cb := b.Constraints()

	if ca.MagicOffset == 0 && cb.MagicOffset == 0 &&
		!ca.CanPrependArbitraryBytes && !cb.CanPrependArbitraryBytes {
		return fmt.Errorf(
			"format %q and %q both require magic at offset 0 and neither tolerates prepended bytes",
			a.ID(), b.ID(),
		)
	}

	if ca.RequiresTrailer && cb.RequiresTrailer &&
		!ca.CanAppendArbitraryBytes && !cb.CanAppendArbitraryBytes {
		return fmt.Errorf(
			"format %q and %q both require trailers and neither tolerates appended bytes",
			a.ID(), b.ID(),
		)
	}

	_, err := ResolveStrategy(a, b)
	return err
}

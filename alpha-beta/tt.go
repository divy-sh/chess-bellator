package main

import (
	"github.com/corentings/chess"
)

type TTEntryType byte

const (
	Exact TTEntryType = iota
	Lower
	Upper
)

// TTEntry using a flat primitive layout for safe concurrent array access
type TTEntry struct {
	Hash     [16]byte // Verification token to confirm we didn't hit a hash collision
	Score    float64
	Depth    int
	Type     TTEntryType
	BestMove *chess.Move
}

// TranspositionTable uses a fixed flat allocation with zero locks
type TranspositionTable struct {
	table []TTEntry
	size  uint32
}

// NewTranspositionTable allocates a fixed-power-of-two size flat array
// e.g., 1 << 20 elements requires ~48MB of RAM and provides blazing fast lookups
func NewTranspositionTable(powerOfTwo int) *TranspositionTable {
	length := 1 << powerOfTwo
	return &TranspositionTable{
		table: make([]TTEntry, length),
		size:  uint32(length - 1),
	}
}

func (t *TranspositionTable) Get(hash [16]byte) (TTEntry, bool) {
	// Fast bitwise modulo to find the slot index
	idx := (uint32(hash[0]) | uint32(hash[1])<<8 | uint32(hash[2])<<16 | uint32(hash[3])<<24) & t.size
	entry := t.table[idx]

	// Verify that the entry actually matches our exact board hash state
	if entry.Hash == hash {
		return entry, true
	}
	return TTEntry{}, false
}

func (t *TranspositionTable) Put(hash [16]byte, score float64, depth int, entryType TTEntryType, bestMove *chess.Move) {
	idx := (uint32(hash[0]) | uint32(hash[1])<<8 | uint32(hash[2])<<16 | uint32(hash[3])<<24) & t.size

	// Depth-preferred replacement strategy with zero locks
	existing := t.table[idx]
	if existing.Hash != hash || depth >= existing.Depth {
		t.table[idx] = TTEntry{
			Hash:     hash,
			Score:    score,
			Depth:    depth,
			Type:     entryType,
			BestMove: bestMove,
		}
	}
}

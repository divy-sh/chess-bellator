package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/corentings/chess"
)

type GameWrapper struct {
	mu   sync.Mutex
	game *chess.Game
}

func NewGameWrapper() *GameWrapper {
	return &GameWrapper{game: chess.NewGame()}
}

// PlayMove attempts to play a move string (handles "g1f3" or fallback "Nf3")
func (g *GameWrapper) PlayMove(s string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	input := strings.TrimSpace(s)

	uci := chess.UCINotation{}
	for _, move := range g.game.Position().ValidMoves() {
		if strings.EqualFold(uci.Encode(g.game.Position(), move), input) {
			if err := g.game.Move(move); err == nil {
				return true
			}
		}
	}

	return false
}

// PlayBestMove gets the best move for whoever's turn it is and plays it
func (g *GameWrapper) PlayBestMove() (*chess.Move, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Time the search for performance logging
	startTime := time.Now()
	defer func() {
		fmt.Printf("[Search Complete] Depth: 7 Time: %s\n",
			time.Since(startTime))
	}()

	pos := g.game.Position()

	move, _, _ := GenMove(7, pos)
	if move == nil {
		return nil, false
	}

	if err := g.game.Move(move); err != nil {
		return nil, false
	}
	return move, true
}

func (g *GameWrapper) Position() *chess.Position {
	return g.game.Position()
}

func (g *GameWrapper) GameOver() bool {
	return g.game.Outcome() != chess.NoOutcome
}

func (g *GameWrapper) PGN() string {
	return g.game.String()
}

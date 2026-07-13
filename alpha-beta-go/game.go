package main

import (
	"strings"
	"sync"

	"github.com/corentings/chess"
)

type GameWrapper struct {
	mu   sync.RWMutex
	game *chess.Game
}

func NewGameWrapper() *GameWrapper {
	return &GameWrapper{game: chess.NewGame()}
}

func (g *GameWrapper) PlayMove(s string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	input := strings.TrimSpace(s)
	if input == "" {
		return false
	}
	if err := g.game.MoveStr(input); err == nil {
		return true
	}
	if move, ok := g.findMoveByUCI(input); ok {
		if err := g.game.Move(move); err != nil {
			return false
		}
		return true
	}
	return false
}

func (g *GameWrapper) Position() *chess.Position {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.game.Position()
}

func (g *GameWrapper) Outcome() chess.Outcome {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.game.Outcome()
}

func (g *GameWrapper) GameOver() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.game.Outcome() != chess.NoOutcome
}

func (g *GameWrapper) String() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.game.String()
}

func (g *GameWrapper) PGN() string {
	return g.String()
}

func (g *GameWrapper) SanMove(move *chess.Move) string {
	return move.String()
}

func (g *GameWrapper) PlayBestMove() bool {
	g.mu.RLock()
	pos := g.game.Position()
	g.mu.RUnlock()
	move, _ := GenMove(3, pos)
	if move == nil {
		return false
	}
	return g.PlayMove(move.String())
}

func (g *GameWrapper) LegalMovesForSquare(square string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	moves := g.game.Position().ValidMoves()
	var result []string
	for _, move := range moves {
		if squareName(move.S1()) == square {
			result = append(result, move.String())
		}
	}
	return result
}

func (g *GameWrapper) findMoveByUCI(uci string) (*chess.Move, bool) {
	moves := g.game.Position().ValidMoves()
	for _, move := range moves {
		if strings.EqualFold(move.String(), strings.TrimSpace(uci)) {
			return move, true
		}
	}
	return nil, false
}

func squareName(sq chess.Square) string {
	file := byte('a') + byte(int(sq)%8)
	rank := byte('1') + byte(int(sq)/8)
	return string([]byte{file, rank})
}

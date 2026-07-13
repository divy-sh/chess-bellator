package main

import (
	"math"
	"time"

	"github.com/corentings/chess"
)

const (
	mateScore     = 100000.0
	initialDepth  = 2
	maxSearchTime = 5.0
)

var endGame bool

func GenMoveIterative(seconds float64, pos *chess.Position) (*chess.Move, float64) {
	timeout := time.Now().Add(time.Duration(seconds * float64(time.Second))).UnixNano()
	depth := initialDepth
	var move *chess.Move
	var value float64
	for time.Now().UnixNano() < timeout {
		depth++
		move, value = GenMove(depth, pos)
	}
	return move, value
}

func GenMove(depth int, pos *chess.Position) (*chess.Move, float64) {
	moves := getOrderedMoves(pos)
	if len(moves) == 0 {
		return nil, 0
	}

	bestValue := -math.MaxFloat64
	var bestMove *chess.Move
	for _, mv := range moves {
		child := pos.Update(mv)
		value := -negamax(depth-1, -math.MaxFloat64, math.MaxFloat64, child)
		if value > bestValue {
			bestValue = value
			bestMove = mv
		}
	}
	return bestMove, bestValue
}

func negamax(depth int, alpha, beta float64, pos *chess.Position) float64 {
	if pos == nil {
		return 0
	}
	if isCheckmate(pos) {
		return -mateScore * float64(depth+1)
	}
	if canClaimDraw(pos) {
		return 0
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return 0
	}
	if depth == 0 {
		return quiescenceSearch(alpha, beta, pos)
	}

	for _, mv := range moves {
		child := pos.Update(mv)
		value := -negamax(depth-1, -beta, -alpha, child)
		if value >= beta {
			return beta
		}
		if value > alpha {
			alpha = value
		}
	}
	return alpha
}

func quiescenceSearch(alpha, beta float64, pos *chess.Position) float64 {
	value := evaluate(pos)
	if value >= beta {
		return beta
	}
	if value > alpha {
		alpha = value
	}

	for _, mv := range pos.ValidMoves() {
		if !isCapture(mv, pos) {
			continue
		}
		child := pos.Update(mv)
		value = -quiescenceSearch(-beta, -alpha, child)
		if value >= beta {
			return beta
		}
		if value > alpha {
			alpha = value
		}
	}
	return alpha
}

func getOrderedMoves(pos *chess.Position) []*chess.Move {
	moves := pos.ValidMoves()
	ordered := make([]*chess.Move, len(moves))
	copy(ordered, moves)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if evaluateMove(ordered[j], pos) > evaluateMove(ordered[i], pos) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	return ordered
}

package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/notnil/chess"
)

type Engine struct {
	game *chess.Game
}

func newEngine(game *chess.Game) *Engine {
	return &Engine{
		game: game,
	}
}

func (e *Engine) genMoveIterative(seconds int) *chess.Move {
	defer timer("genMoveIterative")()
	done := make(chan bool)
	go func() {
		time.Sleep(time.Duration(seconds) * time.Second)
		done <- true
	}()
	depth := 2
	var move *chess.Move = nil
	for {
		select {
		case <-done:
			fmt.Printf("depth reached - %d\n", depth)
			return move
		default:
			depth++
			move = e.genMove(depth)
		}
	}
}

func (e *Engine) genMove(depth int) *chess.Move {
	defer timer("genMove")()
	move, _ := e.alphaBeta(depth, *e.game.Position(), math.Inf(-1), math.Inf(1), e.game.Position().Turn() == chess.White)
	return move
}

func (e *Engine) alphaBeta(depth int, pos chess.Position, alpha float64, beta float64, isMax bool) (*chess.Move, float64) {
	if depth == 0 {
		return nil, e.eval(pos, isMax)
	}
	bestValue := 0.0
	if isMax {
		bestValue = math.Inf(-1)
	} else {
		bestValue = math.Inf(1)
	}
	var bestMove *chess.Move
	moves := pos.ValidMoves()
	sort.Slice(moves, func(i int, j int) bool {
		iPriority := 0.0
		jPriority := 0.0
		// check if promotion move
		if moves[i].Promo() != chess.NoPiece.Type() {
			iPriority += pieceVal[moves[i].Promo()] / pieceVal[chess.Queen]
		}
		if moves[j].Promo() != chess.NoPiece.Type() {
			jPriority += pieceVal[moves[j].Promo()] / pieceVal[chess.Queen]
		}
		// check if capture move
		if pos.Board().Piece(moves[i].S2()) != chess.NoPiece {
			iPriority += pieceVal[pos.Board().Piece(moves[i].S2()).Type()] / pieceVal[chess.Queen]
		}
		if pos.Board().Piece(moves[j].S2()) != chess.NoPiece {
			iPriority += pieceVal[pos.Board().Piece(moves[j].S2()).Type()] / pieceVal[chess.Queen]
		}
		//check if check
		if moves[i].HasTag(chess.Check) {
			iPriority += 0.5
		}
		if moves[j].HasTag(chess.Check) {
			jPriority += 0.5
		}
		return iPriority > jPriority
	})
	for _, move := range moves {
		prevPos := pos
		pos = *pos.Update(move)
		_, value := e.alphaBeta(depth-1, pos, alpha, beta, !isMax)
		pos = prevPos
		if isMax {
			if bestValue <= value {
				bestValue = value
				bestMove = move
			}
			alpha = math.Max(alpha, bestValue)
		} else {
			if bestValue >= value {
				bestValue = value
				bestMove = move
			}
			beta = math.Min(beta, bestValue)
		}
		if beta <= alpha {
			return bestMove, bestValue
		}
	}
	return bestMove, bestValue
}

func (e *Engine) eval(pos chess.Position, isMax bool) float64 {
	var value float64
	var materialVal float64
	var positionalVal float64
	var mobilityVal float64
	for square, piece := range pos.Board().SquareMap() {
		if piece.Color() == chess.White {
			materialVal += pieceVal[piece.Type()]
			positionalVal += e.getPieceSquareTable(pos, &square, &piece)
		} else {
			materialVal -= pieceVal[piece.Type()]
			positionalVal -= e.getPieceSquareTable(pos, &square, &piece)
		}
	}
	if isMax {
		mobilityVal -= float64(len(pos.ValidMoves()))
	} else {
		mobilityVal += float64(len(pos.ValidMoves()))
	}
	value = materialVal + positionalVal + mobilityVal

	return value
}

func (e *Engine) getPieceSquareTable(pos chess.Position, square *chess.Square, piece *chess.Piece) float64 {
	sq := int(*square)
	if piece.Color() == chess.White {
		if piece.Type() == chess.Pawn {
			return pawnTable[sq]
		}
		if piece.Type() == chess.Knight {
			return knightTable[sq]
		}
		if piece.Type() == chess.Bishop {
			return bishopTable[sq]
		}
		if piece.Type() == chess.Rook {
			return rookTable[sq]
		}
		if piece.Type() == chess.Queen {
			return queenTable[sq]
		}
		if piece.Type() == chess.King {
			gamePhase := e.getGamePhase(pos)
			return gamePhase*kingMiddleGameTable[sq] + (1-gamePhase)*kingEndGameTable[sq]
		}
	} else {
		if piece.Type() == chess.Pawn {
			return revPawnTable[sq]
		}
		if piece.Type() == chess.Knight {
			return revKnightTable[sq]
		}
		if piece.Type() == chess.Bishop {
			return revBishopTable[sq]
		}
		if piece.Type() == chess.Rook {
			return revRookTable[sq]
		}
		if piece.Type() == chess.Queen {
			return revQueenTable[sq]
		}
		if piece.Type() == chess.King {
			gamePhase := e.getGamePhase(pos)
			return gamePhase*revKingMiddleGameTable[sq] + (1-gamePhase)*revKingEndGameTable[sq]
		}
	}
	return 0
}

func (e *Engine) getGamePhase(pos chess.Position) float64 {
	totalPieces := len(pos.Board().SquareMap())
	maxPieces := 32
	gamePhase := totalPieces / maxPieces
	return float64(gamePhase)
}

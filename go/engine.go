package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/notnil/chess"
)

type MoveEvalPair struct {
	move *chess.Move
	eval float64
}

type Engine struct {
	game              *chess.Game
	tranposTable      map[string]MoveEvalPair
	transposTableHits int
}

func newEngine(game *chess.Game) *Engine {
	return &Engine{
		game:         game,
		tranposTable: map[string]MoveEvalPair{},
	}
}

func (e *Engine) genMove(depth int) *chess.Move {
	move, _ := e.alphaBeta(depth, *e.game.Position(), math.Inf(-1), math.Inf(1), e.game.Position().Turn() == chess.White)
	return move
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

func (e *Engine) alphaBeta(depth int, pos chess.Position, alpha float64, beta float64, isMax bool) (*chess.Move, float64) {
	transPosTableKey := pos.String() + fmt.Sprint(depth) + strconv.FormatFloat(alpha, 'f', -1, 64) + strconv.FormatFloat(beta, 'f', -1, 64)
	if val, ok := e.tranposTable[transPosTableKey]; ok {
		e.transposTableHits++
		return val.move, val.eval
	}
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
	for _, move := range pos.ValidMoves() {
		pos = *pos.Update(move)
		_, value := e.alphaBeta(depth-1, pos, alpha, beta, !isMax)
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
			e.tranposTable[transPosTableKey] = MoveEvalPair{move: bestMove, eval: bestValue}
			return bestMove, bestValue
		}
	}
	e.tranposTable[transPosTableKey] = MoveEvalPair{move: bestMove, eval: bestValue}
	return bestMove, bestValue
}

func (e *Engine) eval(pos chess.Position, isMax bool) float64 {
	var value float64
	var materialVal float64
	var positionalVal float64
	var mobilityVal float64
	for square, piece := range pos.Board().SquareMap() {
		if piece.Color().String() == "w" {
			materialVal += pieceVal[piece.Type().String()]
			positionalVal += e.getPieceSquareTable(pos, &square, &piece)
		} else {
			materialVal -= pieceVal[piece.Type().String()]
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
	if piece.Color().String() == "w" {
		if piece.Type().String() == "p" {
			return pawnTable[sq]
		}
		if piece.Type().String() == "n" {
			return knightTable[sq]
		}
		if piece.Type().String() == "b" {
			return bishopTable[sq]
		}
		if piece.Type().String() == "r" {
			return rookTable[sq]
		}
		if piece.Type().String() == "q" {
			return queenTable[sq]
		}
		if piece.Type().String() == "k" {
			gamePhase := e.getGamePhase(pos)
			return gamePhase*kingMiddleGameTable[sq] + (1-gamePhase)*kingEndGameTable[sq]
		}
	} else {
		if piece.Type().String() == "p" {
			return revPawnTable[sq]
		}
		if piece.Type().String() == "n" {
			return revKnightTable[sq]
		}
		if piece.Type().String() == "b" {
			return revBishopTable[sq]
		}
		if piece.Type().String() == "r" {
			return revRookTable[sq]
		}
		if piece.Type().String() == "q" {
			return revQueenTable[sq]
		}
		if piece.Type().String() == "k" {
			gamePhase := e.getGamePhase(pos)
			return gamePhase*revKingMiddleGameTable[sq] + (1-gamePhase)*revKingEndGameTable[sq]
		}
	}
	fmt.Println("getPieceSquareTable not working correctly")
	return 0
}

func (e *Engine) getGamePhase(pos chess.Position) float64 {
	totalPieces := len(pos.Board().SquareMap())
	maxPieces := 32
	gamePhase := totalPieces / maxPieces
	return float64(gamePhase)
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

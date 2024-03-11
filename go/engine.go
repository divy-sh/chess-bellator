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
	move, _ := e.alphaBeta(depth, e.game.Position(), math.Inf(-1), math.Inf(1), e.game.Position().Turn() == chess.White)
	return move
}

func (e *Engine) alphaBeta(depth int, pos *chess.Position, alpha float64, beta float64, isMax bool) (*chess.Move, float64) {
	if depth == 0 {
		return nil, e.eval(pos)
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
		pos = pos.Update(move)
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

// returns a basic eval of the position
func (e *Engine) eval(pos *chess.Position) float64 {
	var value float64
	var materialVal float64
	var positionalVal float64
	var mobilityVal float64
	centerControlVal := evaluateCenterControl(pos)
	for square, piece := range pos.Board().SquareMap() {
		if piece.Color() == chess.White {
			materialVal += pieceVal[piece.Type()]
			positionalVal += e.getPieceSquareTable(pos, &square, &piece)
		} else {
			materialVal -= pieceVal[piece.Type()]
			positionalVal -= e.getPieceSquareTable(pos, &square, &piece)
		}
	}
	value = materialVal + positionalVal + mobilityVal + centerControlVal
	gamePhase := e.getGamePhase(pos)
	if pos.String() == "r1bk1b1r/p1p2pp1/2p2n1p/2q5/3p3P/2N3Q1/PPPP1PP1/R1B1K1NR b KQ - 0 12" {
		fmt.Println(value, materialVal, positionalVal, mobilityVal, centerControlVal)
	}
	value = value*gamePhase + (1-gamePhase)*materialVal
	return value
}

func (e *Engine) getPieceSquareTable(pos *chess.Position, square *chess.Square, piece *chess.Piece) float64 {
	sq := int(*square)
	if piece.Color() == chess.White {
		switch piece.Type() {
		case chess.Pawn:
			return pawnTable[sq]
		case chess.Knight:
			return knightTable[sq]
		case chess.Bishop:
			return bishopTable[sq]
		case chess.Rook:
			return rookTable[sq]
		case chess.Queen:
			return queenTable[sq]
		case chess.King:
			gamePhase := e.getGamePhase(pos)
			return gamePhase*kingMiddleGameTable[sq] + (1-gamePhase)*kingEndGameTable[sq]
		}
	} else {
		switch piece.Type() {
		case chess.Pawn:
			return revPawnTable[sq]
		case chess.Knight:
			return revKnightTable[sq]
		case chess.Bishop:
			return revBishopTable[sq]
		case chess.Rook:
			return revRookTable[sq]
		case chess.Queen:
			return revQueenTable[sq]
		case chess.King:
			gamePhase := e.getGamePhase(pos)
			return gamePhase*revKingMiddleGameTable[sq] + (1-gamePhase)*revKingEndGameTable[sq]
		}
	}
	return 0
}

// evaluateCenterControl evaluates the control of the center squares (e4, e5, d4, d5)
func evaluateCenterControl(pos *chess.Position) float64 {
	var score float64
	centerSquares := []chess.Square{chess.E4, chess.E5, chess.D4, chess.D5}

	for _, square := range centerSquares {
		if pos.Board().Piece(square) != chess.NoPiece {
			if pos.Board().Piece(square).Color() == chess.White {
				score += pieceVal[pos.Board().Piece(square).Type()]
			} else {
				score -= pieceVal[pos.Board().Piece(square).Type()]
			}
		}
	}
	return score
}

func (e *Engine) getGamePhase(pos *chess.Position) float64 {
	totalPieces := float64(len(pos.Board().SquareMap()))
	maxPieces := 32.0
	gamePhase := totalPieces / maxPieces
	return float64(gamePhase)
}

package main

import "github.com/corentings/chess"

func evaluate(pos *chess.Position) float64 {
	if !endGame {
		endGame = getGamePhase(pos)
	}
	white := 1.0
	if pos.Turn() == chess.Black {
		white = -1.0
	}

	score := 0.0
	for sq := range boardSquares() {
		pc := pos.Board().Piece(sq)
		if pc == chess.NoPiece {
			continue
		}
		eval := evalPiece(sq, pc, endGame)
		if pc.Color() == chess.White {
			score += eval
		} else {
			score -= eval
		}
	}
	return score * white
}

func evalPiece(sq chess.Square, pc chess.Piece, endGame bool) float64 {
	return getPositionalValue(sq, pc, endGame) + pieceValue[pc.Type()]
}

func evaluateMove(mv *chess.Move, pos *chess.Position) float64 {
	white := 1.0
	if pos.Turn() == chess.Black {
		white = -1.0
	}

	pc := pos.Board().Piece(mv.S1())
	if pc == chess.NoPiece {
		return 0
	}
	if mv.Promo() != chess.NoPieceType {
		return 1e9 * white
	}

	posChange := getPositionalValue(mv.S2(), pc, endGame) - getPositionalValue(mv.S1(), pc, endGame)
	matDiff := 0.0
	if isCapture(mv, pos) {
		matDiff = evalCapture(mv, pos)
	}
	return (posChange + matDiff) * white
}

func evalCapture(mv *chess.Move, pos *chess.Position) float64 {
	captured := pos.Board().Piece(mv.S2())
	moved := pos.Board().Piece(mv.S1())
	if captured == chess.NoPiece {
		return 0
	}
	return pieceValue[captured.Type()] - pieceValue[moved.Type()]
}

func getGamePhase(pos *chess.Position) bool {
	minorPieces := 0
	queens := 0
	for sq := range boardSquares() {
		pc := pos.Board().Piece(sq)
		if pc == chess.NoPiece {
			continue
		}
		switch pc.Type() {
		case chess.Bishop, chess.Knight:
			minorPieces++
		case chess.Queen:
			queens++
		}
	}
	return minorPieces < 2 || queens == 0
}

func getPositionalValue(sq chess.Square, pc chess.Piece, endGame bool) float64 {
	if pc.Color() == chess.Black {
		sq = chess.Square(int(sq) ^ 56)
	}
	if pc.Type() == chess.King {
		if endGame {
			return kingEndGameTable[int(sq)]
		}
		return kingMiddleGameTable[int(sq)]
	}
	return piecePositionalScore[pc.Type()][int(sq)]
}

func isCapture(mv *chess.Move, pos *chess.Position) bool {
	return pos.Board().Piece(mv.S2()) != chess.NoPiece
}

func canClaimDraw(pos *chess.Position) bool {
	return pos.Status() != chess.NoMethod
}

func isCheckmate(pos *chess.Position) bool {
	return pos.Status() == chess.Checkmate
}

func boardSquares() chan chess.Square {
	ch := make(chan chess.Square)
	go func() {
		defer close(ch)
		for i := 0; i < 64; i++ {
			ch <- chess.Square(i)
		}
	}()
	return ch
}

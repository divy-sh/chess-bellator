package main

import "github.com/corentings/chess"

func evaluateFast(fb *fastBoard, turn chess.Color) float64 {
	mgScore := 0.0
	egScore := 0.0
	minorPieces := 0
	queens := 0

	for sq := 0; sq < 64; sq++ {
		pc := fb[sq]
		if pc == chess.NoPiece {
			continue
		}

		pType := pc.Type()
		switch pType {
		case chess.Bishop, chess.Knight:
			minorPieces++
		case chess.Queen:
			queens++
		}

		val := pieceValue[pType]
		sqInt := sq
		if pc.Color() == chess.Black {
			sqInt ^= 56
		}

		var posMG, posEG float64
		if pType == chess.King {
			posMG = kingMiddleGameTable[sqInt]
			posEG = kingEndGameTable[sqInt]
		} else {
			posMG = piecePositionalScore[pType][sqInt]
			posEG = posMG
		}

		if pc.Color() == chess.White {
			mgScore += val + posMG
			egScore += val + posEG
		} else {
			mgScore -= val + posMG
			egScore -= val + posEG
		}
	}

	isEndGame := (minorPieces < 2 || queens == 0)
	finalScore := mgScore
	if isEndGame {
		finalScore = egScore
	}

	if turn == chess.Black {
		return -finalScore
	}
	return finalScore
}
func evaluateMove(mv *chess.Move, fb *fastBoard) float64 {
	from, to := int(mv.S1()), int(mv.S2())
	pc := fb[from]
	if pc == chess.NoPiece {
		return 0
	}

	if mv.Promo() != chess.NoPieceType {
		return 1e9 // Always search promotions first
	}

	// Calculate positional change directly from array values (using false/midgame for quick sort estimation)
	posChange := getPositionalValueFast(to, pc, false) - getPositionalValueFast(from, pc, false)

	matDiff := 0.0
	captured := fb[to]
	if captured != chess.NoPiece {
		// Inlined MVV-LVA (Most Valuable Victim - Least Valuable Assault)
		matDiff = pieceValue[captured.Type()] - (pieceValue[pc.Type()] / 100.0)
	}

	return posChange + matDiff
}

func getPositionalValueFast(sqInt int, pc chess.Piece, endGame bool) float64 {
	if pc.Color() == chess.Black {
		sqInt ^= 56
	}
	if pc.Type() == chess.King {
		if endGame {
			return kingEndGameTable[sqInt]
		}
		return kingMiddleGameTable[sqInt]
	}
	return piecePositionalScore[pc.Type()][sqInt]
}

func canClaimDraw(pos *chess.Position) bool {
	// Fastest check: returns true if the current position meets any draw criteria
	return pos.Status() != chess.NoMethod && pos.Status() != chess.Checkmate
}

func isCheckmate(pos *chess.Position) bool {
	return pos.Status() == chess.Checkmate
}

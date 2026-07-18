package main

import (
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/corentings/chess"
)

const (
	mateScore = 100000.0
)

var (
	stopSearch bool
	tt         *TranspositionTable
)

func init() {
	// Initialize the lock-free flat table with 2^21 elements (~96MB RAM footprint)
	tt = NewTranspositionTable(21)
}

type fastBoard [64]chess.Piece

type moveResult struct {
	move  *chess.Move
	score float64
}

func GenMove(maxDepth int, pos *chess.Position) (*chess.Move, float64, bool) {
	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return nil, 0, false
	}

	var rootFb fastBoard
	board := pos.Board()
	for sq := 0; sq < 64; sq++ {
		rootFb[sq] = board.Piece(chess.Square(sq))
	}

	sort.Slice(moves, func(i, j int) bool {
		return evaluateMove(moves[j], &rootFb) > evaluateMove(moves[i], &rootFb)
	})

	var absoluteBestMove *chess.Move
	var absoluteBestScore = -math.MaxFloat64

	for currentDepth := 1; currentDepth <= maxDepth; currentDepth++ {
		if stopSearch {
			break
		}

		if absoluteBestMove != nil {
			for i, mv := range moves {
				if mv.S1() == absoluteBestMove.S1() && mv.S2() == absoluteBestMove.S2() && mv.Promo() == absoluteBestMove.Promo() {
					moves[0], moves[i] = moves[i], moves[0]
					break
				}
			}
		}

		pvMove := moves[0]
		child := pos.Update(pvMove)

		var pvFb fastBoard
		copy(pvFb[:], rootFb[:])
		from, to := int(pvMove.S1()), int(pvMove.S2())
		capturedPiece := pvFb[to]
		movingPiece := pvFb[from]
		pvFb[to] = movingPiece
		pvFb[from] = chess.NoPiece

		childHash := child.Hash()

		bestScore := -negamaxFast(currentDepth-1, -math.MaxFloat64, math.MaxFloat64, child, &pvFb, child.Turn(), childHash)
		bestMove := pvMove

		pvFb[from] = movingPiece
		pvFb[to] = capturedPiece

		if len(moves) == 1 {
			absoluteBestMove = bestMove
			absoluteBestScore = bestScore
			continue
		}

		remainingMoves := moves[1:]
		moveChan := make(chan *chess.Move, len(remainingMoves))
		for _, mv := range remainingMoves {
			moveChan <- mv
		}
		close(moveChan)

		numWorkers := runtime.NumCPU()
		var wg sync.WaitGroup

		var searchMu sync.Mutex
		alpha := bestScore

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var localFb fastBoard
				copy(localFb[:], rootFb[:])

				for mv := range moveChan {
					searchMu.Lock()
					currentAlpha := alpha
					searchMu.Unlock()

					localChild := pos.Update(mv)
					localChildHash := localChild.Hash()

					lFrom, lTo := int(mv.S1()), int(mv.S2())
					lCaptured := localFb[lTo]
					lMoving := localFb[lFrom]
					localFb[lTo] = lMoving
					localFb[lFrom] = chess.NoPiece

					value := -negamaxFast(currentDepth-1, -currentAlpha-1.0, -currentAlpha, localChild, &localFb, localChild.Turn(), localChildHash)

					if value > currentAlpha {
						value = -negamaxFast(currentDepth-1, -math.MaxFloat64, -currentAlpha, localChild, &localFb, localChild.Turn(), localChildHash)
					}

					localFb[lFrom] = lMoving
					localFb[lTo] = lCaptured

					searchMu.Lock()
					if value > bestScore {
						bestScore = value
						bestMove = mv
						if value > alpha {
							alpha = value
						}
					}
					searchMu.Unlock()
				}
			}()
		}
		wg.Wait()

		absoluteBestMove = bestMove
		absoluteBestScore = bestScore
	}

	return absoluteBestMove, absoluteBestScore, false
}

func negamaxFast(depth int, alpha, beta float64, pos *chess.Position, fb *fastBoard, currentTurn chess.Color, currentHash [16]byte) float64 {
	if pos == nil {
		return 0
	}

	// LOCK-FREE TT LOOKUP (Using uncorrupted accurate hashes)
	var ttMove *chess.Move
	if entry, found := tt.Get(currentHash); found {
		if entry.Depth >= depth {
			if entry.Type == Exact {
				return entry.Score
			} else if entry.Type == Lower && entry.Score >= beta {
				return beta
			} else if entry.Type == Upper && entry.Score <= alpha {
				return alpha
			}
		}
		ttMove = entry.BestMove
	}

	status := pos.Status()
	if status == chess.Checkmate {
		return -mateScore * float64(depth+1)
	}
	if status != chess.NoMethod {
		return 0
	}

	if depth == 0 {
		return evaluateFast(fb, currentTurn)
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return 0
	}

	var moveScores [256]float64
	for i, mv := range moves {
		if ttMove != nil && mv.S1() == ttMove.S1() && mv.S2() == ttMove.S2() && mv.Promo() == ttMove.Promo() {
			moveScores[i] = 10000000.0
		} else if mv.Promo() != chess.NoPieceType {
			moveScores[i] = 1000000.0
		} else {
			targetPiece := fb[int(mv.S2())]
			if targetPiece != chess.NoPiece {
				moveScores[i] = 10000.0 + pieceValue[targetPiece.Type()] - (pieceValue[fb[int(mv.S1())].Type()] / 100.0)
			} else {
				moveScores[i] = 0.0
			}
		}
	}

	for i := 0; i < len(moves)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(moves); j++ {
			if moveScores[j] > moveScores[maxIdx] {
				maxIdx = j
			}
		}
		if maxIdx != i {
			moves[i], moves[maxIdx] = moves[maxIdx], moves[i]
			moveScores[i], moveScores[maxIdx] = moveScores[maxIdx], moveScores[i]
		}
	}

	originalAlpha := alpha
	var bestValue = -math.MaxFloat64
	var bestMove *chess.Move

	for _, mv := range moves {
		from, to := int(mv.S1()), int(mv.S2())
		capturedPiece := fb[to]
		movingPiece := fb[from]

		fb[to] = movingPiece
		fb[from] = chess.NoPiece

		child := pos.Update(mv)

		// FIXED: Get the authentic hash straight from the engine position object
		nextHash := child.Hash()

		value := -negamaxFast(depth-1, -beta, -alpha, child, fb, child.Turn(), nextHash)

		fb[from] = movingPiece
		fb[to] = capturedPiece

		if value > bestValue {
			bestValue = value
			bestMove = mv
		}

		if value > alpha {
			alpha = value
		}

		if alpha >= beta {
			break
		}
	}

	var entryType TTEntryType = Exact
	if bestValue <= originalAlpha {
		entryType = Upper
	} else if bestValue >= beta {
		entryType = Lower
	}
	tt.Put(currentHash, bestValue, depth, entryType, bestMove)

	return bestValue
}

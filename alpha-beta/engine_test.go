package main

import (
	"testing"

	"github.com/corentings/chess"
)

func TestGenMoveReturnsLegalMoveFromStartingPosition(t *testing.T) {
	game := chess.NewGame()
	move, _, _ := GenMove(1, game.Position())
	if move == nil {
		t.Fatal("expected a move, got nil")
	}

	for _, legal := range game.Position().ValidMoves() {
		if legal.String() == move.String() {
			return
		}
	}

	t.Fatalf("generated illegal move %s", move.String())
}

func TestGameWrapperAcceptsAndAppliesMove(t *testing.T) {
	game := NewGameWrapper()
	if !game.PlayMove("e4") {
		t.Fatal("expected e4 to be accepted")
	}
	if game.Position().Turn() != chess.Black {
		t.Fatalf("expected black to move after e4, got %s", game.Position().Turn().String())
	}
}

package main

import (
	"fmt"

	"github.com/notnil/chess"
)

func main() {
	run()
}

func run() {
	game := chess.NewGame()
	engine := newEngine(game)
	for game.Outcome() == chess.NoOutcome {
		move := engine.genMoveIterative(1)
		game.Move(move)
		fmt.Println(game.Position().Board().Draw(), engine.transposTableHits, move)
	}
	fmt.Println(game.Outcome())
	fmt.Println(game.String())
}

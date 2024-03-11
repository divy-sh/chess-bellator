package main

import (
	"fmt"
	"time"

	"github.com/notnil/chess"
)

func main() {
	run()
}

func run() {
	defer timer("main method")()
	game := chess.NewGame()
	engine := newEngine(game)
	for game.Outcome() == chess.NoOutcome {
		move := engine.genMove(5)
		game.Move(move)
		fmt.Println(game.Position().Board().Draw(), move)
	}
	fmt.Println(game.Outcome())
	fmt.Println(game.String())
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

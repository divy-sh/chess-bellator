package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/corentings/chess"
)

type pageData struct {
	PGN           string
	Message       string
	Turn          string
	PendingSquare string
	Squares       []squareButton
}

type squareButton struct {
	Name  string
	Label string
	Class string
}

func StartWebUI() {
	game := NewGameWrapper()
	pendingSelection := ""

	handler := func(w http.ResponseWriter, r *http.Request) {
		message := "Click a piece to move, or type below."
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if square := strings.TrimSpace(r.FormValue("square")); square != "" {
				switch pendingSelection {
				case "":
					pendingSelection = square
					message = fmt.Sprintf("Selected %s.", pendingSelection)
				case square:
					pendingSelection = ""
					message = "Selection cleared."
				default:
					candidate := pendingSelection + square
					if game.PlayMove(candidate) {
						pendingSelection = ""
						message = fmt.Sprintf("Played %s", candidate)
						if !game.GameOver() && game.Position().Turn() == chess.Black {
							game.PlayBestMove()
						}
					} else {
						// Fallback: if clicking another piece of your own color, change selection
						pendingSelection = square
						message = fmt.Sprintf("Selected %s.", pendingSelection)
					}
				}
			} else if move := strings.TrimSpace(r.FormValue("move")); move != "" {
				if game.PlayMove(move) {
					pendingSelection = ""
					message = fmt.Sprintf("Played %s", move)
					if !game.GameOver() && game.Position().Turn() == chess.Black {
						game.PlayBestMove()
					}
				} else {
					message = fmt.Sprintf("Invalid move: %s", move)
				}
			}
		}

		data := pageData{
			PGN:           game.PGN(),
			Turn:          game.Position().Turn().String(),
			Message:       message,
			PendingSquare: pendingSelection,
			Squares:       buildSquareButtons(game),
		}

		tmpl := template.Must(template.New("board").Parse(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Chess Bellator</title>
  <style>
    body { font-family: sans-serif; background: #2f241d; color: #f3e7d3; padding: 20px; display: flex; justify-content: center; }
    .panel { width: 480px; }
    .board { display: grid; grid-template-columns: repeat(8, 60px); grid-template-rows: repeat(8, 60px); box-shadow: 0 4px 10px rgba(0,0,0,0.5); }
    .square { width: 60px; height: 60px; border: none; font-size: 40px; display: flex; align-items: center; justify-content: center; cursor: pointer; padding: 0; user-select: none; }
    .light { background: #f0d9b5; color: #b58863; }
    .dark { background: #b58863; color: #f0d9b5; }
    .square:hover { opacity: 0.9; }
    .selected { box-shadow: inset 0 0 0 4px #ffbf00; }
    input, button { padding: 8px; margin-top: 5px; border-radius: 4px; border: 1px solid #6b4f2d; }
    input { width: 200px; background: #f3e7d3; }
    button { background: #b58863; color: white; cursor: pointer; }
    form { margin: 0; }
  </style>
</head>
<body>
  <div class="panel">
    <h2>Chess Bellator</h2>
    <div class="board">
      {{range .Squares}}
      <form method="post">
        <input type="hidden" name="square" value="{{.Name}}" />
        <button class="square {{.Class}} {{if eq .Name $.PendingSquare}}selected{{end}}" type="submit">{{.Label}}</button>
      </form>
      {{end}}
    </div>
    <p><strong>Turn:</strong> {{.Turn}} | <strong>Status:</strong> {{.Message}}</p>
    <p><small><strong>PGN:</strong> {{if .PGN}}{{.PGN}}{{else}}None{{end}}</small></p>
    <form method="post">
      <input name="move" placeholder="e.g., e4, Nf3, or e2e4" />
      <button type="submit">Play</button>
    </form>
  </div>
</body>
</html>`))
		_ = tmpl.Execute(w, data)
	}

	http.HandleFunc("/", handler)
	fmt.Println("Open http://localhost:8080 to play")
	_ = http.ListenAndServe(":8080", nil)
}

func buildSquareButtons(game *GameWrapper) []squareButton {
	buttons := make([]squareButton, 0, 64)
	for rank := 8; rank >= 1; rank-- {
		for file := 'a'; file <= 'h'; file++ {
			squareStr := fmt.Sprintf("%c%d", file, rank)
			piece := game.Position().Board().Piece(squareToSquare(squareStr))

			class := "light"
			if (rank+int(file))%2 == 0 {
				class = "dark"
			}

			buttons = append(buttons, squareButton{
				Name:  squareStr,
				Label: getUnicodePiece(piece),
				Class: class,
			})
		}
	}
	return buttons
}

func getUnicodePiece(piece chess.Piece) string {
	if piece == chess.NoPiece {
		return ""
	}
	// Maps chess piece types to their corresponding visual unicode characters
	switch piece.Color() {
	case chess.White:
		switch piece.Type() {
		case chess.King:
			return "♔"
		case chess.Queen:
			return "♕"
		case chess.Rook:
			return "♖"
		case chess.Bishop:
			return "♗"
		case chess.Knight:
			return "♘"
		case chess.Pawn:
			return "♙"
		}
	case chess.Black:
		switch piece.Type() {
		case chess.King:
			return "♚"
		case chess.Queen:
			return "♛"
		case chess.Rook:
			return "♜"
		case chess.Bishop:
			return "♝"
		case chess.Knight:
			return "♞"
		case chess.Pawn:
			return "♟"
		}
	}
	return ""
}

func squareToSquare(s string) chess.Square {
	if len(s) != 2 {
		return chess.NoSquare
	}
	return chess.NewSquare(chess.File(s[0]-'a'), chess.Rank(s[1]-'1'))
}

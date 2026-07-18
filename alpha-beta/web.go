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
	Name       string
	Label      string
	Class      string // Controls the square color (light/dark)
	PieceClass string // Controls the piece color (p-white/p-black)
}

func StartWebUI() {
	game := NewGameWrapper()
	pendingSelection := ""

	handler := func(w http.ResponseWriter, r *http.Request) {
		message := "Your turn. Click a piece, then click a destination square."

		if r.Method == http.MethodPost {
			_ = r.ParseForm()

			if square := strings.TrimSpace(r.FormValue("square")); square != "" {
				switch pendingSelection {
				case "":
					pendingSelection = square
					message = fmt.Sprintf("Selected %s.", square)
				case square:
					pendingSelection = ""
					message = "Selection cleared."
				default:
					moveAttempt := pendingSelection + square
					pendingSelection = ""

					if game.PlayMove(moveAttempt) {
						message = fmt.Sprintf("You played %s.", moveAttempt)

						if !game.GameOver() {
							if _, ok := game.PlayBestMove(); !ok {
								message += " AI failed to respond."
							}
						}
					} else {
						message = "Invalid move selection."
					}
				}
			}
		}

		if game.GameOver() {
			message = "Game Over!"
		}

		data := pageData{
			PGN:           game.PGN(),
			Turn:          strings.Title(game.Position().Turn().String()),
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
    :root {
      --bg-dark: #1e1e24;
      --panel-bg: #2a2a35;
      --light-square: #eae9d2;
      --dark-square: #4b7355;
      --accent: #e5a93c;
      --white-piece: #ffffff;
      --black-piece: #111116;
    }
    body { 
      font-family: system-ui, -apple-system, sans-serif; 
      background: var(--bg-dark); 
      color: #f3e7d3; 
      padding: 40px 20px; 
      display: flex; 
      justify-content: center;
      margin: 0;
    }
    .panel { 
      width: 480px; 
      background: var(--panel-bg);
      padding: 24px;
      border-radius: 16px;
      box-shadow: 0 10px 30px rgba(0,0,0,0.4);
    }
    h2 { margin-top: 0; text-align: center; letter-spacing: 1px; color: #fff; }
    
    .turn-badge {
      text-align: center;
      padding: 6px 12px;
      border-radius: 20px;
      font-weight: bold;
      margin-bottom: 20px;
      background: rgba(0,0,0,0.2);
      display: inline-block;
      width: calc(100% - 24px);
    }
    .turn-white { color: #eae9d2; border: 1px solid rgba(234,233,210, 0.3); }
    .turn-black { color: #e5a93c; border: 1px solid rgba(229,169,60, 0.3); }

    .board { 
      display: grid; 
      grid-template-columns: repeat(8, 60px); 
      grid-template-rows: repeat(8, 60px); 
      border-radius: 8px;
      overflow: hidden;
      box-shadow: 0 8px 24px rgba(0,0,0,0.3);
      margin: 0 auto 20px auto;
      width: 480px;
    }
    .square { 
      width: 60px; 
      height: 60px; 
      border: none; 
      font-size: 44px; 
      display: flex; 
      align-items: center; 
      justify-content: center; 
      cursor: pointer; 
      padding: 0; 
      user-select: none; 
      transition: background 0.15s ease;
    }
    .light { background: var(--light-square); }
    .dark { background: var(--dark-square); }
    
    /* Solid chess piece custom styling */
    .p-white { color: var(--white-piece); filter: drop-shadow(0px 2px 3px rgba(0,0,0,0.6)); }
    .p-black { color: var(--black-piece); filter: drop-shadow(0px 1px 1px rgba(255,255,255,0.4)); }

    .square:hover { filter: brightness(1.1); }
    .selected { background: var(--accent) !important; }
    
    .status-text {
      background: rgba(0,0,0,0.15);
      padding: 12px;
      border-radius: 8px;
      font-size: 0.95rem;
      border-left: 4px solid var(--accent);
      margin-bottom: 15px;
    }
    .pgn-box {
      font-family: monospace;
      font-size: 0.85rem;
      background: #1e1e24;
      padding: 10px;
      border-radius: 6px;
      max-height: 60px;
      overflow-y: auto;
      color: #a0a0b0;
    }
    form { margin: 0; display: inline; }
  </style>
</head>
<body>
  <div class="panel">
    <h2>Chess Bellator</h2>
    <div class="turn-badge {{if eq .Turn "White"}}turn-white{{else}}turn-black{{end}}">
      {{.Turn}} to Move
    </div>
    
    <div class="board">
      {{range .Squares}}
      <form method="post">
        <input type="hidden" name="square" value="{{.Name}}" />
        <button class="square {{.Class}} {{if eq .Name $.PendingSquare}}selected{{end}}" type="submit">
          <span class="{{.PieceClass}}">{{.Label}}</span>
        </button>
      </form>
      {{end}}
    </div>
    
    <div class="status-text"><strong>Status:</strong> {{.Message}}</div>
    <div><small style="color:#aaa;"><strong>PGN History:</strong></small></div>
    <div class="pgn-box">{{if .PGN}}{{.PGN}}{{else}}<em>No moves played yet.</em>{{end}}</div>
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

			pieceClass := ""
			if piece != chess.NoPiece {
				if piece.Color() == chess.White {
					pieceClass = "p-white"
				} else {
					pieceClass = "p-black"
				}
			}

			buttons = append(buttons, squareButton{
				Name:       squareStr,
				Label:      getSolidPiece(piece),
				Class:      class,
				PieceClass: pieceClass,
			})
		}
	}
	return buttons
}

func getSolidPiece(piece chess.Piece) string {
	if piece == chess.NoPiece {
		return ""
	}
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
	return ""
}

func squareToSquare(s string) chess.Square {
	if len(s) != 2 {
		return chess.NoSquare
	}
	return chess.NewSquare(chess.File(s[0]-'a'), chess.Rank(s[1]-'1'))
}

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

					// Get the requested promotion piece, default to Queen
					promo := strings.TrimSpace(r.FormValue("promo"))
					if promo == "" {
						promo = "q"
					}

					movePlayed := ""

					// 1. Try a standard move first
					if game.PlayMove(moveAttempt) {
						movePlayed = moveAttempt
					} else if game.PlayMove(moveAttempt + promo) {
						// 2. If standard fails, try appending the promotion character
						movePlayed = moveAttempt + "=" + strings.ToUpper(promo)
					}

					if movePlayed != "" {
						message = fmt.Sprintf("You played %s.", movePlayed)

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
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
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
      margin: 0;
      min-height: 100vh;
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 20px;
      box-sizing: border-box;
    }
    
    /* Default: Stacked Layout (Mobile/Narrow/Long Screens) */
    .panel { 
      background: var(--panel-bg);
      padding: 24px;
      border-radius: 16px;
      box-shadow: 0 10px 30px rgba(0,0,0,0.4);
      display: flex;
      flex-direction: column;
      gap: 24px;
      width: 100%;
      max-width: 600px;
      box-sizing: border-box;
    }

    /* Desktop/Wide Screens */
    @media (min-width: 900px) {
      .panel {
        flex-direction: row;
        max-width: 1200px;
        width: 95vw;
        align-items: stretch;
      }
      .sidebar {
        width: 320px;
        min-width: 320px;
        display: flex;
        flex-direction: column;
      }
      .board-container {
        flex: 1;
        display: flex;
        justify-content: center;
        align-items: center;
      }
    }

    .board { 
      display: grid; 
      grid-template-columns: repeat(8, 1fr); 
      grid-template-rows: repeat(8, 1fr); 
      border-radius: 8px;
      overflow: hidden;
      box-shadow: 0 8px 24px rgba(0,0,0,0.3);
      width: 100%;
      aspect-ratio: 1;
      max-width: 80vh; /* Stops board from overflowing vertically on wide screens */
    }

    .sidebar h2 { 
      margin-top: 0; 
      text-align: center; 
      letter-spacing: 1px; 
      color: #fff; 
      margin-bottom: 20px;
    }
    
    .turn-badge {
      text-align: center;
      padding: 10px 12px;
      border-radius: 20px;
      font-weight: bold;
      margin-bottom: 15px;
      background: rgba(0,0,0,0.2);
    }
    .turn-white { color: #eae9d2; border: 1px solid rgba(234,233,210, 0.3); }
    .turn-black { color: #e5a93c; border: 1px solid rgba(229,169,60, 0.3); }

    .controls {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 15px;
      background: rgba(0,0,0,0.15);
      padding: 12px;
      border-radius: 8px;
    }
    .controls select {
      background: var(--bg-dark);
      color: #fff;
      border: 1px solid #444;
      border-radius: 4px;
      padding: 6px 8px;
      outline: none;
    }

    .square { 
      width: 100%; 
      height: 100%; 
      border: none; 
      font-size: var(--piece-size, 44px); 
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
    
    .pgn-container {
      display: flex;
      flex-direction: column;
      flex: 1;
      min-height: 100px;
    }
    
    .pgn-box {
      font-family: monospace;
      font-size: 0.85rem;
      background: #1e1e24;
      padding: 12px;
      border-radius: 6px;
      flex: 1;
      overflow-y: auto;
      color: #a0a0b0;
      word-break: break-word;
      max-height: 250px;
    }

    @media (max-width: 899px) {
      .pgn-box { max-height: 120px; }
    }
  </style>
</head>
<body>
  <!-- By wrapping everything in the form, layout scaling natively captures both board clicks & options -->
  <form method="post" class="panel">
    <div class="board-container">
      <div class="board">
        {{range .Squares}}
          <button class="square {{.Class}} {{if eq .Name $.PendingSquare}}selected{{end}}" type="submit" name="square" value="{{.Name}}">
            <span class="{{.PieceClass}}">{{.Label}}</span>
          </button>
        {{end}}
      </div>
    </div>
    
    <div class="sidebar">
      <h2>Chess Bellator</h2>
      <div class="turn-badge {{if eq .Turn "White"}}turn-white{{else}}turn-black{{end}}">
        {{.Turn}} to Move
      </div>
      
      <div class="controls">
        <span style="font-size: 0.9rem; color: #a0a0b0;"><strong>Auto-Promote To:</strong></span>
        <select name="promo">
          <option value="q">Queen (♛)</option>
          <option value="r">Rook (♜)</option>
          <option value="b">Bishop (♝)</option>
          <option value="n">Knight (♞)</option>
        </select>
      </div>
      
      <div class="status-text"><strong>Status:</strong> {{.Message}}</div>
      
      <div class="pgn-container">
        <div style="margin-bottom: 8px;"><small style="color:#aaa;"><strong>PGN History:</strong></small></div>
        <div class="pgn-box">{{if .PGN}}{{.PGN}}{{else}}<em>No moves played yet.</em>{{end}}</div>
      </div>
    </div>
  </form>

  <script>
    // Dynamically adjust Unicode piece sizing to perfectly fit responsive grid cells
    function resizePieces() {
      const board = document.querySelector('.board');
      const firstSquare = document.querySelector('.square');
      if (firstSquare) {
        // Calculate 75% of the square's width for optimal padding
        const size = firstSquare.getBoundingClientRect().width * 0.75;
        board.style.setProperty('--piece-size', size + 'px');
      }
    }
    
    // Bind to window events
    window.addEventListener('resize', resizePieces);
    window.addEventListener('DOMContentLoaded', resizePieces);
    
    // Safety fallback for initial DOM rendering phase delays
    setTimeout(resizePieces, 100);
  </script>
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

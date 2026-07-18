package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/corentings/chess"
)

type pageData struct {
	PGN           string
	Message       string
	Turn          string
	PendingSquare string
	AIPending     bool
	LockedBoard   bool
	Squares       []squareButton
}

type squareButton struct {
	Name       string
	Label      string
	Class      string
	PieceClass string
	IsLastMove bool
}

func StartWebUI() {
	game := NewGameWrapper()
	pendingSelection := ""
	lastSrc := ""
	lastDst := ""
	playerColor := chess.White

	var mu sync.Mutex

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		message := "Your turn. Click a piece, then a destination square."
		if game.GameOver() {
			message = "Game Over! Start a new game."
		}

		aiPending := false

		if r.Method == http.MethodPost {
			_ = r.ParseForm()

			// Differentiate between button clicks and automatic AI triggers
			action := r.FormValue("action")
			if action == "" {
				action = r.FormValue("auto_action")
			}

			if action == "new_white" {
				game = NewGameWrapper()
				playerColor = chess.White
				pendingSelection = ""
				lastSrc, lastDst = "", ""
				message = "New game started. You are White."
			} else if action == "new_black" {
				game = NewGameWrapper()
				playerColor = chess.Black
				pendingSelection = ""
				lastSrc, lastDst = "", ""
				message = "New game started. You are Black. AI is thinking..."
				aiPending = true
			} else if action == "aimove" {
				if !game.GameOver() && game.Position().Turn() != playerColor {
					if aiMove, ok := game.PlayBestMove(); ok {
						aiMoveStr := fmt.Sprintf("%v", aiMove)
						if len(aiMoveStr) >= 4 {
							lastSrc = aiMoveStr[:2]
							lastDst = aiMoveStr[2:4]
						}
						message = "AI played. Your turn."
						if game.GameOver() {
							message = "Game Over!"
						}
					} else {
						message = "AI failed to respond."
					}
				}
			} else if square := strings.TrimSpace(r.FormValue("square")); square != "" {
				if game.GameOver() {
					message = "Game Over! Please start a new game."
					pendingSelection = ""
				} else if game.Position().Turn() != playerColor {
					message = "Please wait for the AI to move."
					pendingSelection = ""
				} else {
					switch pendingSelection {
					case "":
						p := game.Position().Board().Piece(squareToSquare(square))
						if p != chess.NoPiece && p.Color() == playerColor {
							pendingSelection = square
							message = fmt.Sprintf("Selected %s.", square)
						} else {
							message = "Please select one of your own pieces."
						}
					case square:
						pendingSelection = ""
						message = "Selection cleared."
					default:
						moveAttempt := pendingSelection + square
						pendingSelection = ""

						promo := strings.TrimSpace(r.FormValue("promo"))
						if promo == "" {
							promo = "q"
						}

						movePlayed := ""
						if game.PlayMove(moveAttempt) {
							movePlayed = moveAttempt
						} else if game.PlayMove(moveAttempt + promo) {
							movePlayed = moveAttempt + "=" + strings.ToUpper(promo)
						}

						if movePlayed != "" {
							message = fmt.Sprintf("You played %s. AI is thinking...", movePlayed)
							lastSrc = moveAttempt[:2]
							lastDst = moveAttempt[2:4]

							if !game.GameOver() {
								aiPending = true
							} else {
								message = "Game Over!"
							}
						} else {
							message = "Invalid move selection."
						}
					}
				}
			}
		}

		isLocked := aiPending || game.GameOver()

		data := pageData{
			PGN:           game.PGN(),
			Turn:          strings.Title(game.Position().Turn().String()),
			Message:       message,
			PendingSquare: pendingSelection,
			AIPending:     aiPending,
			LockedBoard:   isLocked,
			Squares:       buildSquareButtons(game, lastSrc, lastDst, playerColor),
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
      max-width: 80vh;
    }

    .locked {
      pointer-events: none;
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
      flex-direction: column;
      gap: 12px;
      margin-bottom: 15px;
      background: rgba(0,0,0,0.15);
      padding: 12px;
      border-radius: 8px;
    }

    .btn-row {
      display: flex;
      gap: 8px;
      width: 100%;
    }

    .btn {
      background: var(--dark-square);
      color: #fff;
      border: none;
      padding: 10px;
      border-radius: 6px;
      cursor: pointer;
      flex: 1;
      font-weight: bold;
      transition: filter 0.15s ease;
    }
    .btn:hover { filter: brightness(1.2); }

    .promo-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .controls select {
      background: var(--bg-dark);
      color: #fff;
      border: 1px solid #444;
      border-radius: 4px;
      padding: 6px 8px;
      outline: none;
      flex: 1;
      margin-left: 10px;
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
    
    .light.last-move { background: #f6f669; }
    .dark.last-move { background: #baca44; }

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
      transition: all 0.3s ease;
    }

    .thinking-anim {
      animation: pulseBG 1s infinite alternate;
      border-left-color: #f6f669;
    }

    @keyframes pulseBG {
      0% { background: rgba(0,0,0,0.15); }
      100% { background: rgba(229, 169, 60, 0.25); }
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
  <form method="post" class="panel" id="chess-form">
    <!-- Updated the name here to auto_action to avoid conflicts -->
    <input type="hidden" name="auto_action" id="auto-action" value="">

    <div class="board-container">
      <div class="board {{if .LockedBoard}}locked{{end}}" id="chess-board">
        {{range .Squares}}
          <button class="square {{.Class}} {{if .IsLastMove}}last-move{{end}} {{if eq .Name $.PendingSquare}}selected{{end}}" type="submit" name="square" value="{{.Name}}">
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
        <div class="btn-row">
          <button type="submit" name="action" value="new_white" class="btn">New (White)</button>
          <button type="submit" name="action" value="new_black" class="btn">New (Black)</button>
        </div>
        <div class="promo-row">
          <span style="font-size: 0.85rem; color: #a0a0b0;"><strong>Promote To:</strong></span>
          <select name="promo">
            <option value="q">Queen (♛)</option>
            <option value="r">Rook (♜)</option>
            <option value="b">Bishop (♝)</option>
            <option value="n">Knight (♞)</option>
          </select>
        </div>
      </div>
      
      <div id="status-box" class="status-text {{if .AIPending}}thinking-anim{{end}}"><strong>Status:</strong> {{.Message}}</div>
      
      <div class="pgn-container">
        <div style="margin-bottom: 8px;"><small style="color:#aaa;"><strong>PGN History:</strong></small></div>
        <div class="pgn-box">{{if .PGN}}{{.PGN}}{{else}}<em>No moves played yet.</em>{{end}}</div>
      </div>
    </div>
  </form>

  <script>
    function resizePieces() {
      const board = document.querySelector('.board');
      const firstSquare = document.querySelector('.square');
      if (firstSquare) {
        const size = firstSquare.getBoundingClientRect().width * 0.75;
        board.style.setProperty('--piece-size', size + 'px');
      }
    }
    
    window.addEventListener('resize', resizePieces);
    window.addEventListener('DOMContentLoaded', resizePieces);
    setTimeout(resizePieces, 100);

    document.querySelectorAll('.square').forEach(btn => {
      btn.addEventListener('click', () => {
        document.getElementById('chess-board').classList.add('locked');
        document.getElementById('status-box').classList.add('thinking-anim');
      });
    });

    {{if .AIPending}}
      setTimeout(() => {
        // Updated to target the correct input
        document.getElementById('auto-action').value = 'aimove';
        document.getElementById('chess-form').submit();
      }, 50);
    {{end}}
  </script>
</body>
</html>`))

		_ = tmpl.Execute(w, data)
	}

	http.HandleFunc("/", handler)
	fmt.Println("Open http://localhost:8080 to play")
	_ = http.ListenAndServe(":8080", nil)
}

func buildSquareButtons(game *GameWrapper, lastSrc, lastDst string, playerColor chess.Color) []squareButton {
	buttons := make([]squareButton, 0, 64)

	ranks := []int{8, 7, 6, 5, 4, 3, 2, 1}
	files := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	if playerColor == chess.Black {
		ranks = []int{1, 2, 3, 4, 5, 6, 7, 8}
		files = []rune{'h', 'g', 'f', 'e', 'd', 'c', 'b', 'a'}
	}

	for _, rank := range ranks {
		for _, file := range files {
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

			isLastMove := (squareStr == lastSrc || squareStr == lastDst)

			buttons = append(buttons, squareButton{
				Name:       squareStr,
				Label:      getSolidPiece(piece),
				Class:      class,
				PieceClass: pieceClass,
				IsLastMove: isLastMove,
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

from view import View
from game import Game
import engine
import random
import math
import time

def main():
    game = Game()
    view = View(game.getBoard())
    AIvsAI(game, engine, view)
    # playerVsAI(game, engine, view)
    
def AIvsAI(game, engine, view):
    pgn = []
    while True:
        if game.gameOver():
            print("game over!")
            print(game.outcome())
            print(' '.join(pgn))
            return
        start_time = time.time()
        move, eval = engine.genMoveIterative(5, game.getBoard())
        pgn.append(game.getBoard().san(move))
        game.playMove(move.uci())
        view.update()
        print(f"move - {move.uci()}, eval - {eval}, Time taken to run: {time.time() - start_time:.6f} seconds")
        
def playerVsAI(game, engine, view):
    playerWhite = math.ceil(random.random() * 100) % 2 == 0
    playerWhite = True
    while True:
        if game.gameOver():
            print("game over!")
            print(game.outcome())
            return
        if playerWhite:
            move = input("your move: ")
            if not game.playMove(move):
                print("invalid move, try again")
                playerWhite = not playerWhite
        else:
            start_time = time.time()
            move, eval = engine.genMove(4, game.getBoard())
            game.playMove(move.uci())
            view.update()
            print(f"move - {move.uci()}, eval - {eval}, Time taken to run: {time.time() - start_time:.6f} seconds")
        playerWhite = not playerWhite

if __name__ == '__main__':
    main()
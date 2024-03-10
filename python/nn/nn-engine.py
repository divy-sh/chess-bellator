import chess
import chess.engine
import random
import numpy

def genRandomBoard(maxDepth = 100):
    board = chess.Board()
    depth = random.randrange(0, maxDepth)
    for _ in range(depth):
        board.push(random.choice(list(board.legal_moves)))
        if board.is_game_over():
            break
    return board

def reinforcer(board, depth):
    with chess.engine.SimpleEngine.popen_uci('./stockfish') as stockfish:
        return stockfish.analyse(board, chess.engine.Limit(1))['score'].white().score()
    
board = genRandomBoard()

print(reinforcer(board, 10))
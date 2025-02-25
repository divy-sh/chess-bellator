import time
import multiprocessing
from functools import partial

import chess
import eval

def genMoveIterative(seconds: float, board: chess.Board) -> str:    
    timeout = time.time() + seconds
    depth = 2
    move = None
    while time.time() < timeout:
        depth += 1
        move = genMove(depth, board)
    print(f"depth reached - {depth}")
    return move

# generate move with fixed depth alpha-beta search
def genMove(depth: int, board: chess.Board) -> str:
    moves = getOrderedMoves(board=board)
    if not moves:
        return None, 0
    if board.is_checkmate():
        return None, float('-inf')
    
    pool = multiprocessing.Pool()
    eval = partial(evaluate_move, depth, board.copy())
    results = pool.map(eval, moves)
    pool.close()
    pool.join()

    bestValue = float('-inf')
    bestMove = None
    for move, value in results:
        if value >= bestValue:
            bestValue = value
            bestMove = move
    return bestMove, bestValue

def evaluate_move(depth: int, board: chess.Board, move: chess.Move) -> tuple:
    board.push(move)
    value = -negaMax(depth - 1, float('-inf'), float('inf'), board)
    board.pop()
    return move, value

def negaMax(depth: int, alpha: float, beta: float, board: chess.Board) -> float:
    if board.is_checkmate():
        return -100000 * (depth + 1)
    if board.can_claim_draw():
        return 0
    moves = board.legal_moves
    if not moves:
        return 0
    if depth == 0:
        return qSearch(alpha, beta, board)
    
    for move in moves:
        board.push(move)
        value = -negaMax(depth - 1, -beta, -alpha, board)
        board.pop()
        if value >= beta:
            return beta
        alpha = max(alpha, value)
    
    return alpha
    
def qSearch(alpha: float, beta: float, board: chess.Board):
    value = eval.evaluate(board)
    if value >= beta:
        return beta
    alpha = max(alpha, value)
    moves = board.legal_moves
    for move in moves:
        if not board.is_capture(move):
            continue
        board.push(move)
        value = -qSearch( -beta, -alpha, board)
        board.pop()
        if value >= beta:
            return beta
        alpha = max(alpha, value)
    return alpha

def getOrderedMoves(board: chess.Board):
    return sorted(list(board.legal_moves), key=lambda move: eval.evaluateMove(move, board))
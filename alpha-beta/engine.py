import chess
import chess.engine
import time
import eval as evaluation
import multiprocessing
from functools import partial


def genMoveIterative(seconds: float, board: chess.Board) -> str:    
    timeout = time.time() + seconds
    depth = 2
    move = None
    while time.time() < timeout:
        depth += 1
        move = genMove(depth)
    print(f"depth reached - {depth}")
    return move

# generate move with fixed depth alpha-beta search
def genMove(depth: int, board: chess.Board) -> str:
    if board.is_checkmate():
        return None, float('-inf')
    moves = getOrderedMoves(onlyCaptures=False, board=board)
    if not moves:
        return None, 0
    
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
    value = -alphaBeta(depth - 1, float('-inf'), float('inf'), board)
    board.pop()
    return move, value

def alphaBeta(depth: int, alpha: float, beta: float, board: chess.Board) -> float:
    if board.is_checkmate():
        return float('-inf')
    moves = getOrderedMoves(onlyCaptures=False, board=board)
    if not moves:
        return 0
    if depth == 0:
        return qSearch(alpha, beta, board)
    
    for move in moves:
        board.push(move)
        value = -alphaBeta(depth - 1, -beta, -alpha, board)
        board.pop()
        if value >= beta:
            return beta
        alpha = max(alpha, value)
    
    return alpha
    
def qSearch(alpha: float, beta: float, board: chess.Board):
    value = evaluation.evaluate(board)
    if value >= beta:
        return beta
    return value
    alpha = max(alpha, value)
    moves = getOrderedMoves(onlyCaptures=True)
    for move in moves:
        board.push(move)
        value = -qSearch( -beta, -alpha, board)
        board.pop()
        if value >= beta:
            return beta
        alpha = max(alpha, value)
    
    return alpha

def getOrderedMoves(onlyCaptures, board: chess.Board):
    if onlyCaptures:
        return sorted([move for move in board.legal_moves if board.is_capture(move)], key=lambda move: getMovePriority(move))
    return sorted(list(board.legal_moves), key=lambda move: getMovePriority(move, board))

def getMovePriority(move: chess.Move, board: chess.Board) -> int:
    if board.piece_at(move.to_square) is None: 
        return 0
    else: 
        return evaluation.pieceValue[board.piece_at(move.to_square).symbol().upper()]
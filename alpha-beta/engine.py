import chess
import chess.engine
import time
import eval as evaluation

class Engine:
    def __init__(self, board : chess.Board) -> None:
        self.board = board

    #iterative deepening until time runs out
    def genMoveIterative(self, seconds: float) -> str:    
        timeout = time.time() + seconds
        depth = 2
        move = None
        while time.time() < timeout:
            depth += 1
            move = self.genMove(depth)
        print(f"depth reached - {depth}")
        return move

    # generate move with fixed depth alpha-beta search
    def genMove(self, depth: int) -> str:
        move, _ = self.alphaBeta(depth, float('-inf'), float('inf'), self.board.turn, False)
        return move.uci()
    
    def alphaBeta(self, depth: int, alpha: float, beta: float, isMax: bool, lastMoveSpecial: bool) -> tuple[chess.Move, float]:
        if depth == 0:
            if self.board.is_checkmate():
                if isMax:
                    return None, float('-inf')
                else:
                    return None, float('inf')
            # if lastMoveSpecial:
            #     return self.qSearch(alpha, beta, not isMax, lastMoveSpecial)
            return None, evaluation.evaluate(self.board, isMax)
        
        if isMax:
            bestValue = float('-inf')
        else:
            bestValue = float('inf')
        bestMove = None
        moves = sorted(list(self.board.legal_moves), 
                       key=lambda move: (self.board.piece_at(move.to_square) 
                        is None, move.from_square, move.to_square))
        for move in moves:
            isSpecial = self.board.is_capture(move)
            self.board.push_uci(move.uci())
            isSpecial = isSpecial or self.board.is_check()
            _, value = self.alphaBeta(depth - 1, alpha, beta, not isMax, isSpecial)
            self.board.pop()
            if isMax:
                if bestValue <= value:
                    bestValue = value
                    bestMove = move
                alpha = max(alpha, bestValue)
            else:
                if bestValue >= value:
                    bestValue = value
                    bestMove = move
                beta = min(beta, bestValue)
            if beta <= alpha:
                return bestMove, bestValue
        return bestMove, bestValue

    def qSearch(self, alpha: float, beta: float, isMax: bool, lastMoveSpecial: bool):
        if not lastMoveSpecial:
            return None, evaluation.evaluate(self.board, isMax)
        if isMax:
            bestValue = float('-inf')
        else:
            bestValue = float('inf')
        bestMove = None
        for move in self.board.legal_moves:
            isSpecial = self.board.is_capture(move)
            self.board.push_uci(move.uci())
            isSpecial = isSpecial or self.board.is_check()
            _, value = self.qSearch(alpha, beta, not isMax, isSpecial)
            self.board.pop()
            if isMax:
                if bestValue <= value:
                    bestValue = value
                    bestMove = move
                alpha = max(alpha, bestValue)
            else:
                if bestValue >= value:
                    bestValue = value
                    bestMove = move
                beta = min(beta, bestValue)
            if beta <= alpha:
                return bestMove, bestValue
        return bestMove, bestValue

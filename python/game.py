from typing import Optional
import chess

class Game:
    def __init__(self) -> None:
        self.board = chess.Board()
    
    def playMove(self, move: chess.Move) -> bool:
        try:
            self.board.push_san(move)
            return True
        except:
            return False
    
    def gameOver(self) -> bool:
        return self.board.is_game_over()
    
    def outcome(self) -> Optional[chess.Outcome]:
        return self.board.outcome()
    
    def getBoard(self) -> chess.Board:
        return self.board
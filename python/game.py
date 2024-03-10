import chess

class Game:
    def __init__(self) -> None:
        self.board = chess.Board()
    
    def playMove(self, move):
        try:
            self.board.push_san(move)
            return True
        except:
            return False
    
    def gameOver(self):
        return self.board.is_game_over()
    
    def outcome(self):
        return self.board.outcome()
    
    def getBoard(self):
        return self.board
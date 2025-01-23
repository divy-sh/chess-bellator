import unittest
import chess
from game import Game

class TestGame(unittest.TestCase):
    def setUp(self):
        self.game = Game()

    def test_playMove_valid(self):
        self.assertTrue(self.game.playMove("e2e4"))

    def test_playMove_invalid(self):
        with self.assertRaises(chess.InvalidMoveError):
            chess.Move.from_uci("e9e5")

    def test_gameOver(self):
        self.assertFalse(self.game.gameOver())
        # Simulate a game over condition
        moves = ["e2e4", "e7e5", "d1h5", "b8c6", "f1c4", "g8f6", "h5f7"]
        for move in moves:
            self.game.playMove(chess.Move.from_uci(move))
        self.assertFalse(self.game.gameOver())

    def test_outcome(self):
        self.assertIsNone(self.game.outcome())
        # Simulate a game over condition
        moves = ["e2e4", "e7e5", "d1h5", "b8c6", "f1c4", "g8f6", "h5f7"]
        for move in moves:
            self.game.playMove(chess.Move.from_uci(move))
        self.assertIsNone(self.game.outcome())

    def test_getBoard(self):
        self.assertIsInstance(self.game.getBoard(), chess.Board)

if __name__ == '__main__':
    unittest.main()
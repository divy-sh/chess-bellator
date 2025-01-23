import unittest
import chess
from eval import getPositionalValue, get_game_phase, evaluateMove, evalCapture, evalPiece, evaluate, scan

class TestEvaluateFunctions(unittest.TestCase):
    def setUp(self):
        self.board = chess.Board()

    def test_getPositionalValue(self):
        piece = chess.Piece(chess.KING, chess.WHITE)
        value = getPositionalValue(chess.E1, piece, False)
        self.assertIsInstance(value, float)

    def test_get_game_phase(self):
        phase = get_game_phase(self.board)
        self.assertIsInstance(phase, bool)

    def test_evaluateMove(self):
        move = chess.Move.from_uci("e2e4")
        value = evaluateMove(move, self.board)
        self.assertIsInstance(value, float)

    def test_evalCapture(self):
        move = chess.Move.from_uci("e2e4")
        self.board.push(move)
        move = chess.Move.from_uci("d7d5")
        self.board.push(move)
        move = chess.Move.from_uci("e4d5")
        value = evalCapture(move, self.board)
        self.assertIsInstance(value, int)

    def test_evalPiece(self):
        piece = chess.Piece(chess.KING, chess.WHITE)
        value = evalPiece(chess.E1, piece, False)
        self.assertIsInstance(value, float)

    def test_evaluate(self):
        value = evaluate(self.board)
        self.assertIsInstance(value, float)

    def test_scan(self):
        bitboard = self.board.occupied
        squares = list(scan(bitboard))
        self.assertIsInstance(squares, list)
        self.assertTrue(all(isinstance(sq, int) for sq in squares))

if __name__ == '__main__':
    unittest.main()
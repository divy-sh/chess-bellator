import unittest
from unittest.mock import patch, MagicMock
import chess
from engine import genMoveIterative

class TestGenMoveIterative(unittest.TestCase):
    def setUp(self):
        self.board = chess.Board()

    @patch('engine.genMove')
    @patch('engine.time.time')
    def test_genMoveIterative_timeout(self, mock_time, mock_genMove):
        mock_time.side_effect = [0, 1, 2, 3, 4, 5, 6]
        mock_genMove.side_effect = ["e2e4", "e7e5", "g1f3", "b8c6", "f1b5"]
        move = genMoveIterative(6, self.board)
        self.assertEqual(move, "f1b5")
        self.assertEqual(mock_genMove.call_count, 5)

    @patch('engine.genMove')
    @patch('engine.time.time')
    def test_genMoveIterative_no_moves(self, mock_time, mock_genMove):
        mock_time.side_effect = [0, 1, 2, 3, 4, 5, 6]
        mock_genMove.side_effect = [None, None, None, None, None]
        move = genMoveIterative(6, self.board)
        self.assertIsNone(move)
        self.assertEqual(mock_genMove.call_count, 5)

    @patch('engine.genMove')
    @patch('engine.time.time')
    def test_genMoveIterative_depth_increase(self, mock_time, mock_genMove):
        mock_time.side_effect = [0, 1, 2, 3, 4, 5, 6]
        mock_genMove.side_effect = ["e2e4", "e7e5", "g1f3", "b8c6", "f1b5"]
        move = genMoveIterative(6, self.board)
        self.assertEqual(move, "f1b5")
        self.assertEqual(mock_genMove.call_count, 5)

if __name__ == '__main__':
    unittest.main()
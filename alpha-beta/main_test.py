import unittest
from unittest.mock import MagicMock

# FILE: alpha-beta/test_main.py
from main import AIvsAI

class TestAIvsAI(unittest.TestCase):
    def setUp(self):
        self.game = MagicMock()
        self.engine = MagicMock()
        self.view = MagicMock()

    def test_AIvsAI_game_over(self):
        self.game.gameOver.return_value = True
        self.game.outcome.return_value = "1-0"
        AIvsAI(self.game, self.engine, self.view)
        self.game.gameOver.assert_called()
        self.game.outcome.assert_called()
        self.view.update.assert_not_called()

    def test_AIvsAI_play_move(self):
        self.game.gameOver.side_effect = [False, True]
        self.engine.genMoveIterative.return_value = (MagicMock(uci=lambda: "e2e4"), 0.5)
        self.game.getBoard.return_value.san.return_value = "e2e4"
        AIvsAI(self.game, self.engine, self.view)
        self.engine.genMoveIterative.assert_called()
        self.game.playMove.assert_called_with("e2e4")
        self.view.update.assert_called()

if __name__ == '__main__':
    unittest.main()
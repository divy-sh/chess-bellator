import chess.svg

from PyQt5.QtSvg import QSvgWidget
from PyQt5.QtWidgets import QApplication, QPushButton, QGridLayout

class View:
    def __init__(self, board) -> None:
        self.app = QApplication([])
        self.board = board
        self.svgWidget = QSvgWidget()
        self.svgWidget.renderer().load(bytearray(chess.svg.board(board=self.board, size=500), encoding='utf-8'))
        self.svgWidget.setFixedSize(500, 500)
        self.svgWidget.show()
    
    def update(self) -> None:
        self.svg = chess.svg.board(board=self.board, size=500)
        self.svgWidget.renderer().load(bytearray(self.svg, encoding='utf-8'))
        self.svgWidget.update()
        self.app.processEvents()
import chess
import chess.engine
import time

class Engine:
    def __init__(self, board) -> None:
        self.transpositionTable = {}
        self.transpositionTableHit = 0
        self.board = board
        self.pieceValue = {'rook': 500, 'pawn' : 100, 'bishop': 300, 'knight' : 300, 'queen': 900, 'king': 10000}

        self.pawn_table = [
            0,  0,  0,  0,  0,  0,  0,  0,
            50, 50, 50, 50, 50, 50, 50, 50,
            10, 10, 20, 30, 30, 20, 10, 10,
            5,  5, 10, 25, 25, 10,  5,  5,
            0,  0,  0, 20, 20,  0,  0,  0,
            5, -5,-10,  0,  0,-10, -5,  5,
            5, 10, 10,-20,-20, 10, 10,  5,
            0,  0,  0,  0,  0,  0,  0,  0
        ]
        self.rev_pawn_table = list(reversed(self.pawn_table))

        self.knight_table = [
            -50,-40,-30,-30,-30,-30,-40,-50,
            -40,-20,  0,  0,  0,  0,-20,-40,
            -30,  0, 10, 15, 15, 10,  0,-30,
            -30,  5, 15, 20, 20, 15,  5,-30,
            -30,  0, 15, 20, 20, 15,  0,-30,
            -30,  5, 10, 15, 15, 10,  5,-30,
            -40,-20,  0,  5,  5,  0,-20,-40,
            -50,-40,-30,-30,-30,-30,-40,-50
        ]
        self.rev_knight_table = list(reversed(self.knight_table))

        self.bishop_table = [
            -20,-10,-10,-10,-10,-10,-10,-20,
            -10,  0,  0,  0,  0,  0,  0,-10,
            -10,  0,  5, 10, 10,  5,  0,-10,
            -10,  5,  5, 10, 10,  5,  5,-10,
            -10,  0, 10, 10, 10, 10,  0,-10,
            -10, 10, 10, 10, 10, 10, 10,-10,
            -10,  5,  0,  0,  0,  0,  5,-10,
            -20,-10,-10,-10,-10,-10,-10,-20
        ]
        self.rev_bishop_table = list(reversed(self.bishop_table))

        self.rook_table = [
            0,  0,  0,  0,  0,  0,  0,  0,
            5, 10, 10, 10, 10, 10, 10,  5,
            -5,  0,  0,  0,  0,  0,  0, -5,
            -5,  0,  0,  0,  0,  0,  0, -5,
            -5,  0,  0,  0,  0,  0,  0, -5,
            -5,  0,  0,  0,  0,  0,  0, -5,
            -5,  0,  0,  0,  0,  0,  0, -5,
            0,  0,  0,  5,  5,  0,  0,  0
        ]
        self.rev_rook_table = list(reversed(self.rook_table))

        self.queen_table = [
            -20,-10,-10, -5, -5,-10,-10,-20,
            -10,  0,  0,  0,  0,  0,  0,-10,
            -10,  0,  5,  5,  5,  5,  0,-10,
            -5,  0,  5,  5,  5,  5,  0, -5,
            0,  0,  5,  5,  5,  5,  0, -5,
            -10,  5,  5,  5,  5,  5,  0,-10,
            -10,  0,  5,  0,  0,  0,  0,-10,
            -20,-10,-10, -5, -5,-10,-10,-20
        ]
        self.rev_queen_table = list(reversed(self.queen_table))

        self.king_middle_game_table = [
            -30,-40,-40,-50,-50,-40,-40,-30,
            -30,-40,-40,-50,-50,-40,-40,-30,
            -30,-40,-40,-50,-50,-40,-40,-30,
            -30,-40,-40,-50,-50,-40,-40,-30,
            -20,-30,-30,-40,-40,-30,-30,-20,
            -10,-20,-20,-20,-20,-20,-20,-10,
            20, 20,  0,  0,  0,  0, 20, 20,
            20, 30, 10,  0,  0, 10, 30, 20
        ]
        self.rev_king_middle_game_table = list(reversed(self.king_middle_game_table))

        self.king_end_game_table = [
            -50,-40,-30,-20,-20,-30,-40,-50,
            -30,-20,-10,  0,  0,-10,-20,-30,
            -30,-10, 20, 30, 30, 20,-10,-30,
            -30,-10, 30, 40, 40, 30,-10,-30,
            -30,-10, 30, 40, 40, 30,-10,-30,
            -30,-10, 20, 30, 30, 20,-10,-30,
            -30,-30,  0,  0,  0,  0,-30,-30,
            -50,-30,-30,-30,-30,-30,-30,-50
        ]
        self.rev_king_end_game_table = list(reversed(self.king_end_game_table))

    # generate move with fixed depth alpha-beta search
    def genMove(self, depth):
        move, _ = self.alphaBeta(depth, float('-inf'), float('inf'), self.board.turn)
        return move.uci()
    
    #iterative deepening until time runs out
    def genMoveIterative(self, seconds):    
        timeout = time.time() + seconds
        depth = 2
        move = None
        while time.time() < timeout:
            depth += 1
            move, _ = self.alphaBeta(depth, float('-inf'), float('inf'), self.board.turn)

        print(f"depth reached - {depth}")
        return move.uci()

    def alphaBeta(self, depth, alpha, beta, isMax):
        position_key = self.board.fen() + str(depth) + str(alpha) + str(beta)
        if position_key in self.transpositionTable:
            return self.transpositionTable[position_key]
        
        if depth == 0:
            return None, self.evaluate(isMax)
        
        if isMax:
            bestValue = float('-inf')
        else:
            bestValue = float('inf')
        bestMove = None
        moves = sorted(list(self.board.legal_moves), 
                       key=lambda move: (self.board.piece_at(move.to_square) 
                        is None, move.from_square, move.to_square))
        for move in moves:
            self.board.push_uci(move.uci())
            _, value = self.alphaBeta(depth - 1, alpha, beta, not isMax)
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
                self.transpositionTable[position_key] = (bestMove, bestValue)
                return bestMove, bestValue
        self.transpositionTable[position_key] = (bestMove, bestValue)
        return bestMove, bestValue

    def evaluate(self, isMax):
        value = 0
        board = self.board
        material_value = 0
        positional_value = 0
        for square, piece in board.piece_map().items():
            if piece.color == chess.WHITE:
                material_value += self.pieceValue[chess.piece_name(piece.piece_type)]
                positional_value += self.get_piece_square_table(piece.piece_type, square, chess.WHITE)
            else:
                material_value -= self.pieceValue[chess.piece_name(piece.piece_type)]
                positional_value -= self.get_piece_square_table(piece.piece_type, square, chess.BLACK)
        
        mobility_value = 0
        if isMax:
            mobility_value -= len(list(board.legal_moves))
        else:
            mobility_value += len(list(board.legal_moves))
        
        center_squares = [chess.E4, chess.D4, chess.E5, chess.D5]
        center_control_value = 0
        for square in center_squares:
            piece = board.piece_at(square)
            if piece is not None:
                if piece.color == chess.WHITE:
                    center_control_value += self.pieceValue[chess.piece_name(piece.piece_type)]
                else:
                    center_control_value -= self.pieceValue[chess.piece_name(piece.piece_type)]
        
        value = material_value + positional_value + mobility_value + center_control_value
        game_phase = self.get_game_phase()
        value = value * game_phase + (1 - game_phase) * material_value
        return value

    def get_piece_square_table(self, piece_type, square, color):
        # Define piece-square tables for both colors
        # Flip the table based on the color
        if color == chess.BLACK:
            pawn_table = self.rev_pawn_table
            knight_table = self.rev_knight_table
            bishop_table = self.rev_bishop_table
            rook_table = self.rev_rook_table
            queen_table = self.rev_queen_table
            king_middle_game_table = self.rev_king_middle_game_table
            king_end_game_table = self.rev_king_end_game_table
        else:
            pawn_table = self.pawn_table
            knight_table = self.knight_table
            bishop_table = self.bishop_table
            rook_table = self.rook_table
            queen_table = self.queen_table
            king_middle_game_table = self.king_middle_game_table
            king_end_game_table = self.king_end_game_table

        # Return the corresponding table for the given piece type
        if piece_type == chess.PAWN:
            return pawn_table[square]
        elif piece_type == chess.KNIGHT:
            return knight_table[square]
        elif piece_type == chess.BISHOP:
            return bishop_table[square]
        elif piece_type == chess.ROOK:
            return rook_table[square]
        elif piece_type == chess.QUEEN:
            return queen_table[square]
        elif piece_type == chess.KING:
            # Use a weighted average of the middle game and end game tables
            game_phase = self.get_game_phase()
            return game_phase * king_middle_game_table[square] + (1 - game_phase) * king_end_game_table[square]

    def get_game_phase(self):
        total_pieces = sum(1 for _ in self.board.occupied_co)
        max_pieces = 32  # Maximum number of pieces at the start of the game
        game_phase = total_pieces / max_pieces
        return game_phase
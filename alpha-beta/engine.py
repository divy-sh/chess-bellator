import chess
import chess.engine
import time
import evalConstants

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
            if lastMoveSpecial:
                return self.qSearch(alpha, beta, not isMax, lastMoveSpecial)
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
            return None, self.evaluate(isMax)
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

    def evaluate(self, isMax: bool) -> float:
        value = 0
        board = self.board
        material_value = 0
        positional_value = 0
        for square, piece in board.piece_map().items():
            if piece.color == chess.WHITE:
                material_value += evalConstants.pieceValue[chess.piece_name(piece.piece_type)]
                positional_value += self.get_piece_square_table(piece.piece_type, square, chess.WHITE)
            else:
                material_value -= evalConstants.pieceValue[chess.piece_name(piece.piece_type)]
                positional_value -= self.get_piece_square_table(piece.piece_type, square, chess.BLACK)
        
        mobility_value = 0
        for square in chess.SQUARES:
            if board.is_attacked_by(chess.WHITE, square):
                mobility_value += 50

            if board.is_attacked_by(chess.BLACK, square):
                mobility_value -= 50
        
        center_squares = [chess.E4, chess.D4, chess.E5, chess.D5]
        center_control_value = 0
        for square in center_squares:
            piece = board.piece_at(square)
            if piece is not None:
                if piece.color == chess.WHITE:
                    center_control_value += 50
                else:
                    center_control_value -= 50
        
        value = material_value + positional_value + mobility_value + center_control_value
        game_phase = self.get_game_phase()
        value = value * game_phase + (1 - game_phase) * material_value
        return value

    def get_piece_square_table(self, piece_type: chess.PieceType, square: chess.Square, color: chess.Color) -> float:
        # Define piece-square tables for both colors
        # Flip the table based on the color
        if color == chess.BLACK:
            pawn_table = evalConstants.rev_pawn_table
            knight_table = evalConstants.rev_knight_table
            bishop_table = evalConstants.rev_bishop_table
            rook_table = evalConstants.rev_rook_table
            queen_table = evalConstants.rev_queen_table
            king_middle_game_table = evalConstants.rev_king_middle_game_table
            king_end_game_table = evalConstants.rev_king_end_game_table
        else:
            pawn_table = evalConstants.pawn_table
            knight_table = evalConstants.knight_table
            bishop_table = evalConstants.bishop_table
            rook_table = evalConstants.rook_table
            queen_table = evalConstants.queen_table
            king_middle_game_table = evalConstants.king_middle_game_table
            king_end_game_table = evalConstants.king_end_game_table

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
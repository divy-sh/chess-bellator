import chess

pieceValue = {'rook': 500, 'pawn' : 100, 'bishop': 300, 'knight' : 300, 'queen': 900, 'king': 10000}

pawn_table = [
    0,  0,  0,  0,  0,  0,  0,  0,
    50, 50, 50, 50, 50, 50, 50, 50,
    10, 10, 20, 30, 30, 20, 10, 10,
    5,  5, 10, 25, 25, 10,  5,  5,
    0,  0,  0, 20, 20,  0,  0,  0,
    5, -5,-10,  0,  0,-10, -5,  5,
    5, 10, 10,-20,-20, 10, 10,  5,
    0,  0,  0,  0,  0,  0,  0,  0
]
rev_pawn_table = list(reversed(pawn_table))

knight_table = [
    -50,-40,-30,-30,-30,-30,-40,-50,
    -40,-20,  0,  0,  0,  0,-20,-40,
    -30,  0, 10, 15, 15, 10,  0,-30,
    -30,  5, 15, 20, 20, 15,  5,-30,
    -30,  0, 15, 20, 20, 15,  0,-30,
    -30,  5, 10, 15, 15, 10,  5,-30,
    -40,-20,  0,  5,  5,  0,-20,-40,
    -50,-40,-30,-30,-30,-30,-40,-50
]
rev_knight_table = list(reversed(knight_table))

bishop_table = [
    -20,-10,-10,-10,-10,-10,-10,-20,
    -10,  0,  0,  0,  0,  0,  0,-10,
    -10,  0,  5, 10, 10,  5,  0,-10,
    -10,  5,  5, 10, 10,  5,  5,-10,
    -10,  0, 10, 10, 10, 10,  0,-10,
    -10, 10, 10, 10, 10, 10, 10,-10,
    -10,  5,  0,  0,  0,  0,  5,-10,
    -20,-10,-10,-10,-10,-10,-10,-20
]
rev_bishop_table = list(reversed(bishop_table))

rook_table = [
    0,  0,  0,  0,  0,  0,  0,  0,
    5, 10, 10, 10, 10, 10, 10,  5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    0,  0,  0,  5,  5,  0,  0,  0
]
rev_rook_table = list(reversed(rook_table))

queen_table = [
    -20,-10,-10, -5, -5,-10,-10,-20,
    -10,  0,  0,  0,  0,  0,  0,-10,
    -10,  0,  5,  5,  5,  5,  0,-10,
    -5,  0,  5,  5,  5,  5,  0, -5,
    0,  0,  5,  5,  5,  5,  0, -5,
    -10,  5,  5,  5,  5,  5,  0,-10,
    -10,  0,  5,  0,  0,  0,  0,-10,
    -20,-10,-10, -5, -5,-10,-10,-20
]
rev_queen_table = list(reversed(queen_table))

king_middle_game_table = [
    -30,-40,-40,-50,-50,-40,-40,-30,
    -30,-40,-40,-50,-50,-40,-40,-30,
    -30,-40,-40,-50,-50,-40,-40,-30,
    -30,-40,-40,-50,-50,-40,-40,-30,
    -20,-30,-30,-40,-40,-30,-30,-20,
    -10,-20,-20,-20,-20,-20,-20,-10,
    20, 20,  0,  0,  0,  0, 20, 20,
    20, 30, 10,  0,  0, 10, 30, 20
]
rev_king_middle_game_table = list(reversed(king_middle_game_table))

king_end_game_table = [
    -50,-40,-30,-20,-20,-30,-40,-50,
    -30,-20,-10,  0,  0,-10,-20,-30,
    -30,-10, 20, 30, 30, 20,-10,-30,
    -30,-10, 30, 40, 40, 30,-10,-30,
    -30,-10, 30, 40, 40, 30,-10,-30,
    -30,-10, 20, 30, 30, 20,-10,-30,
    -30,-30,  0,  0,  0,  0,-30,-30,
    -50,-30,-30,-30,-30,-30,-30,-50
]
rev_king_end_game_table = list(reversed(king_end_game_table))

def evaluate(board: chess.Board, isMax: bool) -> float:
        value = 0
        material_value = 0
        positional_value = 0
        for square, piece in board.piece_map().items():
            if piece.color == chess.WHITE:
                material_value += pieceValue[chess.piece_name(piece.piece_type)]
                positional_value += get_piece_square_table(board, piece.piece_type, square, chess.WHITE)
            else:
                material_value -= pieceValue[chess.piece_name(piece.piece_type)]
                positional_value -= get_piece_square_table(board, piece.piece_type, square, chess.BLACK)
        
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
        game_phase = get_game_phase(board)
        value = value * game_phase + (1 - game_phase) * material_value
        return value

def get_piece_square_table(board: chess.Board, piece_type: chess.PieceType, square: chess.Square, color: chess.Color) -> float:
    # Define piece-square tables for both colors
    # Flip the table based on the color
    if color == chess.BLACK:
        pt = rev_pawn_table
        kt = rev_knight_table
        bt = rev_bishop_table
        rt = rev_rook_table
        qt = rev_queen_table
        kmt = rev_king_middle_game_table
        ket = rev_king_end_game_table
    else:
        pt = pawn_table
        kt = knight_table
        bt = bishop_table
        rt = rook_table
        qt = queen_table
        kmt = king_middle_game_table
        ket = king_end_game_table

    # Return the corresponding table for the given piece type
    if piece_type == chess.PAWN:
        return pt[square]
    elif piece_type == chess.KNIGHT:
        return kt[square]
    elif piece_type == chess.BISHOP:
        return bt[square]
    elif piece_type == chess.ROOK:
        return rt[square]
    elif piece_type == chess.QUEEN:
        return qt[square]
    elif piece_type == chess.KING:
        # Use a weighted average of the middle game and end game tables
        game_phase = get_game_phase(board)
        return game_phase * kmt[square] + (1 - game_phase) * ket[square]

def get_game_phase(board: chess.Board):
    total_pieces = sum(1 for _ in board.occupied_co)
    max_pieces = 32  # Maximum number of pieces at the start of the game
    game_phase = total_pieces / max_pieces
    return game_phase
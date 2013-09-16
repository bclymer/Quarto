package realtime

import (
	"log"
)

type Game struct {
	GameState       int
	UsedPieces      []int
	AvailablePieces []int
	Board           []int
	SelectedPiece   int
}

type Piece struct {
	Square bool
	Hole   bool
	White  bool
	Tall   bool
}

const (
	GameStateNoPlayers         = 0
	GameStatePlayerOneChoosing = 1
	GameStatePlayerOnePlaying  = 2
	GameStatePlayerTwoChoosing = 3
	GameStatePlayerTwoPlaying  = 4
)

var (
	pieces = []Piece{Piece{Square: true, Hole: true, White: true, Tall: true},
		Piece{Square: true, Hole: true, White: true, Tall: false},
		Piece{Square: true, Hole: true, White: false, Tall: true},
		Piece{Square: true, Hole: true, White: false, Tall: false},
		Piece{Square: true, Hole: false, White: true, Tall: true},
		Piece{Square: true, Hole: false, White: true, Tall: false},
		Piece{Square: true, Hole: false, White: false, Tall: true},
		Piece{Square: true, Hole: false, White: false, Tall: false},
		Piece{Square: false, Hole: true, White: true, Tall: true},
		Piece{Square: false, Hole: true, White: true, Tall: false},
		Piece{Square: false, Hole: true, White: false, Tall: true},
		Piece{Square: false, Hole: true, White: false, Tall: false},
		Piece{Square: false, Hole: false, White: true, Tall: true},
		Piece{Square: false, Hole: false, White: true, Tall: false},
		Piece{Square: false, Hole: false, White: false, Tall: true},
		Piece{Square: false, Hole: false, White: false, Tall: false}}
)

func MakeNewGame() *Game {
	pieces := make([]int, 16, 16)
	board := make([]int, 16, 16)
	for i := 0; i < 16; i++ {
		pieces[i] = i
		board[i] = -1
	}
	game := Game{
		GameState:       GameStateNoPlayers,
		UsedPieces:      board,
		AvailablePieces: pieces,
		Board:           board,
		SelectedPiece:   -1}
	return &game
}

func (game *Game) Reset() {
	pieces := make([]int, 16, 16)
	board := make([]int, 16, 16)
	for i := 0; i < 16; i++ {
		pieces[i] = i
		board[i] = -1
	}
	game.GameState = 0
	game.UsedPieces = board
	game.AvailablePieces = pieces
	game.Board = board
	game.SelectedPiece = -1
}

func whoIsWinning(game *Game) int {
	if game.GameState == GameStatePlayerOneChoosing {
		return 1
	} else if game.GameState == GameStatePlayerTwoChoosing {
		return 2
	} else {
		log.Fatal("Shouldn't check winner with this state")
		return 0
	}
}

func (game *Game) CheckWinner() int {
	if checkSequence(func(i, j int) int {
		return i*4 + j
	}, game) {
		return whoIsWinning(game)
	}
	if checkSequence(func(i, j int) int {
		return j*4 + i
	}, game) {
		return whoIsWinning(game)
	}
	{ // limit scope
		square, hole, white, tall := 10, 10, 10, 10
		for j := 0; j < 16; j += 5 {
			pieceId := game.Board[j]
			if pieceId == -1 {
				square = 0
				hole = 0
				white = 0
				tall = 0
				break
			}
			piece := pieces[pieceId]
			square += intForBool(piece.Square)
			hole += intForBool(piece.Hole)
			white += intForBool(piece.White)
			tall += intForBool(piece.Tall)
		}
		if checkValues(square, hole, white, tall) {
			return whoIsWinning(game)
		}
	}

	{ // limit scope
		square, hole, white, tall := 10, 10, 10, 10
		for j := 0; j < 5; j++ {
			pieceId := game.Board[j]
			if pieceId == -1 {
				square = 0
				hole = 0
				white = 0
				tall = 0
				break
			}
			piece := pieces[pieceId]
			square += intForBool(piece.Square)
			hole += intForBool(piece.Hole)
			white += intForBool(piece.White)
			tall += intForBool(piece.Tall)
		}
		if checkValues(square, hole, white, tall) {
			return whoIsWinning(game)
		}
	}
	return 0
}

type locationFunction func(int, int) int

func checkSequence(locationFunction locationFunction, game *Game) bool {
	for i := 0; i < 4; i++ {
		square, hole, white, tall := 10, 10, 10, 10
		for j := 0; j < 4; j++ {
			pieceId := game.Board[locationFunction(i, j)]
			if pieceId == -1 {
				square = 0
				hole = 0
				white = 0
				tall = 0
				break
			}
			piece := pieces[pieceId]
			square += intForBool(piece.Square)
			hole += intForBool(piece.Hole)
			white += intForBool(piece.White)
			tall += intForBool(piece.Tall)
		}
		if checkValues(square, hole, white, tall) {
			return true
		}
	}
	return false
}

func checkValues(square, hole, white, tall int) bool {
	if isSameKind(square) || isSameKind(hole) || isSameKind(white) || isSameKind(tall) {
		return true
	}
	return false
}

func isSameKind(num int) bool {
	return num == 10 || num == 14
}

func (room *Room) UpdateGame() {
	if room.Game.GameState == GameStateNoPlayers && room.PlayerOne != nil && room.PlayerTwo != nil {
		room.Game.Reset()
		room.Game.GameState = GameStatePlayerOneChoosing
	} else if room.PlayerOne == nil || room.PlayerTwo == nil {
		room.Game.GameState = GameStateNoPlayers
	}
}

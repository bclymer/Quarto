package realtime

import ()

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

func (game *Game) CheckWinner() int {

	return 0
}

func (room *Room) UpdateGame() {
	if room.Game.GameState == GameStateNoPlayers && room.PlayerOne != nil && room.PlayerTwo != nil {
		room.Game.GameState = GameStatePlayerOneChoosing
	} else if room.PlayerOne == nil || room.PlayerTwo == nil {
		room.Game.GameState = GameStateNoPlayers
	}
}

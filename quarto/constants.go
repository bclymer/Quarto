package quarto

import (
	"log"
)

var (
	Config ConfigContainer
)

func InitConstants() {
	log.Println("Initing Constant Values")
	Config = ConfigContainer{
		"com.bclymer.quarto.user.add",
		"com.bclymer.quarto.user.remove",
		"com.bclymer.quarto.user.challenge",
		"com.bclymer.quarto.user.room.join",
		"com.bclymer.quarto.user.room.leave",
		"com.bclymer.quarto.room.add",
		"com.bclymer.quarto.room.remove",
		"com.bclymer.quarto.room.name.change",
		"com.bclymer.quarto.room.privacy.change",
		"com.bclymer.quarto.room.change",
		"com.bclymer.quarto.chat",
		"com.bclymer.quarto.game.playerOne.request",
		"com.bclymer.quarto.game.playerOne.leave",
		"com.bclymer.quarto.game.playerTwo.request",
		"com.bclymer.quarto.game.playerTwo.leave",
		"com.bclymer.quarto.game.change",
		"com.bclymer.quarto.game.piece.played",
		"com.bclymer.quarto.game.piece.chosen",
		"com.bclymer.quarto.game.winner",
		"com.bclymer.quarto.info",
		"com.bclymer.quarto.error"}
}

type ConfigContainer struct {
	UserAdd              string
	UserRemove           string
	UserChallenge        string
	UserRoomJoin         string
	UserRoomLeave        string
	RoomAdd              string
	RoomRemove           string
	RoomNameChange       string
	RoomPrivacyChange    string
	RoomChange           string
	Chat                 string
	GamePlayerOneRequest string
	GamePlayerOneLeave   string
	GamePlayerTwoRequest string
	GamePlayerTwoLeave   string
	GameChange           string
	GamePiecePlayed      string
	GamePieceChosen      string
	GameWinner           string
	Info                 string
	Error                string
}

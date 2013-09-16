package realtime

import (
	"container/list"
	"encoding/json"
)

type Room struct {
	PlayerOne *User
	PlayerTwo *User
	Observers *list.List
	Events    *list.List // history of events
	Name      string
	Private   bool
	Password  string // only used if the room is private
	Game      *Game
}

type User struct {
	Username string            // selected username
	Room     *Room             // room the user is in.
	Events   chan *ClientEvent // event channel to send messages to user
}

type AddUserDTO struct {
	Username string
}

type RemoveUserDTO struct {
	Username string
}

type UserChallengedUserDTO struct {
	Challenger string
}

type JoinRoomDTO struct {
	Name string
}

type LeaveRoomDTO struct {
	Name string
}

type UserRoomDTO struct {
	Username string
	RoomName string
}

type AddRoomDTO struct {
	Name     string
	Private  bool
	Password string
}

type RemoveRoomDTO struct {
	Name string
}

type RoomPrivacyChangedDTO struct {
	Private  bool
	Password string
}

type RoomNameChangedDTO struct {
	Name string
}

type RoomRoomDTO struct { // The room DTO that gets sent to users inside the room
	Name      string
	Private   bool
	PlayerOne string
	PlayerTwo string
	Observers []string
}

func MakeRoomRoomDTO(room *Room) RoomRoomDTO {
	observers := make([]string, room.Observers.Len())
	i := 0
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		observers[i] = observer.Value.(*User).Username
		i++
	}
	playerOneName := ""
	if room.PlayerOne != nil {
		playerOneName = room.PlayerOne.Username
	}
	playerTwoName := ""
	if room.PlayerTwo != nil {
		playerTwoName = room.PlayerTwo.Username
	}
	return RoomRoomDTO{room.Name, room.Private, playerOneName, playerTwoName, observers}
}

type LobbyRoomDTO struct { // The room DTO that users in the lobby get
	Name    string
	Private bool
	Members int
}

func MakeLobbyRoomDTO(room *Room) LobbyRoomDTO {
	return LobbyRoomDTO{room.Name, room.Private, room.Observers.Len()}
}

type LobbyUserDTO struct {
	Username string
	RoomName string
}

type ClientEvent struct {
	Action string
	Data   string // json encoding of some data
}

type IncomingChatDTO struct {
	Message string
}

type OutgoingChatDTO struct {
	Message  string
	Username string
}

type GamePiecePlayedDTO struct {
	Location int
}

type GamePieceChosenDTO struct {
	Piece int
}

type InfoOrErrorMessageDTO struct {
	Message string
}

type GameWinnerDTO struct {
	Winner string
}

func DtoToString(thing interface{}) (string, error) {
	thingByteArray, err := json.Marshal(thing)
	if err != nil {
		return "", err
	}
	return string(thingByteArray), err
}

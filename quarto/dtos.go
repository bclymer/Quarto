package quarto

import (
	"container/list"
	"encoding/json"
)

type Room struct {
	PlayerOne *User      `json:"-"`
	PlayerTwo *User      `json:"-"`
	Observers *list.List `json:"-"`
	Events    *list.List `json:"-"` // history of events
	Name      string     `json:"name"`
	Private   bool       `json:"private"`
	Password  string     `json:"-"` // only used if the room is private
	Game      *Game      `json:"-"`
}

type User struct {
	Username string            `json:"username"` // selected username
	Room     *Room             `json:"room"`     // room the user is in.
	Events   chan *ClientEvent `json:"-"`        // event channel to send messages to user
	Active   bool              `json:"-"`
}

type AddUserDTO struct {
	Username string `json:"username"`
}

type RemoveUserDTO struct {
	Username string `json:"username"`
}

type UserChallengedUserDTO struct {
	Challenger string `json:"challenger"`
}

type JoinRoomDTO struct {
	Name string `json:"name"`
}

type LeaveRoomDTO struct {
	Name string `json:"name"`
}

type UserRoomDTO struct {
	Username string `json:"username"`
	RoomName string `json:"roomname"`
}

type AddRoomDTO struct {
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Password string `json:"password"`
}

type RemoveRoomDTO struct {
	Name string `json:"name"`
}

type RoomRoomDTO struct { // The room DTO that gets sent to users inside the room
	Name      string   `json:"name"`
	Private   bool     `json:"private"`
	PlayerOne string   `json:"playerOne"`
	PlayerTwo string   `json:"playerTwo"`
	Observers []string `json:"observers"`
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
	Name    string `json:"name"`
	Private bool   `json:"private"`
	Members int    `json:"members"`
}

func MakeLobbyRoomDTO(room *Room) LobbyRoomDTO {
	return LobbyRoomDTO{room.Name, room.Private, room.Observers.Len()}
}

type LobbyUserDTO struct {
	Username string `json:"username"`
	RoomName string `json:"roomName"`
}

type ClientEvent struct {
	Action string `json:"action"`
	Data   string `json:"data"` // json encoding of some data
}

type IncomingChatDTO struct {
	Message string `json:"message"`
}

type OutgoingChatDTO struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

type GamePiecePlayedDTO struct {
	Location int `json:"location"`
}

type GamePieceChosenDTO struct {
	Piece int `json:"piece"`
}

type InfoOrErrorMessageDTO struct {
	Message string `json:"message"`
}

type GameWinnerDTO struct {
	Winner string `json:"winner"`
}

func DtoToString(thing interface{}) (string, error) {
	thingByteArray, err := json.Marshal(thing)
	if err != nil {
		return "", err
	}
	return string(thingByteArray), err
}

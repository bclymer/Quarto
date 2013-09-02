package realtime

import (
	"container/list"
)

type Room struct {
	PlayerOne	*User
	PlayerTwo	*User
	Observers	*list.List
	Events		*list.List // history of events
	Name		string
	Private		bool
	Password	string // only used if the room is private
}

type User struct {
	Username	string // selected username
	Room		*Room // room the user is in.
	Events		chan ClientEvent // event channel to send messages to user
}

type AddUserDTO struct {
	Username	string
}

type RemoveUserDTO struct {
	Username	string
}

type UserChallengedUserDTO struct {
	Challenger	string
}

type JoinRoomDTO struct {
	Name		string
}

type LeaveRoomDTO struct {
	Name		string	
}

type UserRoomDTO struct {
	Username	string
	RoomName	string
}

type AddRoomDTO struct {
	Name		string
}

type RemoveRoomDTO struct {
	Name		string
}

type RoomPrivacyChangedDTO struct {
	Private		bool
	Password	string
}

type RoomNameChangedDTO struct {
	Name		string
}

type RoomPlayerOneChangedDTO struct {
	Username	string
}

type RoomPlayerTwoChangedDTO struct {
	Username	string	
}

type RoomObserversChangedDTO struct {
	Username	string
}

type RoomRoomDTO struct { // The room DTO that gets sent to users inside the room
	Name		string
	Private		string
	PlayerOne	string
	PlayerTwo	string
	Observers	[]string
}

type LobbyRoomDTO struct { // The room DTO that users in the lobby get
	Name		string
	Private		bool
	Members		int
}

type LobbyUserDTO struct {
	Username	string
	RoomName	string
}

type ClientEvent struct {
	Action		string
	Data		string // json encoding of some data
}
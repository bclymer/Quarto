package quarto

import (
	"encoding/json"
	"log"
	"strings"
)

type NewUserInfoAndChannel struct {
	User     chan *User
	Username string
}

type CheckUsername struct {
	Username string
	Valid    chan bool
}

type RecievedEvent struct {
	clientEvent ClientEvent
	Username    string
}

// Owner must cancel itself when they stop listening to events.
func (user *User) Cancel() {
	log.Println("+realtime.Cancel")
	removeUserDTO, err := json.Marshal(RemoveUserDTO{user.Username})
	if err != nil {
		log.Fatal("Cancel: Couldn't marshal the thing", err)
	}
	clientEvent := ClientEvent{Config.UserRemove, string(removeUserDTO)}
	recievedEvent := RecievedEvent{clientEvent, user.Username}
	user.Active = false
	recievedEventChannel <- &recievedEvent
	log.Println("-realtime.Cancel")
}

func Subscribe(username string) *User {
	log.Println("+realtime.Subscribe")
	newUserInfoAndChannel := NewUserInfoAndChannel{make(chan *User), username}
	subscribeChannel <- &newUserInfoAndChannel
	user := <-newUserInfoAndChannel.User
	log.Println("-realtime.Subscribe")
	return user
}

func ServerSideAction(clientEvent ClientEvent, username string) {
	log.Println("+realtime.ServerSideAction", clientEvent)
	recievedEvent := RecievedEvent{clientEvent, username}
	recievedEventChannel <- &recievedEvent
	log.Println("-realtime.ServerSideAction")
}

func GetUserMap() *map[string]*User {
	userMapChannel := make(chan *map[string]*User)
	getUserMapChannel <- userMapChannel
	return <-userMapChannel
}

func GetRoomMap() *map[string]*Room {
	roomMapChannel := make(chan *map[string]*Room)
	getRoomMapChannel <- roomMapChannel
	return <-roomMapChannel
}

var (
	recievedEventChannel = make(chan *RecievedEvent)
	subscribeChannel     = make(chan *NewUserInfoAndChannel)
	checkUsernameChannel = make(chan *CheckUsername)
	getUserMapChannel    = make(chan (chan *map[string]*User))
	getRoomMapChannel    = make(chan (chan *map[string]*Room))
)

func ValidateUsername(name string) (bool, string) {
	log.Println("+realtime.ValidateUsername")
	if strings.TrimSpace(name) == "" {
		return false, ""
	}
	check := CheckUsername{name, make(chan bool)}
	checkUsernameChannel <- &check
	log.Println("-realtime.ValidateUsername")
	valid := <-check.Valid
	if valid {
		//mongoUser := InsertUser(NewMongoUser(name))
		return true, "" //mongoUser.Token
	} else {
		return false, ""
	}
}

// This function loops forever
func realtime() {

	userMap := make(map[string]*User)
	roomMap := make(map[string]*Room)

	setupEventProcessor(&userMap, &roomMap)

	for {
		select {
		case newUserInfoAndChannel := <-subscribeChannel:
			log.Println("+subscribeChannel")
			addUserDTO, err := json.Marshal(AddUserDTO{newUserInfoAndChannel.Username})
			if err != nil {
				log.Fatal("realtime: Couldn't marshal the thing")
				return
			}
			user := AddUser(string(addUserDTO))
			newUserInfoAndChannel.User <- user
			log.Println("-subscribeChannel")
		case recievedEvent := <-recievedEventChannel:
			log.Println("+recievedEventChannel", recievedEvent)
			clientEvent := recievedEvent.clientEvent
			username := recievedEvent.Username
			switch clientEvent.Action {
			case Config.UserRemove:
				RemoveUser(clientEvent.Data)
			case Config.UserChallenge:
				UserChallengedUser(clientEvent.Data, username)
			case Config.UserRoomJoin:
				JoinRoom(clientEvent.Data, username)
			case Config.UserRoomLeave:
				LeaveRoom(username)
			case Config.RoomAdd:
				AddRoom(clientEvent.Data, username)
			case Config.RoomRemove:
				RemoveRoom(clientEvent.Data)
			case Config.RoomNameChange:
				RoomNameChange(clientEvent.Data)
			case Config.RoomPrivacyChange:
				RoomNameChange(clientEvent.Data)
			case Config.Chat:
				Chat(clientEvent.Data, username)
			case Config.GamePlayerOneRequest:
				RequestPlayerOne(username)
			case Config.GamePlayerTwoRequest:
				RequestPlayerTwo(username)
			case Config.GamePlayerOneLeave:
				LeavePlayerOne(username)
			case Config.GamePlayerTwoLeave:
				LeavePlayerTwo(username)
			case Config.GamePiecePlayed:
				GamePiecePlayed(clientEvent.Data, username)
			case Config.GamePieceChosen:
				GamePieceChosen(clientEvent.Data, username)
			}
			log.Println("-recievedEventChannel")
		case check := <-checkUsernameChannel:
			log.Println("+checkUsernameChannel")
			check.Valid <- userMap[check.Username] == nil
			log.Println("-checkUsernameChannel")
		case userMapRequest := <-getUserMapChannel:
			log.Println("+getUserMapChannel")
			userMapRequest <- &userMap
			log.Println("-getUserMapChannel")
		case roomMapRequest := <-getRoomMapChannel:
			log.Println("+getRoomMapChannel")
			roomMapRequest <- &roomMap
			log.Println("-getRoomMapChannel")
		}
	}
}

func GetRoomUserCount(room *Room) int {
	log.Println("+realtime.GetRoomUserCount")
	members := 0
	if room.PlayerOne != nil {
		members++
	}
	if room.PlayerTwo != nil {
		members++
	}
	members += room.Observers.Len()
	log.Println("-realtime.GetRoomUserCount")
	return members
}

func init() {
	go realtime()
}
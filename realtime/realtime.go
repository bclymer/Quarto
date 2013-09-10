package realtime

import (
	"encoding/json"
	"log"
	"quarto/constants"
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
		log.Fatal("Cancel: Couldn't marshal the thing")
	}
	clientEvent := ClientEvent{constants.Config.UserRemove, string(removeUserDTO)}
	recievedEvent := RecievedEvent{clientEvent, user.Username}
	recievedEventChannel <- &recievedEvent
	log.Println("-realtime.Cancel")
}

func Subscribe(username string) *User {
	log.Println("+realtime.Subscribe")
	newUserInfoAndChannel := NewUserInfoAndChannel{make(chan *User), username}
	subscribeChannel <- &newUserInfoAndChannel
	log.Println("-realtime.Subscribe")
	return <-newUserInfoAndChannel.User
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

const archiveSize = 10

var (
	recievedEventChannel = make(chan *RecievedEvent, 10)
	subscribeChannel     = make(chan *NewUserInfoAndChannel, 10)
	checkUsernameChannel = make(chan *CheckUsername, 10)
	getUserMapChannel    = make(chan (chan *map[string]*User), 10)
	getRoomMapChannel    = make(chan (chan *map[string]*Room), 10)
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
		mongoUser := InsertUser(NewMongoUser(name))
		return true, mongoUser.Token
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
			case constants.Config.UserAdd:
				AddUser(clientEvent.Data)
			case constants.Config.UserRemove:
				RemoveUser(clientEvent.Data)
			case constants.Config.UserChallenge:
				UserChallengedUser(clientEvent.Data, username)
			case constants.Config.UserRoomJoin:
				JoinRoom(clientEvent.Data, username)
			case constants.Config.UserRoomLeave:
				LeaveRoom(username)
			case constants.Config.RoomAdd:
				AddRoom(clientEvent.Data, username)
			case constants.Config.RoomRemove:
				RemoveRoom(clientEvent.Data)
			case constants.Config.RoomNameChange:
				RoomNameChange(clientEvent.Data)
			case constants.Config.RoomPrivacyChange:
				RoomNameChange(clientEvent.Data)
			case constants.Config.Chat:
				Chat(clientEvent.Data, username)
			case constants.Config.GamePlayerOneRequest:
				RequestPlayerOne(username)
			case constants.Config.GamePlayerTwoRequest:
				RequestPlayerTwo(username)
			case constants.Config.GamePlayerOneLeave:
				LeavePlayerOne(username)
			case constants.Config.GamePlayerTwoLeave:
				LeavePlayerTwo(username)
			case constants.Config.GamePiecePlayed:
				GamePiecePlayed(clientEvent.Data, username)
			case constants.Config.GamePieceChosen:
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

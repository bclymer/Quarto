package realtime

import (
	"log"
	"encoding/json"
    "quarto/constants"
)

type NewUserInfoAndChannel struct {
	User		chan *User
	Username	string
}

type CheckUsername struct {
	Username	string
	Valid 		chan bool
}

type RecievedEvent struct {
	clientEvent	ClientEvent
	Username	string
}

// Owner must cancel itself when they stop listening to events.
func (user *User) Cancel() {
	log.Println("+realtime.Cancel")
	removeUserDTO, err := json.Marshal(RemoveUserDTO { user.Username })
	if (err != nil) {
		log.Fatal("Cancel: Couldn't marshal the thing")
	}
	clientEvent := ClientEvent { constants.RemoveUser, string(removeUserDTO) }
	recievedEvent := RecievedEvent { clientEvent, user.Username }
	recievedEventChannel <- &recievedEvent
	log.Println("-realtime.Cancel")
}

func Subscribe(username string) *User {
	log.Println("+realtime.Subscribe")
	newUserInfoAndChannel := NewUserInfoAndChannel { make(chan *User), username }
	subscribeChannel <- &newUserInfoAndChannel
	log.Println("-realtime.Subscribe")
	return <- newUserInfoAndChannel.User
}

func ServerSideAction(clientEvent ClientEvent, username string) {
	log.Println("+realtime.ServerSideAction", clientEvent)
	recievedEvent := RecievedEvent { clientEvent, username }
	recievedEventChannel <- &recievedEvent
	log.Println("-realtime.ServerSideAction")
}

func GetUserMap() *map[string]*User {
	userMapChannel := make(chan *map[string]*User)
	getUserMapChannel <- userMapChannel
	return <- userMapChannel
}

func GetRoomMap() *map[string]*Room {
	roomMapChannel := make(chan *map[string]*Room)
	getRoomMapChannel <- roomMapChannel
	return <- roomMapChannel
}

const archiveSize = 10

var (
	recievedEventChannel = make(chan *RecievedEvent)
	subscribeChannel = make(chan *NewUserInfoAndChannel)
	checkUsernameChannel = make(chan *CheckUsername)
	getUserMapChannel = make(chan (chan *map[string]*User))
	getRoomMapChannel = make(chan (chan *map[string]*Room))
)

func ValidateUsername(name string) bool {
	log.Println("+realtime.ValidateUsername")
	check := CheckUsername { name, make(chan bool) }
	checkUsernameChannel <- &check
	log.Println("-realtime.ValidateUsername")
	return <-check.Valid
}

// This function loops forever
func realtime() {

	userMap := make(map[string]*User)
	roomMap := make(map[string]*Room)

	for {
		select {
		case newUserInfoAndChannel := <- subscribeChannel:
			addUserDTO, err := json.Marshal(AddUserDTO { newUserInfoAndChannel.Username })
			if (err != nil) {
				log.Fatal("realtime: Couldn't marshal the thing")
			}
			user := AddUser(string(addUserDTO), userMap)
			newUserInfoAndChannel.User <- user
		case recievedEvent := <- recievedEventChannel:
			clientEvent := recievedEvent.clientEvent
			username := recievedEvent.Username
			switch (clientEvent.Action) {
			case constants.AddRoom:
				AddUser(clientEvent.Data, userMap)
			case constants.RemoveUser:
				RemoveUser(clientEvent.Data, userMap, roomMap)
			case constants.UserChallenge:
				UserChallengedUser(clientEvent.Data, username)
			case constants.JoinRoom:
				JoinRoom(clientEvent.Data, username, userMap, roomMap)
			case constants.LeaveRoom:
				LeaveRoom(clientEvent.Data, username, userMap, roomMap)
			case constants.AddRoom:
				AddRoom(clientEvent.Data, username)
			case constants.RemoveRoom:
				RemoveRoom(clientEvent.Data)
			case constants.ChangeRoomName:
				RoomNameChange(clientEvent.Data)
			case constants.ChangeRoomPrivacy:
				RoomNameChange(clientEvent.Data)
			case constants.ChangeRoomPlayerOne:
				RoomPlayerOneChanged(clientEvent.Data)
			case constants.ChangeRoomPlayerTwo:
				RoomPlayerTwoChanged(clientEvent.Data)
			case constants.ChangeRoomObservers:
				RoomObserversChanged(clientEvent.Data)
			}
		case check := <-checkUsernameChannel:
			check.Valid <- userMap[check.Username] == nil
		case userMapRequest := <- getUserMapChannel:
			userMapRequest <- &userMap
		case roomMapRequest := <- getRoomMapChannel:
			roomMapRequest <- &roomMap
		}
	}
}

func GetRoomUserCount(room *Room) int {
	log.Println("+realtime.GetRoomUserCount")
	members := 0
	if (room.PlayerOne != nil) { members++ }
	if (room.PlayerTwo != nil) { members++ }
	members += room.Observers.Len()
	log.Println("-realtime.GetRoomUserCount")
	return members
}

func init() {
	go realtime()
}
package realtime

import (
	"log"
	"container/list"
	"encoding/json"
    "quarto/constants"
	"github.com/nu7hatch/gouuid"
)

type Room struct {
	Player1		*User
	Player2		*User
	Observers	*list.List
	Events		*list.List
	Name		string
	Private		bool
	Password	string // only used if the room is private
	Urid		string
}

type RoomDTO struct {
	Name		string
	Private		bool
	Password	string
	Urid		string
	Members		int
}

type User struct {
	Username	string // selected username
	Uuid		string // uuid returned by the server
	Room		*Room // room the user is in.
	Events		chan Event // event channel to send messages to user
}

type UserDTO struct {
	Username	string
	Uuid		string
	RoomName	string
}

type Event struct {
	Data		string // json string
	Urid		string // urid of room the massage is going to, "" if to lobby
}

type Data struct {
	Action		string
	Data		string
}

type JoinRoomDTO struct {
	Urid		string
}

func MakeDataString(action, data string) string {
	log.Println("+realtime.MakeDataString")
	dataStruct := Data { action, data }
	dataStr, _ := json.Marshal(dataStruct)
	log.Println("-realtime.MakeDataString")
	return string(dataStr)
}

// Owner must cancel itself when they stop listening to events.
func (u User) Cancel() {
	log.Println("+realtime.Cancel")
	userMap := GetUserMap()
	unsubscribe <- userMap[u.Uuid] // Unsubscribe the channel.
	log.Println("-realtime.Cancel")
}

func Subscribe(uuid, username string) User {
	log.Println("+realtime.Subscribe")
	user := User { username, uuid, nil, make(chan Event, 10) }
	subscribe <- &user
	log.Println("-realtime.Subscribe")
	return user
}

func Action(data, uuid string) {
	log.Println("+realtime.Action")
	userMap := GetUserMap()
	user := userMap[uuid]
	var urid string
	if (user.Room != nil) {
		urid = user.Room.Urid
	} else {
		urid = ""
	}
	publish <- Event{ data, urid }
	log.Println("-realtime.Action")
}

func ServerSideAction(requestData Data, uuid string) {
	log.Println("+realtime.ServerSideAction", requestData)
	var data Data
	json.Unmarshal([]byte(requestData.Data), &data)
	log.Println("Server Action -", data.Action)
	userMap := GetUserMap()
	user := userMap[uuid]
	switch (data.Action) {
	case constants.AddRoom:
		var room RoomDTO
		json.Unmarshal([]byte(data.Data), &room)
		user.Room = AddRoom(room.Name, user, room.Private, room.Password)
	case constants.JoinedRoom:
		var joinRoom JoinRoomDTO
		json.Unmarshal([]byte(data.Data), &joinRoom)
		JoinRoom(joinRoom, user)
	case constants.LeftRoom:
		removeUserFromRoom(user)
	case "becomeObserver":

	case "becomePlayer":

	}
	log.Println("-realtime.ServerSideAction", requestData)
}

const archiveSize = 10

var (
	// Send a channel here to get room events back.  It will send the entire
	// archive initially, and then new messages as they come in.
	subscribe = make(chan *User, 10)
	// Send a channel here to unsubscribe.
	unsubscribe = make(chan *User, 10)
	// Send events here to publish them.
	publish = make(chan Event, 10)
	// 
	checkUsername = make(chan CheckUsername, 10)

	addNewRoom = make(chan *Room, 10)

	users = make(map[string]*User)
	rooms = make(map[string]*Room)
)

type CheckUsername struct {
	uuid string
	valid chan bool
}

func ValidateUsername(name string) bool {
	log.Println("+realtime.ValidateUsername")
	check := CheckUsername{name, make(chan bool)}
	checkUsername <- check
	log.Println("-realtime.ValidateUsername")
	return <-check.valid
}

func GetUserMap() map[string]*User {
	return users
}

func GetRoomMap() map[string]*Room {
	return rooms
}

func JoinRoom(joinRoom JoinRoomDTO, user *User) {
	log.Println("+realtime.JoinRoom")
	roomMap := GetRoomMap()
	room, ok := roomMap[joinRoom.Urid]
	if (!ok) {
		log.Println("Couldn't find room with urid")
		return;
	}
	room.Observers.PushBack(user)
	user.Room = room
	publish <- Event{ MakeDataString(constants.JoinedRoom, user.Username), user.Room.Urid }
	for event := room.Events.Front(); event != nil; event = event.Next() {
		log.Println("Sending stored event", *event.Value.(*Event))
		//user.Events <- *event.Value.(*Event)
	}
	log.Println("-realtime.JoinRoom")
}

func AddRoom(name string, user *User, private bool, password string) *Room {
	log.Println("+realtime.AddRoom")
	userUuid, _ := uuid.NewV4()
	uuidStr := userUuid.String()
	room := Room { user, nil, list.New(), list.New(), name, private, password, uuidStr }
	addNewRoom <- &room
	log.Println("Adding room ", uuidStr)
	log.Println("-realtime.AddRoom")
	return &room;
}

func RemoveRoom(urid string) {
	log.Println("+realtime.RemoveRoom")
	roomMap := GetRoomMap()
	room := roomMap[urid]
	delete(roomMap, urid)
	joinRoomDTO := JoinRoomDTO { urid }
	joinRoomStr, _ := json.Marshal(joinRoomDTO)
	eventDataStr := MakeDataString(constants.RemoveRoom, string(joinRoomStr))
	event := Event { eventDataStr, urid }
	sendEventToRoom(&event, room)
	userMap := GetUserMap()
	sendEventToNoRoomUsers(&event, &userMap)
	log.Println("-realtime.RemoveRoom")
}

// This function loops forever
func realtime() {

	for {
		select {
		case newUser := <-subscribe:
			log.Println("+Adding user as subscriber", newUser.Username)
			users[newUser.Uuid] = newUser
			userDTO := UserDTO { newUser.Username, newUser.Uuid, "" }
			newUserStr, _ := json.Marshal(userDTO)
			eventDataStr := MakeDataString(constants.AddUser, string(newUserStr))
			event := Event { eventDataStr, "" }
			sendEventToNoRoomUsers(&event, &users)
		case event := <-publish:
			room, ok := rooms[event.Urid]
			if (!ok) {
				// not in a room, in the lobby, send to all
				log.Println("Sending event to lobby", event)
				sendEventToNoRoomUsers(&event, &users)
			} else {
				log.Println("Sending event to room", room.Name, event)
				sendEventToRoom(&event, room)
			}
		case unsub := <-unsubscribe:
			for uuid, _ := range users {
				if (uuid == unsub.Uuid) {
					log.Println("Removing user from subscribers", unsub.Username)
					delete(users, uuid)
				}
			}
			userDTO := UserDTO { "", unsub.Uuid, "" }
			leaveStr, _ := json.Marshal(userDTO)
			event := Event { MakeDataString(constants.RemoveUser, string(leaveStr)), unsub.Uuid }
			sendEventToNoRoomUsers(&event, &users)
			removeUserFromRoom(unsub)
		case check := <-checkUsername:
			log.Println("Validating username -- THIS ISN'T WORKING FIX IT.")
			check.valid <- users[check.uuid] == nil
		case newRoom := <- addNewRoom:
			log.Println("Adding new room", newRoom.Name)
			rooms[newRoom.Urid] = newRoom
			newRoomDTO := RoomDTO { newRoom.Name, newRoom.Private, newRoom.Password, newRoom.Urid, 1 }
			newRoomStr, _ := json.Marshal(newRoomDTO)
			event := Event { MakeDataString(constants.AddRoom, string(newRoomStr)), newRoom.Player1.Uuid }
			sendEventToNoRoomUsers(&event, &users)
		}
	}
}

func sendEventToNoRoomUsers(event *Event, userMap *map[string]*User) {
	log.Println("+realtime.sendEventToNoRoomUsers")
	for _, user := range *userMap {
		if (user.Room == nil) {
			user.Events <- *event
		}
	}
	log.Println("-realtime.sendEventToNoRoomUsers")
}

func sendEventToRoom(event *Event, room *Room) {
	log.Println("+realtime.sendEventToRoom")
	if (room.Player1 != nil) {
		room.Player1.Events <- *event
	}
	if (room.Player2 != nil) {
		room.Player2.Events <- *event
	}
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		observer.Value.(*User).Events <- *event
	}
	log.Println("-realtime.sendEventToRoom")
	room.Events.PushBack(event)
}

func GetRoomUserCount(room *Room) int {
	log.Println("+realtime.GetRoomUserCount")
	members := 0
	if (room.Player1 != nil) { members++ }
	if (room.Player2 != nil) { members++ }
	members += room.Observers.Len()
	log.Println("-realtime.GetRoomUserCount")
	return members
}

func removeUserFromRoom(user *User) {
	log.Println("+realtime.removeUserFromRoom")
	if (user.Room != nil) {
		removed := false
		if (user.Room.Player1 != nil && user.Room.Player1.Uuid == user.Uuid) {
			log.Println("Removing Player 1 from room", user.Room.Name) // maybe kill the room?
			user.Room.Player1 = nil;
			removed = true;
		} else if (user.Room.Player2 != nil && user.Room.Player2.Uuid == user.Uuid) {
			log.Println("Removing Player 2 from room", user.Room.Name)
			user.Room.Player2 = nil;
			removed = true;
		} else {
			for observer := user.Room.Observers.Front(); observer != nil; observer = observer.Next() {
				if (observer.Value.(*User).Uuid == user.Uuid) {
					log.Println("Removing an observer from room", user.Room.Name)
					user.Room.Observers.Remove(observer)
					removed = true;
					break;
				}
			}
		}
		if (removed) {
			publish <- Event{ MakeDataString(constants.LeftRoom, user.Username), user.Room.Urid }
			remainingMembers := GetRoomUserCount(user.Room)
			if (remainingMembers == 0) {
				RemoveRoom(user.Room.Urid)
			}
		} else {
			log.Println("Couldn't find user in room!")
		}
	} else {
		log.Println(user.Username, "wasn't even in a room!")
	}
	log.Println("-realtime.removeUserFromRoom")
}

func init() {
	go realtime()
}

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
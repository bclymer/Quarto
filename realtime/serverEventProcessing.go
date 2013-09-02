package realtime

import (
	"quarto/constants"
	"encoding/json"
	"log"
)

func AddUser(addUserMessage string, userMap map[string]*User) *User {
	var addUserDTO AddUserDTO
	err := json.Unmarshal([]byte(addUserMessage), &addUserDTO)
	if (err != nil) {
		log.Fatal("AddUser: Couldn't unmarshal", addUserMessage)
		return nil
	}
	user := User { addUserDTO.Username, nil, make(chan ClientEvent) }

	userMap[addUserDTO.Username] = &user

	clientEvent := ClientEvent { constants.AddUser, addUserMessage }
	sendEventToLobby(&clientEvent, userMap)

	return &user
}

func RemoveUser(removeUserMessage string, userMap map[string]*User, roomMap map[string]*Room) {
	var removeUserDTO RemoveUserDTO
	err := json.Unmarshal([]byte(removeUserMessage), &removeUserDTO)
	if (err != nil) {
		log.Fatal("RemoveUser: Couldn't unmarshal", removeUserMessage)
		return
	}
	clientEvent := ClientEvent { constants.RemoveUser, removeUserMessage }

	user := userMap[removeUserDTO.Username]
	delete(userMap, removeUserDTO.Username)

	if (user.Room != nil) {
		sendEventToRoom(&clientEvent, user.Room.Name, roomMap)
	}
	sendEventToLobby(&clientEvent, userMap)
}

func UserChallengedUser(userChallengedUser, username string) {

}

func JoinRoom(joinRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	var joinRoomDTO JoinRoomDTO
	err := json.Unmarshal([]byte(joinRoomMessage), &joinRoomDTO)
	if (err != nil) {
		log.Fatal("JoinRoom: Couldn't unmarshal", joinRoomMessage)
		return
	}
	user, ok := userMap[username]
	if (!ok) {
		log.Fatal("JoinRoom: User didn't exist")
		return
	}
	room, ok := roomMap[joinRoomDTO.Name]
	if (!ok) {
		log.Fatal("JoinRoom: Room didn't exist")
		return
	}

	user.Room = room
	room.Observers.PushBack(user)

	userRoomDTO := UserRoomDTO { username, room.Name }
	userRoomByteArray, err := json.Marshal(userRoomDTO)
	if (err != nil) {
		log.Fatal("JoinRoom: Couldn't marshal", userRoomDTO)
	}
	clientEvent := ClientEvent { constants.JoinRoom, string(userRoomByteArray) }
	sendEventToRoom(&clientEvent, room.Name, roomMap)
	sendEventToLobby(&clientEvent, userMap)

	// TODO: Send all past events
	/*
	publish <- Event{ MakeDataString(constants.JoinedRoom, user.Username), user.Room.Urid }
	for event := room.Events.Front(); event != nil; event = event.Next() {
		log.Println("Sending stored event", *event.Value.(*Event))
		//user.Events <- *event.Value.(*Event)
	}
	*/
}

func LeaveRoom(leaveRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	var leaveRoomDTO LeaveRoomDTO
	err := json.Unmarshal([]byte(leaveRoomMessage), &leaveRoomDTO)
	if (err != nil) {
		log.Fatal("LeaveRoom: Couldn't unmarshal", leaveRoomMessage)
		return
	}
	user, ok := userMap[username]
	if (!ok) {
		log.Fatal("LeaveRoom: User didn't exist")
		return
	}
	room, ok := roomMap[leaveRoomDTO.Name]
	if (!ok) {
		log.Fatal("LeaveRoom: Room didn't exist")
		return
	}

	if (room.PlayerOne != nil && room.PlayerOne.Username == user.Username) {
		room.PlayerOne = nil
		changeEvent, _ := json.Marshal(RoomPlayerOneChangedDTO { "" })
		RoomPlayerOneChanged(string(changeEvent))
	} else if (room.PlayerTwo != nil && room.PlayerTwo.Username == user.Username) {
		room.PlayerTwo = nil
		changeEvent, _ := json.Marshal(RoomPlayerTwoChangedDTO { "" })
		RoomPlayerTwoChanged(string(changeEvent))
	} else {
		for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
			if (observer.Value.(*User).Username == user.Username) {
				room.Observers.Remove(observer)
				changeEvent, _ := json.Marshal(RoomObserversChangedDTO { "" })
				RoomObserversChanged(string(changeEvent))
				break
			}
		}
	}
}

func AddRoom(addRoomMessage, username string) { // AddRoomDTO
	/*
	log.Println("+realtime.AddRoom")
	userUuid, _ := uuid.NewV4()
	uuidStr := userUuid.String()
	room := Room { user, nil, list.New(), list.New(), name, private, password, uuidStr }
	addNewRoom <- &room
	log.Println("Adding room ", uuidStr)
	log.Println("-realtime.AddRoom")
	return &room;
	*/
}

func RemoveRoom(removeRoomMessage string) { // RemoveRoomDTO
	/*
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
	*/
}

func RoomNameChange(roomNameChanged string) { // RoomNameChangedDTO

}

func RoomPlayerOneChanged(roomPlayerOneChanged string) { // RoomPlayerOneChangedDTO

}

func RoomPlayerTwoChanged(roomPlayerTwoChanged string) { // RoomPlayerTwoChangedDTO

}

func RoomObserversChanged(roomObserversAdded string) { // RoomObserversAddedDTO

}

func sendEventToLobby(clientEvent *ClientEvent, userMap map[string]*User) {
	for _, user := range userMap {
		if (user.Room == nil) {
			user.Events <- *clientEvent
		}
	}
}

func sendEventToRoom(clientEvent *ClientEvent, roomName string, roomMap map[string]*Room) {
	room, ok := roomMap[roomName]
	if (!ok) {
		log.Fatal("sendEventToRoom: Room didn't exist")
		return // room didn't exist
	}
	if (room.PlayerOne != nil) {
		room.PlayerOne.Events <- *clientEvent
	}
	if (room.PlayerTwo != nil) {
		room.PlayerTwo.Events <- *clientEvent
	}
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		observer.Value.(*User).Events <- *clientEvent
	}
	room.Events.PushBack(clientEvent)
}

func sendEventToUser(clientEvent *ClientEvent, username string, userMap map[string]*User) {
	user, ok := userMap[username]
	if (!ok) {
		log.Fatal("sendEventToUser: User didn't exist")
		return // user didn't exist
	}
	
	user.Events <- *clientEvent
}

/*
func ServerSideAction(requestData Data, uuid string) {
	log.Println("+realtime.ServerSideAction", requestData)
	var data ClientEvent
	json.Unmarshal([]byte(requestData.Data), &data)
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
*/
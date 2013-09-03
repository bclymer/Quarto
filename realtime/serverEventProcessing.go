package realtime

import (
	"container/list"
	"encoding/json"
	"log"
	"quarto/constants"
)

func AddUser(addUserMessage string, userMap map[string]*User) *User {
	log.Println("+AddUser")
	var addUserDTO AddUserDTO
	err := json.Unmarshal([]byte(addUserMessage), &addUserDTO)
	if err != nil {
		log.Fatal("AddUser: Couldn't unmarshal", addUserMessage)
		return nil
	}
	user := User{addUserDTO.Username, nil, make(chan *ClientEvent, 10)}

	userMap[addUserDTO.Username] = &user

	clientEvent := ClientEvent{constants.AddUser, addUserMessage}
	sendEventToLobby(&clientEvent, userMap)
	log.Println("-AddUser")
	return &user
}

func RemoveUser(removeUserMessage string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+RemoveUser")
	var removeUserDTO RemoveUserDTO
	err := json.Unmarshal([]byte(removeUserMessage), &removeUserDTO)
	if err != nil {
		log.Fatal("RemoveUser: Couldn't unmarshal", removeUserMessage)
		return
	}
	clientEvent := ClientEvent{constants.RemoveUser, removeUserMessage}

	user := userMap[removeUserDTO.Username]

	if user.Room != nil {
		leaveRoomDTO := LeaveRoomDTO{user.Room.Name}
		leaveRoomDTOByteArray, err := json.Marshal(leaveRoomDTO)
		if err != nil {
			log.Fatal("RemoveUser: Couldn't marshal", leaveRoomDTO)
			return
		}
		LeaveRoom(string(leaveRoomDTOByteArray), removeUserDTO.Username, userMap, roomMap)
		sendEventToRoom(&clientEvent, user.Room.Name, roomMap)
	}

	delete(userMap, removeUserDTO.Username)

	sendEventToLobby(&clientEvent, userMap)
	log.Println("-RemoveUser")
}

func UserChallengedUser(userChallengedUser, username string) {

}

func JoinRoom(joinRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+JoinRoom")
	var joinRoomDTO JoinRoomDTO
	err := json.Unmarshal([]byte(joinRoomMessage), &joinRoomDTO)
	if err != nil {
		log.Fatal("JoinRoom: Couldn't unmarshal", joinRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Fatal("JoinRoom: User didn't exist")
		return
	}
	room, ok := roomMap[joinRoomDTO.Name]
	if !ok {
		log.Fatal("JoinRoom: Room didn't exist")
		return
	}

	user.Room = room
	room.Observers.PushBack(user)

	userRoomDTO := UserRoomDTO{username, room.Name}
	userRoomByteArray, err := json.Marshal(userRoomDTO)
	if err != nil {
		log.Fatal("JoinRoom: Couldn't marshal", userRoomDTO)
		return
	}
	lobbyUserEvent := ClientEvent{constants.JoinRoom, string(userRoomByteArray)}
	sendEventToLobby(&lobbyUserEvent, userMap)

	lobbyRoomDTO := MakeLobbyRoomDTO(room)
	lobbyRoomByteArray, err := json.Marshal(lobbyRoomDTO)
	if err != nil {
		log.Fatal("JoinRoom: Couldn't marshal", lobbyRoomDTO)
		return
	}
	lobbyRoomEvent := ClientEvent{constants.ChangeRoom, string(lobbyRoomByteArray)}
	sendEventToLobby(&lobbyRoomEvent, userMap)

	roomRoomDTO := MakeRoomRoomDTO(room)
	roomRoomByteArray, err := json.Marshal(roomRoomDTO)
	if err != nil {
		log.Fatal("JoinRoom: Couldn't marshal", roomRoomDTO)
		return
	}
	roomRoomEvent := ClientEvent{constants.ChangeRoom, string(roomRoomByteArray)}
	sendEventToRoom(&roomRoomEvent, room.Name, roomMap)
	log.Println("-JoinRoom")

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
	log.Println("+LeaveRoom")
	var leaveRoomDTO LeaveRoomDTO
	err := json.Unmarshal([]byte(leaveRoomMessage), &leaveRoomDTO)
	if err != nil {
		log.Fatal("LeaveRoom: Couldn't unmarshal", leaveRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Fatal("LeaveRoom: User didn't exist")
		return
	}
	room, ok := roomMap[leaveRoomDTO.Name]
	if !ok {
		log.Fatal("LeaveRoom: Room didn't exist")
		return
	}

	if room.PlayerOne != nil && room.PlayerOne.Username == user.Username {
		room.PlayerOne = nil
		changeEvent, _ := json.Marshal(RoomPlayerOneChangedDTO{""})
		RoomPlayerOneChanged(string(changeEvent))
	} else if room.PlayerTwo != nil && room.PlayerTwo.Username == user.Username {
		room.PlayerTwo = nil
		changeEvent, _ := json.Marshal(RoomPlayerTwoChangedDTO{""})
		RoomPlayerTwoChanged(string(changeEvent))
	} else {
		for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
			if observer.Value.(*User).Username == user.Username {
				room.Observers.Remove(observer)
				changeEvent, _ := json.Marshal(RoomObserversChangedDTO{"", false})
				RoomObserversChanged(string(changeEvent))
				break
			}
		}
	}
	userRoomDTO := UserRoomDTO{username, room.Name}
	userRoomByteArray, err := json.Marshal(userRoomDTO)
	if err != nil {
		log.Fatal("LeaveRoom: Couldn't marshal", userRoomDTO)
		return
	}
	clientEvent := ClientEvent{constants.LeaveRoom, string(userRoomByteArray)}
	sendEventToLobby(&clientEvent, userMap)
	sendEventToRoom(&clientEvent, room.Name, roomMap)
	log.Println("-LeaveRoom")
}

func AddRoom(addRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+AddRoom")
	var addRoomDTO AddRoomDTO
	err := json.Unmarshal([]byte(addRoomMessage), &addRoomDTO)
	if err != nil {
		log.Fatal("AddRoom: Couldn't unmarshal", addRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Fatal("AddRoom: Couldn't find user who created the room")
		return
	}
	room := Room{user, nil, list.New(), list.New(), addRoomDTO.Name, addRoomDTO.Private, addRoomDTO.Password}
	roomMap[addRoomDTO.Name] = &room
	user.Room = &room

	lobbyRoomDTO := LobbyRoomDTO{addRoomDTO.Name, addRoomDTO.Private, 1}
	lobbyRoomDTOByteArray, err := json.Marshal(lobbyRoomDTO)
	if err != nil {
		log.Fatal("AddRoom: Couldn't marshal", lobbyRoomDTO)
		return
	}
	clientEvent := ClientEvent{constants.AddRoom, string(lobbyRoomDTOByteArray)}
	sendEventToLobby(&clientEvent, userMap)
	log.Println("-AddRoom")
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

func Chat(incomingChat, username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+Chat")
	var incomingChatDTO IncomingChatDTO
	err := json.Unmarshal([]byte(incomingChat), &incomingChatDTO)
	if err != nil {
		log.Fatal("Chat: Couldn't unmarshal", incomingChat)
		return
	}

	user := userMap[username]

	if user.Room != nil {
		outgoingChatDTO := OutgoingChatDTO{incomingChatDTO.Message, username}
		outgoingChatByteArray, err := json.Marshal(outgoingChatDTO)
		if err != nil {
			log.Fatal("LeaveRoom: Couldn't marshal", outgoingChatDTO)
			return
		}
		clientEvent := ClientEvent{constants.Chat, string(outgoingChatByteArray)}
		sendEventToRoom(&clientEvent, user.Room.Name, roomMap)
	}
}

func sendEventToLobby(clientEvent *ClientEvent, userMap map[string]*User) {
	for _, user := range userMap {
		if user.Room == nil {
			user.Events <- clientEvent
		}
	}
}

func sendEventToRoom(clientEvent *ClientEvent, roomName string, roomMap map[string]*Room) {
	room, ok := roomMap[roomName]
	if !ok {
		log.Fatal("sendEventToRoom: Room didn't exist")
		return // room didn't exist
	}
	if room.PlayerOne != nil {
		room.PlayerOne.Events <- clientEvent
	}
	if room.PlayerTwo != nil {
		room.PlayerTwo.Events <- clientEvent
	}
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		observer.Value.(*User).Events <- clientEvent
	}
	room.Events.PushBack(clientEvent)
}

func sendEventToUser(clientEvent *ClientEvent, username string, userMap map[string]*User) {
	user, ok := userMap[username]
	if !ok {
		log.Fatal("sendEventToUser: User didn't exist")
		return // user didn't exist
	}

	user.Events <- clientEvent
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

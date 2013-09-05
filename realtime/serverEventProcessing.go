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
		log.Println("AddUser: Couldn't unmarshal", addUserMessage)
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
		log.Println("RemoveUser: Couldn't unmarshal", removeUserMessage)
		return
	}
	clientEvent := ClientEvent{constants.RemoveUser, removeUserMessage}

	user := userMap[removeUserDTO.Username]

	if user.Room != nil {
		LeaveRoom(removeUserDTO.Username, userMap, roomMap)
	}
	delete(userMap, removeUserDTO.Username)
	sendEventToLobby(&clientEvent, userMap)
	close(user.Events)
	log.Println("-RemoveUser")
}

func UserChallengedUser(userChallengedUser, username string) {

}

func JoinRoom(joinRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+JoinRoom")
	var joinRoomDTO JoinRoomDTO
	err := json.Unmarshal([]byte(joinRoomMessage), &joinRoomDTO)
	if err != nil {
		log.Println("JoinRoom: Couldn't unmarshal", joinRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Println("JoinRoom: User didn't exist")
		return
	}
	room, ok := roomMap[joinRoomDTO.Name]
	if !ok {
		log.Println("JoinRoom: Room didn't exist")
		return
	}

	user.Room = room
	room.Observers.PushBack(user)

	userRoomString, err := DtoToString(UserRoomDTO{username, room.Name})
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	lobbyUserEvent := ClientEvent{constants.JoinRoom, userRoomString}
	sendEventToLobby(&lobbyUserEvent, userMap)

	lobbyRoomString, err := DtoToString(MakeLobbyRoomDTO(room))
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	lobbyRoomEvent := ClientEvent{constants.ChangeRoom, lobbyRoomString}
	sendEventToLobby(&lobbyRoomEvent, userMap)

	roomRoomString, err := DtoToString(MakeRoomRoomDTO(room))
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	roomRoomEvent := ClientEvent{constants.ChangeRoom, roomRoomString}
	sendEventToRoom(&roomRoomEvent, room)
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

func LeaveRoom(username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+LeaveRoom")
	user, ok := userMap[username]
	if !ok {
		log.Println("LeaveRoom: User didn't exist")
		return
	}
	room := user.Room;
	user.Room = nil;
	if room == nil {
		log.Println("LeaveRoom: User wasn't in a room")
		return
	}

	if room.PlayerOne != nil && room.PlayerOne.Username == user.Username {
		room.PlayerOne = nil
		changeEvent, _ := DtoToString(RoomPlayerOneChangedDTO{""})
		RoomPlayerOneChanged(changeEvent)
	} else if room.PlayerTwo != nil && room.PlayerTwo.Username == user.Username {
		room.PlayerTwo = nil
		changeEvent, _ := DtoToString(RoomPlayerTwoChangedDTO{""})
		RoomPlayerTwoChanged(changeEvent)
	} else {
		for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
			if observer.Value.(*User).Username == user.Username {
				room.Observers.Remove(observer)
				changeEvent, _ := DtoToString(RoomObserversChangedDTO{"", false})
				RoomObserversChanged(changeEvent)
				break
			}
		}
	}

	userRoomString, err := DtoToString(UserRoomDTO{username, room.Name})
	if err != nil {
		log.Println("LeaveRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.LeaveRoom, userRoomString}
	sendEventToLobby(&clientEvent, userMap)
	sendEventToRoom(&clientEvent, room)
	log.Println("-LeaveRoom")
}

func AddRoom(addRoomMessage, username string, userMap map[string]*User, roomMap map[string]*Room) {
	log.Println("+AddRoom")
	var addRoomDTO AddRoomDTO
	err := json.Unmarshal([]byte(addRoomMessage), &addRoomDTO)
	if err != nil {
		log.Println("AddRoom: Couldn't unmarshal", addRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Println("AddRoom: Couldn't find user who created the room")
		return
	}
	room := Room{nil, nil, list.New(), list.New(), addRoomDTO.Name, addRoomDTO.Private, addRoomDTO.Password, MakeNewGame()}
	roomMap[addRoomDTO.Name] = &room
	user.Room = &room

	lobbyRoomString, err := DtoToString(LobbyRoomDTO{addRoomDTO.Name, addRoomDTO.Private, 1})
	if err != nil {
		log.Println("AddRoom: Couldn't marshal.", err)
		return
	}
	addRoomClientEvent := ClientEvent{constants.AddRoom, lobbyRoomString}
	sendEventToLobby(&addRoomClientEvent, userMap)

	joinRoomString, err := DtoToString(JoinRoomDTO{addRoomDTO.Name})
	if err != nil {
		log.Println("AddRoom: Couldn't marshal.", err)
		return
	}

	JoinRoom(joinRoomString, username, userMap, roomMap)

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
		log.Println("Chat: Couldn't unmarshal", incomingChat)
		return
	}

	user := userMap[username]

	if user.Room != nil {
		outgoingChatString, err := DtoToString(OutgoingChatDTO{incomingChatDTO.Message, username})
		if err != nil {
			log.Println("LeaveRoom: Couldn't marshal.", err)
			return
		}
		clientEvent := ClientEvent{constants.Chat, outgoingChatString}
		sendEventToRoom(&clientEvent, user.Room)
	}
}

func playerChanged(username string, userMap map[string]*User, roomMap map[string]*Room) (*User, bool) {
	user, ok := userMap[username]
	if !ok {
		log.Println("User who requested player one wasn't a user", username)
		return nil, false
	}
	if user.Room == nil {
		log.Println("User who requested player one wasn't in a room", user)
		return nil, false
	}
	user.Room.Game.Reset()
	gameString, err := DtoToString(user.Room.Game)
	if err != nil {
		log.Println("playerChanged: Couldn't marshal.", err)
		return nil, false
	}
	clientEvent := ClientEvent{constants.ChangeGame, gameString}
	sendEventToRoom(&clientEvent, user.Room)
	return user, true
}

func RequestPlayerOne(username string, userMap map[string]*User, roomMap map[string]*Room) {
	user, ok := playerChanged(username, userMap, roomMap)
	if !ok {
		return
	}
	if user.Room.PlayerOne != nil {
		log.Println("Couldn't change player 1, player 1 is taken")
		return
	}
	user.Room.PlayerOne = user
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("RequestPlayerOne: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.ChangeRoom, roomRoomString}
	user.Room.UpdateGame()
	sendEventToRoom(&clientEvent, user.Room)
}

func RequestPlayerTwo(username string, userMap map[string]*User, roomMap map[string]*Room) {
	user, ok := playerChanged(username, userMap, roomMap)
	if !ok {
		return
	}
	if user.Room.PlayerTwo != nil {
		log.Println("Couldn't change player 2, player 2 is taken")
		return
	}
	user.Room.PlayerTwo = user
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("RequestPlayerTwo: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.ChangeRoom, roomRoomString}
	user.Room.UpdateGame()
	sendEventToRoom(&clientEvent, user.Room)
}

func LeavePlayerOne(username string, userMap map[string]*User, roomMap map[string]*Room) {
	user, ok := playerChanged(username, userMap, roomMap)
	if !ok {
		return
	}
	if user.Room.PlayerOne == nil {
		log.Println("Couldn't leave player 1, player 1 is nil")
		return
	}
	user.Room.PlayerOne = nil
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("LeavePlayerOne: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.ChangeRoom, roomRoomString}
	user.Room.UpdateGame()
	sendEventToRoom(&clientEvent, user.Room)
}

func LeavePlayerTwo(username string, userMap map[string]*User, roomMap map[string]*Room) {
	user, ok := playerChanged(username, userMap, roomMap)
	if !ok {
		return
	}
	if user.Room.PlayerTwo == nil {
		log.Println("Couldn't leave player 2, player 2 is nil")
		return
	}
	user.Room.PlayerTwo = nil
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("LeavePlayerTwo: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.ChangeRoom, roomRoomString}
	user.Room.UpdateGame()
	sendEventToRoom(&clientEvent, user.Room)
}

func sendEventToLobby(clientEvent *ClientEvent, userMap map[string]*User) {
	for _, user := range userMap {
		if user.Room == nil {
			user.Events <- clientEvent
		}
	}
}

func sendEventToRoom(clientEvent *ClientEvent, room *Room) {
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
		log.Println("sendEventToUser: User didn't exist")
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

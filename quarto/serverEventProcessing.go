package quarto

import (
	"container/list"
	"encoding/json"
	"log"
	"strings"
)

var (
	userMap map[string]*User
	roomMap map[string]*Room
)

func setupEventProcessor(users *map[string]*User, rooms *map[string]*Room) {
	userMap = *users
	roomMap = *rooms
}

func AddUser(addUserMessage string) *User {
	log.Println("+AddUser")
	var addUserDTO AddUserDTO
	err := json.Unmarshal([]byte(addUserMessage), &addUserDTO)
	if err != nil {
		log.Println("AddUser: Couldn't unmarshal", addUserMessage, err)
		return nil
	}
	if _, ok := userMap[addUserDTO.Username]; ok {
		return nil
	}

	user := User{addUserDTO.Username, nil, make(chan *ClientEvent), false}

	userMap[addUserDTO.Username] = &user

	//clientEvent := ClientEvent{Config.UserAdd, addUserMessage}

	//sendEventToLobby(&clientEvent)
	log.Println("-AddUser")
	user.Active = true
	return &user
}

func RemoveUser(removeUserMessage string) {
	log.Println("+RemoveUser")
	var removeUserDTO RemoveUserDTO
	err := json.Unmarshal([]byte(removeUserMessage), &removeUserDTO)
	if err != nil {
		log.Println("RemoveUser: Couldn't unmarshal", removeUserMessage, err)
		return
	}

	user, _, ok := validateUserOrRoom(removeUserDTO.Username, "", "RemoveUser", false)
	if !ok {
		return
	}
	user.Active = false
	if user.Room != nil {
		LeaveRoom(removeUserDTO.Username)
	}
	delete(userMap, removeUserDTO.Username)

	//clientEvent := ClientEvent{Config.UserRemove, removeUserMessage}
	//sendEventToLobby(&clientEvent)
	close(user.Events)
	log.Println("-RemoveUser")
}

func UserChallengedUser(userChallengedUser, username string) {

}

func JoinRoom(joinRoomMessage, username string) {
	log.Println("+JoinRoom")
	var joinRoomDTO JoinRoomDTO
	err := json.Unmarshal([]byte(joinRoomMessage), &joinRoomDTO)
	if err != nil {
		log.Println("JoinRoom: Couldn't unmarshal", joinRoomMessage, err, err)
		return
	}
	user, room, ok := validateUserOrRoom(username, joinRoomDTO.Name, "JoinRoom", true)
	if !ok {
		return
	}

	user.Room = room
	room.Observers.PushBack(user)

	// userRoomString, err := DtoToString(UserRoomDTO{username, room.Name})
	// if err != nil {
	// 	log.Println("JoinRoom: Couldn't marshal.", err)
	// 	sendErrorToUser(user, "")
	// 	return
	// }

	// lobbyUserEvent := ClientEvent{Config.UserRoomJoin, userRoomString}
	// sendEventToLobby(&lobbyUserEvent)

	// lobbyRoomString, err := DtoToString(MakeLobbyRoomDTO(room))
	// if err != nil {
	// 	log.Println("JoinRoom: Couldn't marshal.", err)
	// 	return
	// }
	// lobbyRoomEvent := ClientEvent{Config.RoomChange, lobbyRoomString}
	// sendEventToLobby(&lobbyRoomEvent)

	roomRoomString, err := DtoToString(MakeRoomRoomDTO(room))
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	roomRoomEvent := ClientEvent{Config.RoomChange, roomRoomString}
	sendEventToRoom(&roomRoomEvent, room)
	sendInfoToUser(user, "You've been added as an observer to the room "+room.Name)
	log.Println("-JoinRoom")

	for event := room.Events.Front(); event != nil; event = event.Next() {
		sendEventToUser(event.Value.(*ClientEvent), user)
	}

	gameString, err := DtoToString(room.Game)
	if err != nil {
		log.Println("updateGameForRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.GameChange, gameString}
	sendEventToUser(&clientEvent, user)
}

func LeaveRoom(username string) {
	log.Println("+LeaveRoom")
	user, room, ok := validateUserOrRoom(username, "", "LeaveRoom", true)
	if !ok {
		return
	}

	removed := false
	if room.PlayerOne != nil && room.PlayerOne.Username == user.Username {
		room.PlayerOne = nil
		removed = true
	} else if room.PlayerTwo != nil && room.PlayerTwo.Username == user.Username {
		room.PlayerTwo = nil
		removed = true
	} else {
		for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
			if observer.Value.(*User).Username == user.Username {
				room.Observers.Remove(observer)
				removed = true
				break
			}
		}
	}
	if removed {
		roomRoomString, err := DtoToString(MakeRoomRoomDTO(room))
		if err != nil {
			log.Println("LeaveRoom: Couldn't marshal.", err)
			return
		}
		clientEvent := ClientEvent{Config.RoomChange, roomRoomString}
		sendEventToRoom(&clientEvent, room)
		room.UpdateGame()
		updateGameForRoom(room)
	}

	userRoomString, err := DtoToString(UserRoomDTO{username, room.Name})
	if err != nil {
		log.Println("LeaveRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.UserRoomLeave, userRoomString}
	//sendEventToLobby(&clientEvent)
	sendEventToRoom(&clientEvent, room)

	if GetRoomUserCount(room) == 0 {
		removeRoomString, err := DtoToString(RemoveRoomDTO{room.Name})
		if err != nil {
			log.Println("LeaveRoom: Couldn't marshal.", removeRoomString)
			return
		}
		RemoveRoom(removeRoomString)
	}

	log.Println("-LeaveRoom")
}

func AddRoom(addRoomMessage, username string) {
	log.Println("+AddRoom")
	var addRoomDTO AddRoomDTO
	err := json.Unmarshal([]byte(addRoomMessage), &addRoomDTO)
	if err != nil {
		log.Println("AddRoom: Couldn't unmarshal", addRoomMessage, err)
		return
	}

	user, _, ok := validateUserOrRoom(username, "", "AddRoom", false)
	if !ok {
		return
	}
	if user.Room != nil {
		sendErrorToUser(user, "Can't add a room while you're already in a room")
		log.Println("AddRoom: User tried to add room while in a room")
		return
	}
	if _, ok = roomMap[addRoomDTO.Name]; ok {
		sendErrorToUser(user, "A room with the name "+addRoomDTO.Name+" already exists.")
		return
	}
	if strings.TrimSpace(addRoomDTO.Name) == "" {
		sendErrorToUser(user, "The room name can't be blank")
		return
	}
	room := Room{nil, nil, list.New(), list.New(), addRoomDTO.Name, addRoomDTO.Private, addRoomDTO.Password, MakeNewGame()}
	roomMap[addRoomDTO.Name] = &room
	user.Room = &room

	// lobbyRoomString, err := DtoToString(LobbyRoomDTO{addRoomDTO.Name, addRoomDTO.Private, 1})
	// if err != nil {
	// 	log.Println("AddRoom: Couldn't marshal.", err)
	// 	return
	// }
	// addRoomClientEvent := ClientEvent{Config.RoomAdd, lobbyRoomString}
	// sendEventToLobby(&addRoomClientEvent)

	joinRoomString, err := DtoToString(JoinRoomDTO{addRoomDTO.Name})
	if err != nil {
		log.Println("AddRoom: Couldn't marshal.", err)
		return
	}

	JoinRoom(joinRoomString, username)

	log.Println("-AddRoom")
}

func RemoveRoom(removeRoomMessage string) {
	log.Println("+RemoveRoom")
	var removeRoomDTO RemoveRoomDTO
	err := json.Unmarshal([]byte(removeRoomMessage), &removeRoomDTO)
	if err != nil {
		log.Println("Remove Room: Couldn't unmarshal", removeRoomMessage, err)
		return
	}

	delete(roomMap, removeRoomDTO.Name)

	//clientEvent := ClientEvent{Config.RoomRemove, removeRoomMessage}
	//sendEventToLobby(&clientEvent)
	log.Println("-RemoveRoom")
}

func RoomNameChange(roomNameChanged string) { // RoomNameChangedDTO

}

func Chat(incomingChat, username string) {
	log.Println("+Chat")
	var incomingChatDTO IncomingChatDTO
	err := json.Unmarshal([]byte(incomingChat), &incomingChatDTO)
	if err != nil {
		log.Println("Chat: Couldn't unmarshal", incomingChat, err)
		return
	}

	_, room, ok := validateUserOrRoom(username, "", "Chat", true)
	if !ok {
		return
	}

	outgoingChatString, err := DtoToString(OutgoingChatDTO{incomingChatDTO.Message, username})
	if err != nil {
		log.Println("LeaveRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.Chat, outgoingChatString}
	sendEventToRoom(&clientEvent, room)
	if room.Events.Len() >= 10 {
		room.Events.Remove(room.Events.Front())
	}
	room.Events.PushBack(&clientEvent)

	log.Println("-Chat")
}

func RequestPlayerOne(username string) {
	log.Println("+RequestPlayerOne")
	user, ok := playerChanged(username)
	if !ok {
		return
	}
	if user.Room.PlayerOne != nil {
		sendInfoToUser(user, "You can't switch to player 1, player 1 is already taken!")
		log.Println("Couldn't change player 1, player 1 is taken")
		return
	}
	if user.Room.PlayerTwo == user {
		sendInfoToUser(user, "You can't switch to player 1, you're already player 2!")
		log.Println("Couldn't change player 1, user is player 2")
		return
	}
	user.Room.PlayerOne = user
	RemoveFromObservers(user)
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("RequestPlayerOne: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.RoomChange, roomRoomString}
	user.Room.UpdateGame()
	updateGameForRoom(user.Room)
	sendEventToRoom(&clientEvent, user.Room)
	log.Println("-RequestPlayerOne")
}

func RequestPlayerTwo(username string) {
	log.Println("+RequestPlayerTwo")
	user, ok := playerChanged(username)
	if !ok {
		return
	}
	if user.Room.PlayerTwo != nil {
		sendInfoToUser(user, "You can't switch to player 2, player 2 is already taken!")
		log.Println("Couldn't change player 2, player 2 is taken")
		return
	}
	if user.Room.PlayerOne == user {
		sendInfoToUser(user, "You can't switch to player 2, you're already player 1!")
		log.Println("Couldn't change player 2, user is player 1")
		return
	}
	user.Room.PlayerTwo = user
	RemoveFromObservers(user)
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("RequestPlayerTwo: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.RoomChange, roomRoomString}
	user.Room.UpdateGame()
	updateGameForRoom(user.Room)
	sendEventToRoom(&clientEvent, user.Room)
	log.Println("-RequestPlayerTwo")
}

func LeavePlayerOne(username string) {
	log.Println("+LeavePlayerOne")
	user, ok := playerChanged(username)
	if !ok {
		return
	}
	if user.Room.PlayerOne == nil {
		log.Println("Couldn't leave player 1, player 1 is nil")
		return
	}
	user.Room.PlayerOne = nil
	user.Room.Observers.PushBack(user)
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("LeavePlayerOne: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.RoomChange, roomRoomString}
	user.Room.Game.Reset()
	updateGameForRoom(user.Room)
	sendEventToRoom(&clientEvent, user.Room)
	log.Println("-LeavePlayerOne")
}

func LeavePlayerTwo(username string) {
	log.Println("+LeavePlayerTwo")
	user, ok := playerChanged(username)
	if !ok {
		return
	}
	if user.Room.PlayerTwo == nil {
		log.Println("Couldn't leave player 2, player 2 is nil")
		return
	}
	user.Room.Observers.PushBack(user)
	user.Room.PlayerTwo = nil
	roomRoomString, err := DtoToString(MakeRoomRoomDTO(user.Room))
	if err != nil {
		log.Println("LeavePlayerTwo: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.RoomChange, roomRoomString}
	user.Room.Game.Reset()
	updateGameForRoom(user.Room)
	sendEventToRoom(&clientEvent, user.Room)
	log.Println("-LeavePlayerTwo")
}

func GamePiecePlayed(gamePiecePlayed, username string) {
	log.Println("+GamePiecePlayed")
	var gamePiecePlayedDTO GamePiecePlayedDTO
	err := json.Unmarshal([]byte(gamePiecePlayed), &gamePiecePlayedDTO)
	if err != nil {
		log.Println("GamePiecePlayed: Couldn't unmarshal", gamePiecePlayed, err)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Println("GamePiecePlayed: User didn't exist")
		return
	}
	if user.Room == nil {
		log.Println("GamePiecePlayed: User wasn't in room")
		return
	}
	if user.Room.Game.GameState == GameStatePlayerOnePlaying {
		if user.Room.PlayerOne != user {
			sendInfoToUser(user, "You can't play that piece, you're not player 1!")
			log.Println("GamePiecePlayed: User wasn't player one, can't play piece now.")
			return
		}
		user.Room.Game.GameState = GameStatePlayerOneChoosing
	} else if user.Room.Game.GameState == GameStatePlayerTwoPlaying {
		if user.Room.PlayerTwo != user {
			sendInfoToUser(user, "You can't play that piece, you're not player 2!")
			log.Println("GamePiecePlayed: User wasn't player two, can't play piece now.")
			return
		}
		user.Room.Game.GameState = GameStatePlayerTwoChoosing
	} else {
		sendErrorToUser(user, "Are you doing something sneaky? You shouldn't have been able to send that request...")
		log.Println("GamePiecePlayed: Game not in proper state to play piece")
		return
	}
	if user.Room.Game.Board[gamePiecePlayedDTO.Location] != -1 {
		sendInfoToUser(user, "There is already a piece there, you can't play on that board location.")
		return
	}
	user.Room.Game.Board[gamePiecePlayedDTO.Location] = user.Room.Game.SelectedPiece
	user.Room.Game.SelectedPiece = -1

	if winner := user.Room.Game.CheckWinner(); winner != 0 {
		var winnerName string
		if winner == 1 {
			winnerName = user.Room.PlayerOne.Username
		} else {
			winnerName = user.Room.PlayerTwo.Username
		}
		gameWinnerString, err := DtoToString(GameWinnerDTO{winnerName})
		if err != nil {
			log.Println("GamePiecePlayed: Couldn't marshal.", err)
		}
		clientEvent := ClientEvent{Config.GameWinner, gameWinnerString}
		sendEventToRoom(&clientEvent, user.Room)
		user.Room.Game.Reset()
	}
	updateGameForRoom(user.Room)
	log.Println("-GamePiecePlayed")
}

func GamePieceChosen(gamePieceChosen, username string) {
	log.Println("+GamePieceChosen")
	var gamePieceChosenDTO GamePieceChosenDTO
	err := json.Unmarshal([]byte(gamePieceChosen), &gamePieceChosenDTO)
	if err != nil {
		log.Println("GamePieceChosen: Couldn't unmarshal", gamePieceChosen, err)
		return
	}
	user, room, ok := validateUserOrRoom(username, "", "GamePieceChosen", true)
	if !ok {
		return
	}
	if room.Game.GameState == GameStatePlayerOneChoosing {
		if room.PlayerOne != user {
			sendInfoToUser(user, "You can't choose that piece, you're not player 1!")
			log.Println("GamePiecePlayed: User wasn't player one, can't choose piece now.")
			return
		}
		room.Game.GameState = GameStatePlayerTwoPlaying
	} else if room.Game.GameState == GameStatePlayerTwoChoosing {
		if room.PlayerTwo != user {
			sendInfoToUser(user, "You can't choose that piece, you're not player 2!")
			log.Println("GamePiecePlayed: User wasn't player two, can't choose piece now.")
			return
		}
		room.Game.GameState = GameStatePlayerOnePlaying
	} else {
		sendErrorToUser(user, "Are you doing something sneaky? You shouldn't have been able to send that request...")
		log.Println("GamePiecePlayed: Game not in proper state to play piece")
		return
	}
	if !Contains(room.Game.AvailablePieces, gamePieceChosenDTO.Piece) {
		sendErrorToUser(user, "That piece has already been played.")
		return
	}
	Remove(room.Game.AvailablePieces, gamePieceChosenDTO.Piece)
	room.Game.SelectedPiece = gamePieceChosenDTO.Piece
	updateGameForRoom(room)
	log.Println("-GamePieceChosen")
}

func playerChanged(username string) (*User, bool) {
	user, _, ok := validateUserOrRoom(username, "", "PlayerChanged", true)
	if !ok {
		return nil, false
	}
	user.Room.Game.Reset()
	updateGameForRoom(user.Room)
	return user, true
}

func updateGameForRoom(room *Room) {
	gameString, err := DtoToString(room.Game)
	if err != nil {
		log.Println("updateGameForRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.GameChange, gameString}
	sendEventToRoom(&clientEvent, room)
}

func RemoveFromObservers(user *User) {
	room := user.Room
	if room == nil {
		return
	}
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		if observer.Value.(*User).Username == user.Username {
			room.Observers.Remove(observer)
			break
		}
	}
}

// func sendEventToLobby(clientEvent *ClientEvent) {
// 	for _, user := range userMap {
// 		if user.Room == nil && user.Active {
// 			user.Events <- clientEvent
// 		}
// 	}
// }

func sendEventToRoom(clientEvent *ClientEvent, room *Room) {
	if room.PlayerOne != nil && room.PlayerOne.Active {
		room.PlayerOne.Events <- clientEvent
	}
	if room.PlayerTwo != nil && room.PlayerTwo.Active {
		room.PlayerTwo.Events <- clientEvent
	}
	for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
		if observer.Value.(*User).Active {
			observer.Value.(*User).Events <- clientEvent
		}
	}
}

func sendEventToUser(clientEvent *ClientEvent, user *User) {
	if user.Active {
		user.Events <- clientEvent
	}
}

func sendInfoToUser(user *User, message string) {
	if user == nil {
		return
	}
	messageEvent, err := DtoToString(InfoOrErrorMessageDTO{message})
	if err != nil {
		log.Println("sendErrorToUser: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.Info, messageEvent}
	sendEventToUser(&clientEvent, user)
}

func sendErrorToUser(user *User, message string) {
	if user == nil {
		return
	}
	messageEvent, err := DtoToString(InfoOrErrorMessageDTO{message})
	if err != nil {
		log.Println("sendErrorToUser: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{Config.Error, messageEvent}
	sendEventToUser(&clientEvent, user)
}

// validates whether a user or rooms exists
// if checkRoom is true, the functionality of the check depends on roomName
// if roomName == "", check if the user passed in is in a room
// if roomName != "", check if the room name given exists
// functionName is just for printing what method failed.
func validateUserOrRoom(username, roomName, functionName string, checkRoom bool) (*User, *Room, bool) {
	var user *User
	var room *Room
	user, ok := userMap[username]
	if !ok {
		log.Println(functionName + ": Couldn't find user")
		return nil, nil, false
	}
	if checkRoom {
		if roomName == "" {
			if user.Room != nil {
				room = user.Room
			} else {
				log.Println(functionName + ": User wasn't in a room")
				return nil, nil, false
			}
		} else {
			room, ok = roomMap[roomName]
			if !ok {
				log.Println(functionName + ": Couldn't find room")
				return nil, nil, false
			}
		}
	}
	return user, room, true
}
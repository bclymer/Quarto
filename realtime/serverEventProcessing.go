package realtime

import (
	"container/list"
	"encoding/json"
	"log"
	"github.com/bclymer/Quarto/constants"
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
		log.Println("AddUser: Couldn't unmarshal", addUserMessage)
		return nil
	}
	user := User{addUserDTO.Username, nil, make(chan *ClientEvent, 10)}

	userMap[addUserDTO.Username] = &user

	clientEvent := ClientEvent{constants.Config.UserAdd, addUserMessage}

	sendEventToLobby(&clientEvent)
	log.Println("-AddUser")
	return &user
}

func RemoveUser(removeUserMessage string) {
	log.Println("+RemoveUser")
	var removeUserDTO RemoveUserDTO
	err := json.Unmarshal([]byte(removeUserMessage), &removeUserDTO)
	if err != nil {
		log.Println("RemoveUser: Couldn't unmarshal", removeUserMessage)
		return
	}
	clientEvent := ClientEvent{constants.Config.UserRemove, removeUserMessage}

	user, ok := userMap[removeUserDTO.Username]
	if !ok {
		log.Println("Remove User: User didn't exist")
	}

	if user.Room != nil {
		LeaveRoom(removeUserDTO.Username)
	}
	delete(userMap, removeUserDTO.Username)
	sendEventToLobby(&clientEvent)
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
		sendErrorToUser(user, "")
		return
	}

	lobbyUserEvent := ClientEvent{constants.Config.UserRoomJoin, userRoomString}
	sendEventToLobby(&lobbyUserEvent)

	lobbyRoomString, err := DtoToString(MakeLobbyRoomDTO(room))
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	lobbyRoomEvent := ClientEvent{constants.Config.RoomChange, lobbyRoomString}
	sendEventToLobby(&lobbyRoomEvent)

	roomRoomString, err := DtoToString(MakeRoomRoomDTO(room))
	if err != nil {
		log.Println("JoinRoom: Couldn't marshal.", err)
		return
	}
	roomRoomEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
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
	clientEvent := ClientEvent{constants.Config.GameChange, gameString}
	sendEventToUser(&clientEvent, user)
}

func LeaveRoom(username string) {
	log.Println("+LeaveRoom")
	user, ok := userMap[username]
	if !ok {
		log.Println("LeaveRoom: User didn't exist")
		return
	}
	room := user.Room
	user.Room = nil
	if room == nil {
		log.Println("LeaveRoom: User wasn't in a room")
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
		clientEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
		sendEventToRoom(&clientEvent, room)
		room.UpdateGame()
		updateGameForRoom(room)
	}

	userRoomString, err := DtoToString(UserRoomDTO{username, room.Name})
	if err != nil {
		log.Println("LeaveRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.Config.UserRoomLeave, userRoomString}
	sendEventToLobby(&clientEvent)
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
		log.Println("AddRoom: Couldn't unmarshal", addRoomMessage)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Println("AddRoom: Couldn't find user who created the room")
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
	room := Room{nil, nil, list.New(), list.New(), addRoomDTO.Name, addRoomDTO.Private, addRoomDTO.Password, MakeNewGame()}
	roomMap[addRoomDTO.Name] = &room
	user.Room = &room

	lobbyRoomString, err := DtoToString(LobbyRoomDTO{addRoomDTO.Name, addRoomDTO.Private, 1})
	if err != nil {
		log.Println("AddRoom: Couldn't marshal.", err)
		return
	}
	addRoomClientEvent := ClientEvent{constants.Config.RoomAdd, lobbyRoomString}
	sendEventToLobby(&addRoomClientEvent)

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
		log.Println("Remove Room: Couldn't unmarshal", removeRoomMessage)
		return
	}

	delete(roomMap, removeRoomDTO.Name)

	clientEvent := ClientEvent{constants.Config.RoomRemove, removeRoomMessage}
	sendEventToLobby(&clientEvent)
	log.Println("-RemoveRoom")
}

func RoomNameChange(roomNameChanged string) { // RoomNameChangedDTO

}

func Chat(incomingChat, username string) {
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
		clientEvent := ClientEvent{constants.Config.Chat, outgoingChatString}
		sendEventToRoom(&clientEvent, user.Room)
		if user.Room.Events.Len() >= 10 {
			user.Room.Events.Remove(user.Room.Events.Front())
		}
		user.Room.Events.PushBack(&clientEvent)
	}
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
	clientEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
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
	clientEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
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
	clientEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
	user.Room.UpdateGame()
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
	clientEvent := ClientEvent{constants.Config.RoomChange, roomRoomString}
	user.Room.UpdateGame()
	updateGameForRoom(user.Room)
	sendEventToRoom(&clientEvent, user.Room)
	log.Println("-LeavePlayerTwo")
}

func GamePiecePlayed(gamePiecePlayed, username string) {
	log.Println("+GamePiecePlayed")
	var gamePiecePlayedDTO GamePiecePlayedDTO
	err := json.Unmarshal([]byte(gamePiecePlayed), &gamePiecePlayedDTO)
	if err != nil {
		log.Println("GamePiecePlayed: Couldn't unmarshal", gamePiecePlayed)
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
		gameWinnerString, err := DtoToString(GameWinnerDTO{winner})
		if err != nil {
			log.Println("GamePiecePlayed: Couldn't marshal.", err)
		}
		clientEvent := ClientEvent{constants.Config.GameWinner, gameWinnerString}
		sendEventToRoom(&clientEvent, user.Room)
	}
	updateGameForRoom(user.Room)
	log.Println("-GamePiecePlayed")
}

func GamePieceChosen(gamePieceChosen, username string) {
	log.Println("+GamePieceChosen")
	var gamePieceChosenDTO GamePieceChosenDTO
	err := json.Unmarshal([]byte(gamePieceChosen), &gamePieceChosenDTO)
	if err != nil {
		log.Println("GamePieceChosen: Couldn't unmarshal", gamePieceChosen)
		return
	}
	user, ok := userMap[username]
	if !ok {
		log.Println("GamePieceChosen: User didn't exist")
		return
	}
	if user.Room == nil {
		log.Println("GamePiecePlayed: User wasn't in room")
		return
	}
	if user.Room.Game.GameState == GameStatePlayerOneChoosing {
		if user.Room.PlayerOne != user {
			sendInfoToUser(user, "You can't choose that piece, you're not player 1!")
			log.Println("GamePiecePlayed: User wasn't player one, can't choose piece now.")
			return
		}
		user.Room.Game.GameState = GameStatePlayerTwoPlaying
	} else if user.Room.Game.GameState == GameStatePlayerTwoChoosing {
		if user.Room.PlayerTwo != user {
			sendInfoToUser(user, "You can't choose that piece, you're not player 2!")
			log.Println("GamePiecePlayed: User wasn't player two, can't choose piece now.")
			return
		}
		user.Room.Game.GameState = GameStatePlayerOnePlaying
	} else {
		sendErrorToUser(user, "Are you doing something sneaky? You shouldn't have been able to send that request...")
		log.Println("GamePiecePlayed: Game not in proper state to play piece")
		return
	}
	if !Contains(user.Room.Game.AvailablePieces, gamePieceChosenDTO.Piece) {
		sendErrorToUser(user, "That piece has already been played.")
		return
	}
	Remove(user.Room.Game.AvailablePieces, gamePieceChosenDTO.Piece)
	user.Room.Game.SelectedPiece = gamePieceChosenDTO.Piece
	updateGameForRoom(user.Room)
	log.Println("-GamePieceChosen")
}

func playerChanged(username string) (*User, bool) {
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
	updateGameForRoom(user.Room)
	return user, true
}

func updateGameForRoom(room *Room) {
	gameString, err := DtoToString(room.Game)
	if err != nil {
		log.Println("updateGameForRoom: Couldn't marshal.", err)
		return
	}
	clientEvent := ClientEvent{constants.Config.GameChange, gameString}
	sendEventToRoom(&clientEvent, room)
}

func checkForWinnerAndNotifyRoom(room *Room) {

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

func sendEventToLobby(clientEvent *ClientEvent) {
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
}

func sendEventToUser(clientEvent *ClientEvent, user *User) {
	user.Events <- clientEvent
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
	clientEvent := ClientEvent{constants.Config.Info, messageEvent}
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
	clientEvent := ClientEvent{constants.Config.Error, messageEvent}
	sendEventToUser(&clientEvent, user)
}

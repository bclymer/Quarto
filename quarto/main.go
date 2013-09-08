package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"quarto/constants"
	"quarto/realtime"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

type GeneratedUuid struct {
	Uuid string
}

type Success struct {
	Valid bool
}

func realtimeHost(ws *websocket.Conn) {
	username := ws.Request().URL.Query().Get("username")

	user := realtime.Subscribe(username)
	defer user.Cancel()

	newMessages := make(chan string, 10)
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				close(newMessages)
				return
			}
			newMessages <- msg
		}
	}()

	for {
		select {
		case event := <-user.Events:
			if websocket.JSON.Send(ws, &event) != nil {
				// They disconnected.
				return
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				return
			}

			var requestData realtime.ClientEvent
			json.Unmarshal([]byte(msg), &requestData)

			log.Println("Recieved Message", requestData)

			realtime.ServerSideAction(requestData, username)
		}
	}

	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../views/index.html")
}

func validateUsername(w http.ResponseWriter, r *http.Request) {
	log.Println("main.validateUsername")
	valid := realtime.ValidateUsername(r.FormValue("uuid"))
	test := Success{Valid: valid}
	response, _ := json.Marshal(test)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(response))
}

func rooms(w http.ResponseWriter, r *http.Request) {
	log.Println("main.rooms")
	roomMap := realtime.GetRoomMap()
	roomList := make([]realtime.LobbyRoomDTO, len(*roomMap))
	i := 0
	for _, room := range *roomMap {
		members := realtime.GetRoomUserCount(room)
		roomList[i] = realtime.LobbyRoomDTO{room.Name, room.Private, members}
		i++
	}
	serializedRooms, _ := json.Marshal(roomList)
	fmt.Fprint(w, string(serializedRooms))
}

func users(w http.ResponseWriter, r *http.Request) {
	log.Println("main.users")
	userMap := realtime.GetUserMap()
	userList := make([]realtime.LobbyUserDTO, len(*userMap))
	i := 0
	for _, user := range *userMap {
		roomName := ""
		if user.Room != nil {
			roomName = user.Room.Name
		}
		userList[i] = realtime.LobbyUserDTO{user.Username, roomName}
		i = i + 1
	}
	serializedUsers, _ := json.Marshal(userList)
	fmt.Fprint(w, string(serializedUsers))
}

type config struct {
	Config string
}

func configJs(w http.ResponseWriter, r *http.Request) {
	t := template.New("")
	t, _ = t.Parse(conficJs)
	constants.Init()
	constantsStr, _ := json.Marshal(constants.Config)
	w.Header().Set("Content-Type", "text/javascript")
	fmt.Fprint(w, strings.Replace(conficJs, "{{Config}}", string(constantsStr), -1))
}

const conficJs = `(function () {
	Quarto.constants = {{Config}};
})();
`

func test(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "cookie-1", Value: "one"})
	fmt.Fprint(w, "Yup")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/validate", validateUsername)
	http.HandleFunc("/rooms", rooms)
	http.HandleFunc("/users", users)
	http.HandleFunc("/test", test)
	http.HandleFunc("/js/constants.js", configJs)
	http.Handle("/realtime", websocket.Handler(realtimeHost))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css"))))
	http.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir("../views"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("../fonts"))))
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"quarto/realtime"
	"time"
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
	title := "Quarto Online!"
	p := loadPage(title)
	t, _ := template.ParseFiles("../views/index.html")
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, p)
}

func validateUsername(w http.ResponseWriter, r *http.Request) {
	log.Println("main.validateUsername")
	valid := realtime.ValidateUsername(r.FormValue("uuid"))
	test := Success{valid}
	response, _ := json.Marshal(test)
	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{"test", "tcookie", "/", "localhost:8080", expire, expire.Format(time.UnixDate), 86400, true, true, "test=tcookie", []string{"test=tcookie"}}
	http.SetCookie(w, &cookie)
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

func test(w http.ResponseWriter, r *http.Request) {
	test, _ := http.Get("http://www.bclymer.com")
	defer test.Body.Close()
	body, _ := ioutil.ReadAll(test.Body)
	fmt.Fprint(w, string(body))
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/validate", validateUsername)
	http.HandleFunc("/rooms", rooms)
	http.HandleFunc("/users", users)
	http.HandleFunc("/test", test)
	http.Handle("/realtime", websocket.Handler(realtimeHost))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css"))))
	http.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir("../views"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("../fonts"))))
	http.ListenAndServe(":8080", nil)
}

func loadPage(title string) *Page {
	filename := title + ".html"
	body, _ := ioutil.ReadFile(filename)
	return &Page{Title: title, Body: body}
}

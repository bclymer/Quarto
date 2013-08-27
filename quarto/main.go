package main

import (
	"fmt"
    "net/http"
    "html/template"
    "io/ioutil"
    "code.google.com/p/go.net/websocket"
    "encoding/json"
    "quarto/realtime"
    "quarto/admin"
    "time"
)

type Page struct {
    Title string
    Body  []byte
}

type MessageToWhom struct {
	ToUuid	string
}

type Success struct {
	Valid	bool
}

func realtimeHost(ws *websocket.Conn) {
	uuid := ws.Request().URL.Query().Get("uuid")
	subscription := realtime.Subscribe(uuid)
	subscription.Uuid = uuid;
	defer subscription.Cancel()

	realtime.Join(uuid)
	defer realtime.Leave(uuid)

	newMessages := make(chan string)
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
		case event := <-subscription.New:
			if websocket.JSON.Send(ws, &event) != nil {
				// They disconnected.
				return
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				return
			}

			var m MessageToWhom
			b := []byte(msg)
			json.Unmarshal(b, &m)

			realtime.Action("action", msg, uuid, m.ToUuid)
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
	valid := realtime.ValidateUsername(r.FormValue("uuid"))
	test := Success{valid}
	response, _ := json.Marshal(test)
	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{"test", "tcookie", "/", "localhost:8080", expire, expire.Format(time.UnixDate), 86400, true, true, "test=tcookie", []string{"test=tcookie"}}
	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(response))
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/validate", validateUsername)
    http.HandleFunc("/admin", admin.Index)
    http.Handle("/realtime", websocket.Handler(realtimeHost))
    http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js"))))
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css"))))
    http.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir("../views"))))
    http.ListenAndServe(":8080", nil)
}

func loadPage(title string) *Page {
    filename := title + ".html"
    body, _ := ioutil.ReadFile(filename)
    return &Page{Title: title, Body: body}
}
package main

import (
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goauth2/oauth"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"quarto/constants"
	"quarto/realtime"
	"strings"
	"text/template"
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
	if user == nil {
		return
	}
	defer user.Cancel()

	newMessages := make(chan string, 10)
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				user.Active = false
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
				user.Active = false
				return
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				user.Active = false
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
	valid, token := realtime.ValidateUsername(r.FormValue("username"))
	validDto := Success{Valid: valid}
	response, _ := json.Marshal(validDto)
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, &http.Cookie{Name: "quarto", Value: token})
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
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(serializedUsers))
}

type config struct {
	Config string
}

func configJs(w http.ResponseWriter, r *http.Request) {
	constants.Init()
	constantsStr, _ := json.Marshal(constants.Config)

	jsResponse := configJsConst

	cookie, err := r.Cookie("quarto")
	if err == nil {
		mongoUser := realtime.FindUser(cookie.Value)
		if mongoUser.Username != "" {
			jsResponse = strings.Replace(jsResponse, "{{Username}}", "\""+mongoUser.Username+"\"", -1)
		} else {
			jsResponse = strings.Replace(jsResponse, "{{Username}}", "undefined", -1)
		}
	} else {
		log.Println(err)
		jsResponse = strings.Replace(jsResponse, "{{Username}}", "undefined", -1)
	}

	w.Header().Set("Content-Type", "text/javascript")
	fmt.Fprint(w, strings.Replace(jsResponse, "{{Config}}", string(constantsStr), -1))
}

const configJsConst = `(function () {
	Quarto.constants = {{Config}};
	Quarto.username = {{Username}};
})();
`

var oauthCfg = &oauth.Config{}

func main() {
	session := realtime.ConnectMongo()
	defer session.Close()

	mongoOauth := realtime.FetchOauth()
	oauthCfg = &oauth.Config{
		ClientId:     mongoOauth.ClientId,
		ClientSecret: mongoOauth.ClientSecret,
		RedirectURL:  mongoOauth.RedirectURL,
		Scope:        mongoOauth.Scope,
		AuthURL:      mongoOauth.AuthURL,
		TokenURL:     mongoOauth.TokenURL,
		TokenCache:   oauth.CacheFile("cache.json"),
	}

	redis := realtime.ConnectRedis()
	defer redis.Close()

	http.HandleFunc("/", handler)
	http.HandleFunc("/validate", validateUsername)
	http.HandleFunc("/rooms", rooms)
	http.HandleFunc("/users", users)
	http.HandleFunc("/js/constants.js", configJs)

	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	http.HandleFunc("/authorize", handleAuthorize)
	http.HandleFunc("/login", login)

	http.Handle("/realtime", websocket.Handler(realtimeHost))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css"))))
	http.Handle("/views/", http.StripPrefix("/views/", http.FileServer(http.Dir("../views"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("../fonts"))))
	http.ListenAndServe(":8080", nil)
}

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo"
const port = ":8080"

func login(w http.ResponseWriter, r *http.Request) {
	notAuthenticatedTemplate.Execute(w, nil)
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	//Get the Google URL which shows the Authentication page to the user
	url := oauthCfg.AuthCodeURL("")

	// redirect user to that page
	http.Redirect(w, r, url, http.StatusFound)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	//Get the code from the response
	code := r.FormValue("code")

	t := &oauth.Transport{Config: oauthCfg}

	tok, err := oauthCfg.TokenCache.Token()
	if err != nil {
		tok, err = t.Exchange(code)
		if err != nil {
			log.Fatal("Exchange:", err)
		}
		{
			err := oauthCfg.TokenCache.PutToken(tok)
			if err != nil {
				log.Fatal("Cache write:", err)
			}
			log.Printf("Token is cached in %v\n", oauthCfg.TokenCache)
		}

		// Skip TLS Verify
		t.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	t.Token = tok

	// Make the request.
	req, err := t.Client().Get(profileInfoURL)
	if err != nil {
		log.Fatal("Request Error:", err)
	}
	defer req.Body.Close()

	body, _ := ioutil.ReadAll(req.Body)

	log.Println(string(body))
	fmt.Fprintf(w, string(body))
}

var notAuthenticatedTemplate = template.Must(template.New("").Parse(`
<html><body>
You have currently not given permissions to access your data. Please authenticate this app with the Google OAuth provider.
<form action="/authorize" method="POST"><input type="submit" value="Ok, authorize this app with my id"/></form>
</body></html>
`))

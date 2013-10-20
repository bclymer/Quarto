package quarto

import (
	"bclymer/quarto/quarto"
	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goauth2/oauth"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"menteslibres.net/gosexy/redis"
	"net/http"
	"strings"
	"text/template"
)

type Success struct {
	Valid bool `json:"valid"`
}

func realtimeHost(ws *websocket.Conn) {
	username := ws.Request().URL.Query().Get("username")

	user := quarto.Subscribe(username)
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

			var requestData quarto.ClientEvent
			json.Unmarshal([]byte(msg), &requestData)

			log.Println("Recieved Message", requestData)

			quarto.ServerSideAction(requestData, username)
		}
	}
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "quarto/static/views/index.html")
}

func validateUsername(w http.ResponseWriter, r *http.Request) {
	log.Println("main.validateUsername")
	valid, token := quarto.ValidateUsername(r.FormValue("username"))
	validDto := Success{Valid: valid}
	response, _ := json.Marshal(validDto)
	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, &http.Cookie{Name: "quarto", Value: token})
	fmt.Fprint(w, string(response))
}

func rooms(w http.ResponseWriter, r *http.Request) {
	log.Println("main.rooms")
	roomMap := quarto.GetRoomMap()
	roomList := make([]quarto.LobbyRoomDTO, len(*roomMap))
	i := 0
	for _, room := range *roomMap {
		members := quarto.GetRoomUserCount(room)
		roomList[i] = quarto.LobbyRoomDTO{room.Name, room.Private, members}
		i++
	}
	serializedRooms, _ := json.Marshal(roomList)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(serializedRooms))
}

func users(w http.ResponseWriter, r *http.Request) {
	log.Println("main.users")
	var serializedUsers string
	serializedUsers, err := quarto.RedisGet("users")
	if err != nil {
		log.Println("Cache Miss", err)
		userMap := quarto.GetUserMap()
		serializedUsersByteArray, _ := json.Marshal(userMap)
		serializedUsers = string(serializedUsersByteArray)
		quarto.RedisPut("users", serializedUsers)
	} else {
		log.Println("Cache Hit")
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, serializedUsers)
}

type config struct {
	Config string
}

func configJs(w http.ResponseWriter, r *http.Request) {
	constantsStr, _ := json.Marshal(quarto.Config)

	jsResponse := configJsConst

	// cookie, err := r.Cookie("quarto")
	// if err == nil {
	// 	mongoUser := FindUser(cookie.Value)
	// 	if mongoUser.Username != "" {
	// 		jsResponse = strings.Replace(jsResponse, "{{Username}}", "\""+mongoUser.Username+"\"", -1)
	// 	} else {
	// 		jsResponse = strings.Replace(jsResponse, "{{Username}}", "undefined", -1)
	// 	}
	// } else {
	// 	log.Println(err)
	jsResponse = strings.Replace(jsResponse, "{{Username}}", "undefined", -1)
	// }

	w.Header().Set("Content-Type", "text/javascript")
	fmt.Fprint(w, strings.Replace(jsResponse, "{{Config}}", string(constantsStr), -1))
}

const configJsConst = `(function () {
	Quarto.constants = {{Config}};
	Quarto.username = {{Username}};
})();
`

var oauthCfg = &oauth.Config{}

func StartServer(urlPrefix string) *redis.Client {
	quarto.InitConstants()
	log.Println("Connecting to Mongo")
	mongo := quarto.ConnectMongo()
	defer mongo.Close()
	log.Println("Mongo - Success")

	mongoOauth := quarto.FetchOauth()
	oauthCfg = &oauth.Config{
		ClientId:     mongoOauth.ClientId,
		ClientSecret: mongoOauth.ClientSecret,
		RedirectURL:  mongoOauth.RedirectURL,
		Scope:        mongoOauth.Scope,
		AuthURL:      mongoOauth.AuthURL,
		TokenURL:     mongoOauth.TokenURL,
		TokenCache:   oauth.CacheFile("cache.json"),
	}

	log.Println("Connecting to Redis")
	redis := quarto.ConnectRedis()
	log.Println("Redis - Success")

	if urlPrefix != "" {
		urlPrefix = "/" + urlPrefix
	}
	http.HandleFunc(urlPrefix+"/", handler)
	http.HandleFunc(urlPrefix+"/validate", validateUsername)
	http.HandleFunc(urlPrefix+"/rooms", rooms)
	http.HandleFunc(urlPrefix+"/users", users)
	http.HandleFunc(urlPrefix+"/static/js/constants.js", configJs)

	http.HandleFunc(urlPrefix+"/oauth2callback", handleOAuth2Callback)
	http.HandleFunc(urlPrefix+"/authorize", handleAuthorize)
	http.HandleFunc(urlPrefix+"/login", login)

	http.Handle(urlPrefix+"/realtime", websocket.Handler(realtimeHost))

	http.Handle(urlPrefix+"/static/", http.StripPrefix(urlPrefix+"/static", http.FileServer(http.Dir("quarto/static"))))
	log.Println("Quarto is running...")
	return redis
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

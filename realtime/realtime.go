package realtime

import (
	"log"
	"container/list"
)

type Room struct {
	Player1		*User
	Player2		*User
	Observers	*list.List
	History		*list.List
	Messages	*list.List
	Name		string
	Private		bool
	Password	string // only used if the room is private
}

type User struct {
	Username	string // selected username
	Uuid		string // uuid returned by the server
	Room		*Room // room the user is in.
	Events		chan Event // event channel to send messages to user
}

type Event struct {
	Action		string // any function to call on client side
	Data		string // json string
	Uuid		string // uuid of who is sending the message
	Room		string // room that the message is going to. The server should verify that the uuid sender is in the room
}

// Owner must cancel itself when they stop listening to events.
func (u User) Cancel() {
	var roomName string
	if (u.Room != nil) {
		roomName = u.Room.Name
	} else {
		roomName = ""
	}
	publish <- Event{ "left", "", u.Uuid, roomName }
	unsubscribe <- u // Unsubscribe the channel.
}

func Subscribe(uuid, username string) User {
	user := User { username, uuid, nil, make(chan Event, 10) }
	subscribe <- user
	return user
}

func Action(action string, data string, uuid string, toUuid string) {
	publish <- Event{ action, data, uuid, toUuid }
}

const archiveSize = 10

var (
	// Send a channel here to get room events back.  It will send the entire
	// archive initially, and then new messages as they come in.
	subscribe = make(chan User, 10)
	// Send a channel here to unsubscribe.
	unsubscribe = make(chan User, 10)
	// Send events here to publish them.
	publish = make(chan Event, 10)
	// 
	checkUsername = make(chan CheckUsername, 10)
	//
	getAllUsers = make(chan (chan string), 10)
)

type CheckUsername struct {
	uuid string
	valid chan bool
}

func ValidateUsername(name string) bool {
	check := CheckUsername{name, make(chan bool)}
	checkUsername <- check
	return <-check.valid
}

func GetAllUsers() []string {
	names := make([]string, 0)
	c := make(chan string)
	getAllUsers <- c
	i := 0
	for uuid := range c {
		log.Print("Got user ", uuid)
		names[i] =  uuid
		i = i + 1
	}
	return names
}

// This function loops forever, handling the chat room pubsub
func realtime() {

	subscribers := make(map[string]*User)
	rooms := make(map[string]*Room)


	for {
		select {
		case newUser := <-subscribe:
			subscribers[newUser.Uuid] = &newUser
		case event := <-publish:

			room := rooms[event.Room]
			if (room.Player1 != nil) {
				room.Player1.Events <- event
			}
			if (room.Player2 != nil) {
				room.Player2.Events <- event
			}
			for observer := room.Observers.Front(); observer != nil; observer = observer.Next() {
				observer.Value.(User).Events <- event
			}
		case unsub := <-unsubscribe:
			for uuid, _ := range subscribers {
				if (uuid == unsub.Uuid) {
					delete(subscribers, uuid)
				}
			}
		case check := <-checkUsername:
			check.valid <- subscribers[check.uuid] == nil
		case allUsers := <- getAllUsers:
			for uuid, _ := range subscribers {
				allUsers <- uuid
				log.Print("Reporting user ", uuid)
			}
			close(allUsers)
		}
	}
}


func init() {
	go realtime()
}

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
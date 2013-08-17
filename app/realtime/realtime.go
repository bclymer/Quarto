package realtime

import (
	"container/list"
	"github.com/robfig/revel"
)

type Event struct {
	Action		string // "joined", "left", or "action"
	Data		string
	Uuid		string
	ToUuid		string
}

type Subscription struct {
	New     <-chan Event // New events coming in.
	Uuid	string
}

// Owner of a subscription must cancel it when they stop listening to events.
func (s Subscription) Cancel() {
	unsubscribe <- s.New // Unsubscribe the channel.
	drain(s.New)         // Drain it, just in case there was a pending publish.
}

func newEvent(action string, data string, uuid string, toUuid string) Event {
	return Event{action, data, uuid, toUuid}
}

func Subscribe() Subscription {
	resp := make(chan Subscription)
	subscribe <- resp
	revel.INFO.Printf("resp ", resp)
	return <-resp
}

func Join(uuid string) {
	publish <- newEvent("joined", "", uuid, "")
}

func Action(action string, data string, uuid string, toUuid string) {
	publish <- newEvent(action, data, uuid, toUuid)
}

func Leave(uuid string) {
	publish <- newEvent("left", "", uuid, "")
}

const archiveSize = 10

var (
	// Send a channel here to get room events back.  It will send the entire
	// archive initially, and then new messages as they come in.
	subscribe = make(chan (chan<- Subscription), 10)
	// Send a channel here to unsubscribe.
	unsubscribe = make(chan (<-chan Event), 10)
	// Send events here to publish them.
	publish = make(chan Event, 10)
)

// This function loops forever, handling the chat room pubsub
func realtime() {
	subscribers := list.New()

	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber, ""}
		case event := <-publish:
			// TODO: only send to channel of intended client, interating through sucks
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
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
package realtime

import (
)

type Event struct {
	Action		string // any function to call on client side
	Data		string // json string
	Uuid		string // uuid of who is sending the message
	ToUuid		string // uuid of whom the message is going to
}

type Subscription struct {
	New     <-chan Event // New events coming in.
	Uuid	string
}

type NewSubscription struct {
	SubChan	chan Subscription
	Uuid	string
}

// Owner of a subscription must cancel it when they stop listening to events.
func (s Subscription) Cancel() {
	unsubscribe <- s // Unsubscribe the channel.
	drain(s.New)         // Drain it, just in case there was a pending publish.
}

func newEvent(action string, data string, uuid string, toUuid string) Event {
	return Event{action, data, uuid, toUuid}
}

func Subscribe(uuid string) Subscription {
	resp := NewSubscription{ make(chan Subscription), uuid }
	subscribe <- resp
	return <-resp.SubChan
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
	subscribe = make(chan NewSubscription, 10)
	// Send a channel here to unsubscribe.
	unsubscribe = make(chan Subscription, 10)
	// Send events here to publish them.
	publish = make(chan Event, 10)
	// 
	checkUsername = make(chan CheckUsername, 10)
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

// This function loops forever, handling the chat room pubsub
func realtime() {
	subscribers := make(map[string]chan Event)
	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 10)
			subscribers[ch.Uuid] = subscriber
			ch.SubChan <- Subscription{subscriber, ch.Uuid}
		case event := <-publish:
			// TODO: only send to channel of intended client, interating through sucks
			if (event.ToUuid == "") {
				for _, ch := range subscribers {
					ch <- event
				}
			} else {
				subscribers[event.ToUuid] <- event
			}
		case unsub := <-unsubscribe:
			for uuid, _ := range subscribers {
				if (uuid == unsub.Uuid) {
					delete(subscribers, uuid)
				}
			}
		case check := <-checkUsername:
			check.valid <- subscribers[check.uuid] == nil
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
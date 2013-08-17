package controllers

import "github.com/robfig/revel"
import "code.google.com/p/go.net/websocket"
import "quarto/app/realtime"

type App struct {
	*revel.Controller
}

type TestDesc struct {
	PieceId int
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Realtime(uuid string, ws *websocket.Conn) revel.Result {

	subscription := realtime.Subscribe()
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
				return nil
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				return nil
			}

			// Otherwise, say something.
			realtime.Action("action", msg, uuid)
		}
	}

	return nil
}
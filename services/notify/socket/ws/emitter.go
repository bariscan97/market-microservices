package ws

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID]*Client
	Register   chan *Client
	Unregister chan *Client
	Emitter    chan *Message
}

type Message struct {
	CustomerID *uuid.UUID
	Content    string
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Emitter:    make(chan *Message, 5),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case msg := <-h.Emitter:
			switch reflect.TypeOf(msg.CustomerID) {
			case reflect.TypeOf(&uuid.UUID{}):
				if cl, ok := h.Clients[*msg.CustomerID]; ok {
					if ok {
						cl.Message <- msg.Content
					}
				}
			default:
				fmt.Print(msg.CustomerID)
				for _, cl := range h.Clients {
					cl.Message <- msg.Content
				}
			}
		case cl := <-h.Register:
			h.Clients[cl.CustomerID] = cl
			fmt.Printf("current users:%d", len(h.Clients))
		case cl := <-h.Unregister:
			close(cl.Message)
			delete(h.Clients, cl.CustomerID)
		}
	}
}

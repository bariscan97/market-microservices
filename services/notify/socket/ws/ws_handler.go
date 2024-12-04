package ws

import (
	"net/http"
    "github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *Hub
}

func NewHandler(h *Hub) *Handler {
	return &Handler{
		hub: h,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func (h *Handler) JoinWs(w http.ResponseWriter, r *http.Request) {
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	customerID, err := uuid.Parse(r.Header.Get("X-User-ID"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		conn.Close()
		return
	}

	cl := &Client{
		CustomerID: customerID,
		Conn:     conn,
		Message: make(chan string ,5),
	}

	h.hub.Register <- cl
	go cl.writeMessage()
	cl.readMessage(h.hub)
}


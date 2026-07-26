package ws

import (
	"encoding/json"
	"log"
	"sync"
)

const EventDoneTasksChanged = "doneTasksChanged"

// Message is a lightweight server→client event.
type Message struct {
	Type string `json:"type"`
}

// Hub maintains the set of active websocket clients and broadcasts events to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 64),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run processes register/unregister/broadcast until the process exits.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			var stale []*Client
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					stale = append(stale, client)
				}
			}
			h.mu.RUnlock()
			for _, client := range stale {
				h.unregister <- client
			}
		}
	}
}

func (h *Hub) BroadcastDoneTasksChanged() {
	payload, err := json.Marshal(Message{Type: EventDoneTasksChanged})
	if err != nil {
		log.Printf("ws: marshal doneTasksChanged: %v", err)
		return
	}
	select {
	case h.broadcast <- payload:
	default:
		log.Printf("ws: broadcast channel full, dropping doneTasksChanged")
	}
}

package websocket

import (
	"fmt"
	"log"
	"net/http"
	"proxy-interceptor/config"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests
	register chan *Client

	// Unregister requests
	unregister chan *Client
}

// Client represent a client connection
type Client struct {
	hub *Hub

	// Websocket connection
	conn *websocket.Conn

	// Outbound messages
	send chan []byte
}

var hub = Hub{
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var BroadcastChannel = make(chan []byte)

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-BroadcastChannel:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func serveWebsocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
}

func Start() {
	go hub.run()
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWebsocket(&hub, w, r)
	})
	go func() {
		cfg := config.GetInstance()
		addr := fmt.Sprintf("127.0.0.1:%d", cfg.WebSocketPort)
		log.Printf("WebSocket server on %s", addr)
		log.Fatal(http.ListenAndServe(addr, wsMux))
	}()
}

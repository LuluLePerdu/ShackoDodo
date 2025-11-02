package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proxy-interceptor/config"
	"sync"
	"time"

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
var ModifyChannel = make(chan ModifyRequestData, 100) // Buffer de 100 modifications
var PendingModifications = make(map[string]ModifyRequestData)
var PendingRequests = make(map[string]chan ModifyRequestData) // Requêtes en attente de modification
var modifyMutex sync.RWMutex
var requestMutex sync.RWMutex

func (h *Hub) run() {
	log.Printf("WebSocket Hub started, waiting for clients...")
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("✅ New WebSocket client connected. Total clients: %d", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("❌ WebSocket client disconnected. Total clients: %d", len(h.clients))
			}
		case message := <-BroadcastChannel:
			log.Printf("📡 Broadcasting message to %d clients", len(h.clients))
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

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		switch msg.Type {
		case "pause":
			if pause, ok := msg.Data.(bool); ok {
				config.GetInstance().SetPause(pause)
				log.Printf("Set pause to %v", pause)
			} else {
				log.Printf("Invalid data for pause type: %v", msg.Data)
			}
		case "modify_request":
			// Traitement des modifications de requête
			if modifyData, ok := msg.Data.(map[string]interface{}); ok {
				modify := ModifyRequestData{
					ID:     getString(modifyData, "id"),
					Method: getString(modifyData, "method"),
					URL:    getString(modifyData, "url"),
					Body:   getString(modifyData, "body"),
					Action: getString(modifyData, "action"),
				}
				// Conversion des headers
				if headersData, exists := modifyData["headers"]; exists {
					modify.Headers = make(map[string][]string)

					switch h := headersData.(type) {
					case map[string]interface{}:
						// Headers comme objet JSON
						for k, v := range h {
							switch val := v.(type) {
							case string:
								modify.Headers[k] = []string{val}
							case []interface{}:
								var strSlice []string
								for _, item := range val {
									if str, ok := item.(string); ok {
										strSlice = append(strSlice, str)
									}
								}
								if len(strSlice) > 0 {
									modify.Headers[k] = strSlice
								}
							}
						}
					case string:
						// Headers comme string JSON, essayer de parser
						var headerMap map[string]interface{}
						if err := json.Unmarshal([]byte(h), &headerMap); err == nil {
							for k, v := range headerMap {
								if strValue, ok := v.(string); ok {
									modify.Headers[k] = []string{strValue}
								}
							}
						}
					}
				}

				// Stocker la modification dans la map globale
				modifyMutex.Lock()
				PendingModifications[modify.ID] = modify
				modifyMutex.Unlock()

				// Vérifier s'il y a une requête en attente pour cet ID
				requestMutex.Lock()
				if waitChan, exists := PendingRequests[modify.ID]; exists {
					// Envoyer la modification à la requête en attente
					select {
					case waitChan <- modify:
						log.Printf("✅ Modification envoyée à la requête en attente %s", modify.ID)
					default:
						log.Printf("⚠️ Channel bloqué pour la requête %s", modify.ID)
					}
					delete(PendingRequests, modify.ID)
				}
				requestMutex.Unlock()

				// Essayer d'envoyer dans le channel (non-bloquant)
				select {
				case ModifyChannel <- modify:
					// Envoyé avec succès
				default:
					// Channel plein, on ignore (la modification est déjà dans la map)
				}
				log.Printf("Received modification for request %s: %s %s (action: %s)", modify.ID, modify.Method, modify.URL, modify.Action)
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
	log.Printf("🔗 WebSocket connection attempt from %s", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("✅ WebSocket upgraded successfully for %s", r.RemoteAddr)

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func WaitForModification(id string, timeout time.Duration) (ModifyRequestData, bool) {
	// Créer un channel pour cette requête
	waitChan := make(chan ModifyRequestData, 1)

	requestMutex.Lock()
	PendingRequests[id] = waitChan
	requestMutex.Unlock()

	// Attendre la modification ou le timeout
	select {
	case modification := <-waitChan:
		return modification, true
	case <-time.After(timeout):
		// Nettoyer si timeout
		requestMutex.Lock()
		delete(PendingRequests, id)
		requestMutex.Unlock()
		return ModifyRequestData{}, false
	}
}

func getString(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
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

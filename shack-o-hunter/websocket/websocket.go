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

// Fonction pour lancer un navigateur depuis l'UI
func launchBrowserFromUI(browserName string) {
	// Import dynamique du module browsers pour éviter les dépendances circulaires
	// On utilisera une interface ou un channel pour communiquer avec le module browsers
	BrowserLaunchChannel <- BrowserLaunchRequest{
		Browser: browserName,
	}
}

type BrowserLaunchRequest struct {
	Browser string
}

var BrowserLaunchChannel = make(chan BrowserLaunchRequest, 10)

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
var ModifyChannel = make(chan RequestData, 100)
var PendingModifications = make(map[string]RequestData)
var PendingRequests = make(map[string]chan RequestData)
var modifyMutex sync.RWMutex
var requestMutex sync.RWMutex

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

				// Si on désactive la pause, envoyer toutes les requêtes pending
				if !pause {
					go ResumePendingRequests()
				}
			} else {
				log.Printf("Invalid data for pause type: %v", msg.Data)
			}
		case "launch_browser":
			// Lancement d'un navigateur spécifique
			if browserData, ok := msg.Data.(map[string]interface{}); ok {
				browserName := getString(browserData, "browser")
				log.Printf("Received browser launch request: %s", browserName)

				// Déléguer le lancement au module browsers
				go launchBrowserFromUI(browserName)
			}
		case "resume_all":
			// Envoyer toutes les requêtes en attente
			go ResumePendingRequests()
		case "modify_request":
			if modifyData, ok := msg.Data.(map[string]interface{}); ok {
				modify := RequestData{
					Method: getString(modifyData, "method"),
					URL:    getString(modifyData, "url"),
					Body:   getString(modifyData, "body"),
					Action: getString(modifyData, "action"),
				}

				if headersData, exists := modifyData["headers"]; exists {
					modify.Headers = make(map[string][]string)

					switch h := headersData.(type) {
					case map[string]interface{}:
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

				requestID := msg.ID

				modifyMutex.Lock()
				PendingModifications[requestID] = modify
				modifyMutex.Unlock()

				requestMutex.Lock()
				if waitChan, exists := PendingRequests[requestID]; exists {
					select {
					case waitChan <- modify:
					default:
					}
					delete(PendingRequests, requestID)
				}
				requestMutex.Unlock()

				select {
				case ModifyChannel <- modify:
				default:
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
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func WaitForModification(id string, timeout time.Duration) (RequestData, bool) {
	waitChan := make(chan RequestData, 1)

	requestMutex.Lock()
	PendingRequests[id] = waitChan
	requestMutex.Unlock()

	select {
	case modification := <-waitChan:
		return modification, true
	case <-time.After(timeout):
		requestMutex.Lock()
		delete(PendingRequests, id)
		requestMutex.Unlock()
		return RequestData{}, false
	}
}

func ResumePendingRequests() {
	requestMutex.Lock()
	defer requestMutex.Unlock()

	count := 0
	for id, waitChan := range PendingRequests {
		// Créer une requête de "send" automatique pour chaque requête en attente
		autoSend := RequestData{
			Action: "send",
		}

		select {
		case waitChan <- autoSend:
			count++
		default:
			// Si le channel est fermé ou bloqué, on ignore
		}
		delete(PendingRequests, id)
	}

	if count > 0 {
		log.Printf("Resumed %d pending requests", count)
	}
}

func StorePendingModification(id string, modification RequestData) {
	modifyMutex.Lock()
	defer modifyMutex.Unlock()
	PendingModifications[id] = modification
}

func GetModificationForID(id string) (RequestData, bool) {
	modifyMutex.Lock()
	defer modifyMutex.Unlock()

	modification, exists := PendingModifications[id]
	if exists {
		delete(PendingModifications, id)
	}
	return modification, exists
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

package websocket

type Message struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Data any    `json:"data"`
}

type RequestData struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
	Status  string              `json:"status,omitempty"` // "pending", "sent", "dropped"
	Action  string              `json:"action,omitempty"` // "send", "drop"
}

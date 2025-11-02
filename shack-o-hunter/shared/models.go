package shared

// RequestData represents the data of an HTTP request
type RequestData struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// Message represents a WebSocket message
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

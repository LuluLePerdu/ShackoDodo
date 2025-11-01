package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"proxy-interceptor/websocket"
	"time"
)

type RequestData struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

var directClient = &http.Client{
	Transport: &http.Transport{
		Proxy: nil, // IMPORTANT: never proxy the outbound (avoid loops)
	},
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Only plain HTTP requests (absolute-form in proxy mode)
	if r.URL.Scheme != "http" {
		http.Error(w, "only http (no https/connect)", http.StatusBadRequest)
		return
	}

	// Read body so we can print/modify/forward
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	// Hard-coded post update
	if r.URL.String() == "http://localhost:5000/login" {
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			// Parse the form data
			form, err := url.ParseQuery(string(body))
			if err == nil {
				// Check if username and password fields exist
				if form.Has("username") && form.Has("password") {
					// Hardcode the username and password
					time.Sleep(3 * time.Second)
					form.Set("username", "admin")
					form.Set("password", "password123")
					// Encode the form back to a string and update the body
					body = []byte(form.Encode())
				}
			}
		}
	}

	// Log
	requestData := RequestData{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: r.Header,
		Body:    string(body),
	}

	jsonData, err := json.MarshalIndent(requestData, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Println("===== REQUETE =====")
		fmt.Println(string(jsonData))
		websocket.RequestChannel <- jsonData
	}

	// Clean hop-by-hop/proxy headers
	cleanHeaders := r.Header.Clone()
	cleanHeaders.Del("Proxy-Connection")
	cleanHeaders.Del("Connection")
	cleanHeaders.Del("Keep-Alive")
	cleanHeaders.Del("Transfer-Encoding")
	cleanHeaders.Del("TE")
	cleanHeaders.Del("Trailer")
	cleanHeaders.Del("Upgrade")

	// Rebuild & forward directly to origin
	req, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = cleanHeaders

	resp, err := directClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Return response to browser
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func Start() {
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", HandleRequest)
	go func() {
		log.Println("Proxy HTTP on 127.0.0.1:8181")
		log.Fatal(http.ListenAndServe("127.0.0.1:8181", proxyMux))
	}()
}

package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"proxy-interceptor/cert"
	"proxy-interceptor/config"
	"proxy-interceptor/websocket"
	"strings"
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
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // Don't follow redirects
	},
}

// shouldFilterDomain returns true if the domain should be filtered (not logged/sent)
func shouldFilterDomain(host string) bool {
	host = strings.ToLower(host)

	mozillaFirefoxDomains := []string{
		"mozilla.com",
		"mozilla.org",
		"mozilla.net",
		"firefox.com",
		"firefox.org",
		"getpocket.com",
		"firefoxusercontent.com",
		"services.mozilla.com",
	}

	for _, domain := range mozillaFirefoxDomains {
		if strings.Contains(host, domain) {
			return true
		}
	}

	return false
}

// handleConnection handles each incoming connection
func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Read the first request
	reader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("Erreur lors de la lecture de la requête: %v", err)
		return
	}

	// Check if domain should be filtered before logging
	host := req.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	shouldFilter := shouldFilterDomain(host)

	// Log the request only if not filtered
	if !shouldFilter {
		log.Printf("Requête: %s %s %s", req.Method, req.Host, req.URL.Path)
	}

	// Handle CONNECT method for HTTPS tunneling
	if req.Method == http.MethodConnect {
		handleHTTPS(clientConn, req)
		return
	}

	// Handle regular HTTP requests
	handleHTTP(clientConn, req)
}

// handleHTTPS handles HTTPS CONNECT requests with MITM interception
func handleHTTPS(clientConn net.Conn, req *http.Request) {
	host := req.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Filter Mozilla/Firefox domains
	shouldFilter := shouldFilterDomain(host)

	if !shouldFilter {
		log.Printf("HTTPS CONNECT: %s", req.Host)
	}

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	tlsCert, err := cert.GenerateCertForHost(host)
	if err != nil {
		log.Printf("Erreur génération certificat pour %s: %v", host, err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	tlsClientConn := tls.Server(clientConn, tlsConfig)
	if err := tlsClientConn.Handshake(); err != nil {
		log.Printf("Erreur TLS handshake: %v", err)
		return
	}
	defer tlsClientConn.Close()

	if !shouldFilter {
		log.Printf("Interception HTTPS établie pour: %s", req.Host)
	}

	reader := bufio.NewReader(tlsClientConn)
	httpsReq, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("Erreur lecture requête HTTPS: %v", err)
		return
	}

	var body []byte
	if httpsReq.Body != nil {
		body, _ = io.ReadAll(httpsReq.Body)
		httpsReq.Body.Close()
	}

	fullURL := "https://" + req.Host + httpsReq.URL.Path
	if httpsReq.URL.RawQuery != "" {
		fullURL += "?" + httpsReq.URL.RawQuery
	}

	// Only log and send data if not from Mozilla/Firefox
	if !shouldFilter {
		requestData := RequestData{
			Method:  httpsReq.Method,
			URL:     fullURL,
			Headers: httpsReq.Header,
			Body:    string(body),
		}

		jsonData, err := json.MarshalIndent(requestData, "", "  ")
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
		} else {
			fmt.Println("===== REQUETE HTTPS =====")
			fmt.Println(string(jsonData))
			fmt.Println()
		}
	}

	proxyReq, err := http.NewRequest(httpsReq.Method, fullURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("Erreur création requête: %v", err)
		return
	}

	proxyReq.Header = httpsReq.Header.Clone()
	proxyReq.Header.Del("Proxy-Connection")
	proxyReq.Header.Del("Connection")

	resp, err := directClient.Do(proxyReq)
	if err != nil {
		log.Printf("Erreur envoi requête: %v", err)
		tlsClientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer resp.Body.Close()

	if !shouldFilter {
		log.Printf("Réponse HTTPS: %d %s (Host: %s)", resp.StatusCode, resp.Status, req.Host)
	}

	tlsClientConn.Write([]byte(fmt.Sprintf("HTTP/%d.%d %d %s\r\n",
		resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)))

	for key, values := range resp.Header {
		for _, value := range values {
			tlsClientConn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		}
	}
	tlsClientConn.Write([]byte("\r\n"))

	written, _ := io.Copy(tlsClientConn, resp.Body)
	if !shouldFilter {
		log.Printf("Body HTTPS transféré: %d bytes", written)
	}
}

// handleHTTP handles regular HTTP requests
func handleHTTP(clientConn net.Conn, req *http.Request) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	if req.URL.Scheme == "" {
		if req.TLS != nil {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}
	}
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	// Check if domain should be filtered
	shouldFilter := shouldFilterDomain(req.Host)

	if !shouldFilter {
		log.Printf("HTTP Request: %s %s", req.Method, req.URL.String())
	}

	if req.URL.String() == "http://localhost:5000/login" {
		if req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			form, err := url.ParseQuery(string(body))
			if err == nil {
				if form.Has("username") && form.Has("password") {
					time.Sleep(3 * time.Second)
					form.Set("username", "admin")
					form.Set("password", "password123")
					body = []byte(form.Encode())
				}
			}
		}
	}

	// Only log and send data if not from Mozilla/Firefox
	if !shouldFilter {
		requestData := RequestData{
			Method:  req.Method,
			URL:     req.URL.String(),
			Headers: req.Header,
			Body:    string(body),
		}

		message := websocket.Message{
			Type: "http_request",
			Data: requestData,
		}

		jsonData, err := json.MarshalIndent(message, "", "  ")
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
		} else {
			fmt.Println("===== REQUETE HTTP =====")
			fmt.Println(string(jsonData))
			websocket.BroadcastChannel <- jsonData
		}
	}

	proxyReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(body))
	if err != nil {
		log.Printf("Erreur lors de la création de la requête: %v", err)
		clientConn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	proxyReq.Header = req.Header.Clone()
	proxyReq.Header.Del("Proxy-Connection")
	proxyReq.Header.Del("Connection")

	resp, err := directClient.Do(proxyReq)
	if err != nil {
		log.Printf("Erreur lors de l'envoi de la requête: %v", err)
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer resp.Body.Close()

	if !shouldFilter {
		log.Printf("Réponse: %d %s", resp.StatusCode, resp.Status)
	}

	clientConn.Write([]byte(fmt.Sprintf("HTTP/%d.%d %d %s\r\n",
		resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)))

	for key, values := range resp.Header {
		for _, value := range values {
			clientConn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		}
	}
	clientConn.Write([]byte("\r\n"))

	written, _ := io.Copy(clientConn, resp.Body)
	if !shouldFilter {
		log.Printf("Body transféré: %d bytes", written)
	}
}

func Start() {
	go func() {
		cfg := config.GetInstance()
		addr := fmt.Sprintf("127.0.0.1:%d", cfg.ProxyPort)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Erreur lors du démarrage du proxy: %v", err)
		}
		defer listener.Close()

		log.Println("Proxy HTTP/HTTPS Interceptor with MITM")
		log.Printf("En attente de connexions sur %s...", addr)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Erreur lors de l'acceptation de la connexion: %v", err)
				continue
			}

			go handleConnection(conn)
		}
	}()
}

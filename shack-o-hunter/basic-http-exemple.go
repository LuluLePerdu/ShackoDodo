package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var directClient = &http.Client{
	Transport: &http.Transport{
		Proxy: nil, // IMPORTANT: never proxy the outbound (avoid loops)
	},
}

func main() {
	// 1) start proxy
	http.HandleFunc("/", handle)
	go func() {
		log.Println("Proxy HTTP sur 127.0.0.1:8181")
		log.Fatal(http.ListenAndServe("127.0.0.1:8181", nil))
	}()

	// 2) create a temp firefox profile with proxy forced
	prof, err := makeFirefoxProfile()
	if err != nil {
		log.Fatalf("Profil Firefox: %v", err)
	}

	// 3) launch firefox with that profile (no env var needed)
	firefoxPath := `C:\Program Files\Mozilla Firefox\firefox.exe`
	cmd := exec.Command(firefoxPath, "-no-remote", "-profile", prof, "http://neverssl.com")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("Impossible de lancer Firefox: %v", err)
	}

	select {}
}

func makeFirefoxProfile() (string, error) {
	root := filepath.Join(os.TempDir(), "ff-proxy-prof")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	// Force manual HTTP proxy only
	// network.proxy.type = 1 (manual)
	// network.proxy.http = 127.0.0.1
	// network.proxy.http_port = 8181
	// network.proxy.no_proxies_on = "" (proxy everything)
	// Disable proxy for HTTPS to keep this test strictly HTTP-side
	userJS := []string{
		`user_pref("network.proxy.type", 1);`,
		`user_pref("network.proxy.http", "127.0.0.1");`,
		`user_pref("network.proxy.http_port", 8181);`,
		`user_pref("network.proxy.no_proxies_on", "");`,
		// Optional: ensure HTTPS proxy unset so HTTPS won’t try CONNECT via us
		`user_pref("network.proxy.ssl", "");`,
		`user_pref("network.proxy.ssl_port", 0);`,
	}
	if err := os.WriteFile(filepath.Join(root, "user.js"), []byte(strings.Join(userJS, "\n")), 0o644); err != nil {
		return "", err
	}
	return root, nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	// Only plain HTTP requests (absolute-form in proxy mode)
	if r.URL.Scheme != "http" {
		http.Error(w, "only http (no https/connect)", http.StatusBadRequest)
		return
	}

	// Read body so we can print/modify/forward
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
	}
	r.Body.Close()

	// Log
	fmt.Println("===== REQUETE =====")
	fmt.Println("METHOD :", r.Method)
	fmt.Println("URL    :", r.URL.String())
	for k, v := range r.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	if len(body) > 0 {
		fmt.Println("BODY  :")
		fmt.Println(string(body))
	} else {
		fmt.Println("BODY  : (vide)")
	}

	// Pause
	fmt.Print("Entrée pour continuer...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

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

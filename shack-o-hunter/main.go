package main

import (
	"proxy-interceptor/firefox"
	"proxy-interceptor/proxy"
	"proxy-interceptor/websocket"
	"time"
)

func main() {
	// Mode MITM complet pour intercepter HTTPS
	proxy.Start()

	// Petit délai pour s'assurer que le certificat CA est bien généré
	time.Sleep(500 * time.Millisecond)

	websocket.Start()
	firefox.Start()

	select {}
}

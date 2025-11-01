package main

import (
	"proxy-interceptor/firefox"
	"proxy-interceptor/proxy"
	"proxy-interceptor/websocket"
)

func main() {
	proxy.Start()
	websocket.Start()
	firefox.Start()

	select {}
}

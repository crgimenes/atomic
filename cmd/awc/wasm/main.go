package main

import (
	"fmt"

	"golang.org/x/net/websocket"
)

func main() {
	// Connect to the server
	ws, err := websocket.Dial("ws://localhost:8080/echo/", "", "http://localhost")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	// Send a message to the server
	err = websocket.Message.Send(ws, "Hello from the client!")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read in a message from the server
	var msg string
	err = websocket.Message.Receive(ws, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print the message to the console
	fmt.Printf("Received: %s\n", msg)
}

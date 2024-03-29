package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

//go:embed assets
var assets embed.FS

func homeHandler(w http.ResponseWriter, r *http.Request) {
	index, err := assets.ReadFile("assets/index.html")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("index.html").Parse(string(index))
	if err != nil {
		log.Fatal(err)
	}

	// retorna o template renderizado
	err = t.Execute(w, nil)

}

func echoServer(ws *websocket.Conn) {
	defer ws.Close()
	ws.MaxPayloadBytes = 1024 * 1024
	//ws.SetDeadline(time.Now().Add(5 * time.Second))

	// timmer channel
	//timer := time.After(1 * time.Second)

	go func() {
		for {
			err := websocket.Message.Send(ws, "ping")
			if err != nil {
				fmt.Println("Error sending to websocket:", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		var message string
		err := websocket.Message.Receive(ws, &message)
		if err != nil {
			fmt.Println("Error reading from websocket:", err)
			break
		}
		fmt.Println("Received message:", message)
		if message == "pong" {
			continue
		}
		if message == "ping" {
			message = "pong"
		}
		err = websocket.Message.Send(ws, message)
		if err != nil {
			fmt.Println("Error sending to websocket:", err)
			break
		}
	}
}

func main() {
	port := 8080
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.FileServer(http.FS(assets)))
	mux.Handle("/echo", websocket.Handler(echoServer))
	mux.HandleFunc("/", homeHandler)

	s := &http.Server{
		Handler:        mux,
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on port %d\n", port)
	log.Fatal(s.ListenAndServe())
}

package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/server"
)

func main() {
	cfg := config.Get()

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		<-sc

		log.Println("shutting down...")

		os.Exit(0)
	}()

	s := server.New(cfg)

	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

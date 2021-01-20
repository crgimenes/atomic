package main

import (
	"fmt"
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

		// clear

		os.Exit(0)
	}()

	/*
	   0        1          2        3        4        5        6        7        8
	   123456789012234567890234567890234567890234567890234567890234567890234567890
	*/
	fmt.Print(`
 ██████╗██████╗  ██████╗    ███████╗████████╗██╗   ██████╗ ██████╗ 
██╔════╝██╔══██╗██╔════╝    ██╔════╝╚══██╔══╝██║   ██╔══██╗██╔══██╗
██║     ██████╔╝██║  ███╗   █████╗     ██║   ██║   ██████╔╝██████╔╝
██║     ██╔══██╗██║   ██║   ██╔══╝     ██║   ██║   ██╔══██╗██╔══██╗
╚██████╗██║  ██║╚██████╔╝██╗███████╗   ██║   ██║██╗██████╔╝██║  ██║
 ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝
`)

	fmt.Println("Database :", cfg.DatabasePath)
	s := server.New(cfg)

	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

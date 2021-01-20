package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/database/jsonfiles"
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

	db, err := jsonfiles.NewDatabase(cfg)
	if err != nil {
		panic(err)
	}

	s := server.New(cfg, db)

	bucket, err := db.UseBucket("bbs")
	if err != nil {
		panic(err)
	}

	err = bucket.Save("test", cfg)
	if err != nil {
		panic(err)
	}

	err = bucket.Load("test", &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", cfg)

	l, err := bucket.List()
	for k, v := range l {
		fmt.Println(k, v)
	}

	err = s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

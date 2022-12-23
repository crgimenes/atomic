package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"crg.eti.br/go/atomic/config"
	"crg.eti.br/go/atomic/server"
)

func validateRequiredFiles() bool {
	ret := true
	// validate if id_rsa and id_rsa.pub exists
	_, err := os.Stat("id_rsa")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		fmt.Println("id_rsa not found")
		fmt.Println(`run: ssh-keygen -t rsa -f id_rsa -N ""`)
		ret = false
	}

	_, err = os.Stat("id_rsa.pub")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		fmt.Println("id_rsa.pub not found")
		fmt.Println(`run: ssh-keygen -t rsa -f id_rsa -N ""`)
		ret = false
	}

	// validate if atomic.db exists
	_, err = os.Stat("atomic.db")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		fmt.Println("atomic.db not found")
		fmt.Println("run: atomicdb -new")
		ret = false
	}

	return ret
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// change to base bbs dir
	err = os.Chdir(cfg.BaseBBSDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("base bbs dir: %s", cfg.BaseBBSDir)

	// validate required files
	if !validateRequiredFiles() {
		os.Exit(1)
	}

	srv := server.New(cfg)

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		<-sc

		log.Println("shutting down...")

		os.Exit(0)
	}()

	err = srv.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

package main

import (
	"log"
	"os"

	_ "modernc.org/sqlite"

	"crg.eti.br/go/atomic/database"
	"crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
)

type Config struct {
	BaseBBSDir string `json:"base_bbs_dir" ini:"base_bbs_dir" cfg:"base_bbs_dir" cfgDefault:"./"`
	New        bool   `json:"new" ini:"new" cfg:"new" cfgDefault:"false"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// load config
	var cfg = Config{}
	config.PrefixEnv = "ATOMIC"
	config.File = "config.ini"
	err := config.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// change to base bbs dir
	err = os.Chdir(cfg.BaseBBSDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("base bbs dir: %s", cfg.BaseBBSDir)

	// open database
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	version, err := db.GetVersion()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SQLite version: %s", version)

	// create tables
	if cfg.New {
		err = db.RunMigration()
		if err != nil {
			log.Println(err)
		}
	}
}

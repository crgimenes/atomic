package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	c := &serial.Config{
		Name:        "/dev/tty.wchusbserial1410",
		Baud:        115200,
		ReadTimeout: 1 * time.Second,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	atList := []string{
		"ATE0\r\n",
		"AT\r\n",
	}

	for _, at := range atList {
		n, err := s.Write([]byte(at))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("write ", n)
		buf := make([]byte, 128)
		n, err = s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%q", buf[:n])
	}
}

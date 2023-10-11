package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

var (
	width  int
	height int
)

// get mouse position.
// echo "\e[?1003h\e[?1015h\e[?1006h";

// stop getting mouse position.
// echo -e "\e[?1000;1006;1015l"

func main() {

	var err error

	// handle interrupt signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Printf("\r\nInterrupted\r\n")
			os.Exit(1)
		}
	}()

	// handle size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			width, height, err = term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// clear screen.
	fmt.Printf("\033[2J")

	// start get mouse position.
	fmt.Printf("\033[?1003h\033[?1015h\033[?1006h")

	// handle input.
	for {
		var r rune
		_, err := fmt.Scanf("%c", &r)
		if err != nil {
			panic(err)
		}

		switch r {
		case 'q':
			fmt.Printf("Quit\r\n")

			// clear screen.
			fmt.Printf("\033[2J")

			// stop get mouse position.
			fmt.Printf("\033[?1000;1006;1015l")

			return

		default:
			// save cursor position.
			//fmt.Printf("\033[s")
			fmt.Printf("Read: %q 0x%X width: %d height: %d\r\n", r, r, width, height)

			// restore cursor position.
			//fmt.Printf("\033[u")
		}

		//fmt.Printf("%c", r)

	}

}

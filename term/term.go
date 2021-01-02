package term

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/crgimenes/atomic/charset"
)

type Term struct {
	W            int
	H            int
	C            io.Writer
	captureInput bool
	echo         bool
	inputField   string
	inputTrigger chan struct{}
}

func removeLastRune(s string) string {
	r := []rune(s)
	n := len(r) - 1
	if n < 0 {
		n = 0
	}
	return string(r[:n])
}

func (t *Term) Init() {
	t.inputTrigger = make(chan struct{})
}

func (t *Term) Clear() error {
	_, err := io.WriteString(t.C, "\u001b[2J")
	return err
}

func (t *Term) WriteString(s string) {
	_, err := io.WriteString(t.C, s)
	if err != nil {
		log.Println(err.Error())
	}
}

func (t *Term) SetEcho(b bool) {
	t.echo = b
}

func (t *Term) Print(lin, col int, s string) error {
	c := t.W + 1 - col
	l := len([]rune(s))
	if c > l {
		c = l
	}
	if c <= 0 {
		return nil
	}

	w := t.C
	s = string([]rune(s)[:c])

	a := [][]byte{
		([]byte("\u001b[")),
		([]byte(strconv.Itoa(lin))),
		([]byte{';'}),
		([]byte(strconv.Itoa(col))),
		([]byte{'f'}),
		([]byte(s)),
	}

	var err error
	for i := 0; i < 6; i++ {
		_, err = w.Write(a[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// Input receives user input and interprets depending on the state of the engine.
func (t *Term) Input(s string) {
	fmt.Printf("> %q -> %X\r\n", s, s)
	for _, c := range s {
		if t.captureInput {
			switch c {
			case '\u007f':
				fmt.Println("backspace UTF-8")
				fallthrough
			case '\b':
				if len(t.inputField) == 0 {
					return
				}
				t.WriteString("\b \b")
				fmt.Println("backspace")
				t.inputField = removeLastRune(t.inputField)
				return
			case '\r':
				t.captureInput = false
				t.inputTrigger <- struct{}{}
				return
			default:
				t.inputField += s
			}
		}
		if t.echo {
			if s[0] == '\x1b' {
				return
			}
			t.WriteString(s)
		}
	}
}

func (t *Term) GetField() string {
	t.echo = true
	t.captureInput = true
	t.inputField = ""

	<-t.inputTrigger

	res := t.inputField
	t.inputField = ""
	t.echo = false
	return res
}

func (t *Term) GetPassword() string {
	t.echo = false
	t.captureInput = true
	t.inputField = ""

	<-t.inputTrigger

	res := t.inputField
	t.inputField = ""
	return res
}

func (t *Term) WriteFromASCII(fileName string) int {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		t.WriteString(string(charset.ASCII[b]))
	}
	return 0
}

// ResetScreen reset terminal screen
func (t *Term) ResetScreen() int {
	t.WriteString("\u001bc")
	return 0
}

// Cls clear screen
func (t *Term) Cls() int {
	t.WriteString("\u001b[2J")
	return 0
}

// Write string
func (t *Term) Write(s string) int {
	t.WriteString(s)
	return 0
}

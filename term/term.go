package term

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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

var (
	ASCII = [256]rune{
		'\x00', '☺', '☻', '♥', '♦', '♣', '♠', '•', '\b', '\t', '\n', '♂', '♀', '\r', '♫', '☼',
		'►', '◄', '↕', '‼', '¶', '§', '▬', '↨', '↑', '↓', '→', '\x1b', '∟', '↔', '▲', '▼',
		' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
		'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
		'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
		'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
		'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', '⌂',
		'€', '\u0081', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
		'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', '¢', '£', '¥', '₧', 'ƒ',
		'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '⌐', '¬', '½', '¼', '¡', '«', '»',
		'░', '▒', '▓', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
		'└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
		'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫', '╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
		'α', 'ß', 'Γ', 'π', 'Σ', 'σ', 'µ', 'τ', 'Φ', 'Θ', 'Ω', 'δ', '∞', 'φ', 'ε', '∩',
		'≡', '±', '≥', '≤', '⌠', '⌡', '÷', '≈', '°', '∙', '·', '√', 'ⁿ', '²', '■', '\u00a0',
	}

	ASCII2UTF8 = make(map[rune]byte, 256)
)

func Init() {
	for i, r := range ASCII {
		ASCII2UTF8[r] = byte(i)
	}
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
	_, err := io.WriteString(t.C, "\033[2J")
	return err
}

func (t *Term) WriteString(s string) {
	_, err := io.WriteString(t.C, s)
	if err != nil {
		log.Println("term error writing string:", err)
	}
}

func (t *Term) WriteByte(b byte) {
	_, err := t.C.Write([]byte{b})
	if err != nil {
		log.Println("term error writing byte:", err)
	}
}

func (t *Term) WriteRune(r rune) {
	_, err := t.C.Write([]byte(string(r)))
	if err != nil {
		log.Println("term error writing rune:", err)
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
		[]byte("\033["),
		[]byte(strconv.Itoa(lin)),
		{';'},
		[]byte(strconv.Itoa(col)),
		{'f'},
		[]byte(s),
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
		fmt.Printf("\t%q -> %X\r", c, c)
		if t.captureInput {
			switch c {
			case '\x1b':
				if len(s) >= 3 {
					if s[1] == '[' {
						switch s[2] {
						case 'A':
							t.WriteString("\x1b[1A")
							return
						case 'B':
							t.WriteString("\x1b[1B")
							return
						case 'C':
							t.WriteString("\x1b[1C")
							return
						case 'D':
							t.WriteString("\x1b[1D")
							return
						}
					}
				}
				return
			case '\u007f', '\b':
				if len(t.inputField) == 0 {
					return
				}
				t.WriteString("\b \b")
				t.inputField = removeLastRune(t.inputField)
				return
			case '\r':
				t.captureInput = false
				t.inputTrigger <- struct{}{}
				return
			default:
				fmt.Printf("%q", c)
				t.inputField += string(c)
			}
		}

		if t.echo {
			t.WriteString(string(c))
		}
	}
}

func (t *Term) GetField() string {

	// get input from user

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
	var b byte
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	for {
		b, err = r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Fatal(err)
		}

		t.WriteString(string(ASCII[b]))
	}
	return 0
}

// ResetScreen reset terminal screen.
func (t *Term) ResetScreen() int {
	t.WriteString("\033c")
	return 0
}

// Cls clear screen.
func (t *Term) Cls() int {
	t.WriteString("\033[2J")
	return 0
}

func (t *Term) EnterScreen() int {
	t.WriteString("\033[?1049h\033[H")
	return 0
}

func (t *Term) ExitScreen() int {
	t.WriteString("\033[?1049l")
	return 0
}

func (t *Term) MoveCursor(x, y int) int {
	t.WriteString(fmt.Sprintf("\033[%d;%dH", y, x))
	return 0
}

// Write string.
func (t *Term) Write(s string) int {
	t.WriteString(s)
	return 0
}

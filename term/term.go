package term

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type OutputMode int

const (
	UTF8 OutputMode = iota
	CP437
	CP850
)

type Term struct {
	Width          int
	Height         int
	bufferPosition int
	MaxInputLength int
	C              io.Writer
	captureInput   bool
	echo           bool
	replaceInput   bool
	InputField     []rune
	InputTrigger   chan struct{}
	OutputMode     OutputMode
	OutputDelay    time.Duration
}

var (
	CP437_TO_UTF8 = [256]rune{
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

	CP850_TO_UTF8 = [256]rune{
		'\x00', '☺', '☻', '♥', '♦', '♣', '♠', '•', '\b', '\t', '\n', '♂', '♀', '\r', '♫', '☼',
		'►', '◄', '↕', '‼', '¶', '§', '▬', '↨', '↑', '↓', '→', '\x1b', '∟', '↔', '▲', '▼',
		' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
		'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
		'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
		'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
		'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', '⌂',
		'€', '\u0081', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
		'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', 'ø', '£', 'Ø', '×', 'ƒ',
		'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '®', '¬', '½', '¼', '¡', '«', '»',
		'░', '▒', '▓', '│', '┤', 'Á', 'Â', 'À', '©', '╣', '║', '╗', '╝', '¢', '¥', '┐',
		'└', '┴', '┬', '├', '─', '┼', 'ã', 'Ã', '╚', '╔', '╩', '╦', '╠', '═', '╬', '¤',
		'\u00f0', 'Ð', 'Ê', 'Ë', 'È', 'ı', 'Í', 'Î', 'Ï', '┘', '┌', '█', '▄', '¦', 'Ì', '▀',
		'Ó', 'ß', 'Ô', 'Ò', 'õ', 'Õ', 'µ', 'þ', 'Þ', 'Ú', 'Û', 'Ù', 'ý', 'Ý', '¯', '´',
		'\u00ad', '±', '‗', '¾', '¶', '§', '÷', '¸', '°', '¨', '·', '¹', '³', '²', '■', '\u00a0',
	}

	UTF8_TO_CP437 = make(map[rune]byte, 256)
	UTF8_TO_CP850 = make(map[rune]byte, 256)
)

func init() {
	for i, r := range CP437_TO_UTF8 {
		UTF8_TO_CP437[r] = byte(i)
	}

	for i, r := range CP850_TO_UTF8 {
		UTF8_TO_CP850[r] = byte(i)
	}
}

func removeLastRune(s []rune) []rune {
	n := len(s) - 1

	if n < 0 {
		n = 0
	}
	return s[:n]
}

func (t *Term) Clear() error {
	_, err := io.WriteString(t.C, "\033[2J\033[0;0H")
	return err
}

func (t *Term) GetOutputDisplay() string {
	switch t.OutputMode {
	case CP437:
		return "CP437"
	case CP850:
		return "CP850"
	case UTF8:
		return "UTF8"
	}
	return "Unknown"
}

func (t *Term) WriteString(s string) {
	if t.OutputMode == CP437 {
		for _, r := range s {
			t.WriteByte(UTF8_TO_CP437[r])
		}
		return
	}

	if t.OutputMode == CP850 {
		for _, r := range s {
			t.WriteByte(UTF8_TO_CP850[r])
		}
		return
	}

	if t.OutputDelay > 0 {
		for _, r := range s {
			t.C.Write([]byte(string(r)))
			time.Sleep(t.OutputDelay)
		}
		return
	}

	_, err := io.WriteString(t.C, s)
	if err != nil {
		log.Println("term error writing string:", err) // TODO: add return error
	}
}

func (t *Term) WriteByte(b byte) {

	if t.OutputDelay > 0 {
		t.C.Write([]byte{b})
		time.Sleep(t.OutputDelay)
		return
	}

	_, err := t.C.Write([]byte{b})
	if err != nil {
		log.Println("term error writing byte:", err)
	}
}

func (t *Term) WriteRune(r rune) {
	if t.OutputMode == CP437 {
		t.WriteByte(UTF8_TO_CP437[r])
		return
	}

	if t.OutputMode == CP850 {
		t.WriteByte(UTF8_TO_CP850[r])
		return
	}

	if t.OutputDelay > 0 {
		t.C.Write([]byte(string(r)))
		time.Sleep(t.OutputDelay)
		return
	}

	_, err := t.C.Write([]byte(string(r)))
	if err != nil {
		log.Println("term error writing rune:", err)
	}
}

func (t *Term) SetEcho(b bool) {
	t.echo = b
}

func (t *Term) Print(row, col int, s string) error {
	c := t.Width + 1 - col
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
		[]byte(strconv.Itoa(row)),
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
	for _, c := range s {
		if t.captureInput {
			switch c {
			case '\x1b':
				if len(s) >= 3 {
					if s[1] == '[' {
						switch s[2] {
						case 'A': // up
							// t.WriteString("\x1b[1A")
							return
						case 'B': // down
							// t.WriteString("\x1b[1B")
							return
						case 'C': // right
							t.bufferPosition++
							if t.bufferPosition > len(t.InputField) {
								t.bufferPosition = len(t.InputField)

								return
							}

							t.WriteString("\x1b[1C")
							return
						case 'D': // left
							t.bufferPosition--

							if t.bufferPosition < 0 {
								t.bufferPosition = 0

								return
							}

							t.WriteString("\x1b[1D")
							return

						case 'H': // home
							if t.bufferPosition == 0 {
								return
							}
							t.WriteString("\033[" + strconv.Itoa(t.bufferPosition) + "D")
							t.bufferPosition = 0
							return
						case 'F': // end
							if t.bufferPosition == len(t.InputField) {
								return
							}
							t.WriteString("\033[" + strconv.Itoa(len(t.InputField)-t.bufferPosition) + "C")
							t.bufferPosition = len(t.InputField)
							return

						case '3': // delete
							if s[3] == '~' {
								if t.bufferPosition == len(t.InputField) {
									return
								}
								if len(t.InputField) == 0 {
									return
								}
								if t.bufferPosition < len(t.InputField) {
									if t.bufferPosition == 0 {
										t.InputField = t.InputField[1:]
									} else {
										t.InputField = append(t.InputField[:t.bufferPosition], t.InputField[t.bufferPosition+1:]...)
									}
									// save cursor position
									t.WriteString("\033[s")
									// print input field from bufferPosition
									t.WriteString(string(t.InputField[t.bufferPosition:]))
									// print spaces to clear the rest of the line
									t.WriteString(" ")
									// restore cursor position
									t.WriteString("\033[u")

									if t.bufferPosition == len(t.InputField) {
										t.bufferPosition--
									}
								}
								return
							}
						}
					}
				}
				return
			case '\uf746': // insert
				t.replaceInput = !t.replaceInput
				return
			case '\u007f', '\b': // backspace
				if len(t.InputField) == 0 {
					return
				}
				if t.bufferPosition == 0 {
					return
				}
				t.WriteString("\b \b")
				// remove rune at bufferPosition
				if t.bufferPosition == len(t.InputField) {
					t.InputField = removeLastRune(t.InputField)
				} else {
					if t.bufferPosition == 1 {
						t.InputField = t.InputField[1:]
					} else {
						t.InputField = append(t.InputField[:t.bufferPosition-1], t.InputField[t.bufferPosition:]...)
					}
				}
				t.bufferPosition--
				if t.bufferPosition < 0 {
					t.bufferPosition = 0
				}
				// save cursor position
				t.WriteString("\033[s")
				// print input field from bufferPosition
				t.WriteString(string(t.InputField[t.bufferPosition:]))
				// print spaces to clear the rest of the line
				t.WriteString(" ")
				// restore cursor position
				t.WriteString("\033[u")

				return
			case '\r':
				t.captureInput = false
				t.InputTrigger <- struct{}{}
				t.bufferPosition = 0
				return
			case '\x01': // ctrl + a (home)
				if t.bufferPosition == 0 {
					return
				}
				t.WriteString("\033[" + strconv.Itoa(t.bufferPosition) + "D")
				t.bufferPosition = 0
				return
			case '\x05': // ctrl + e (end)
				if t.bufferPosition == len(t.InputField) {
					return
				}
				t.WriteString("\033[" + strconv.Itoa(len(t.InputField)-t.bufferPosition) + "C")
				t.bufferPosition = len(t.InputField)
				return

			default:
				if t.MaxInputLength > 0 && len(t.InputField) >= t.MaxInputLength {
					return
				}

				if t.replaceInput {
					if t.bufferPosition == len(t.InputField) {
						t.InputField = append(t.InputField, c)
						t.bufferPosition++

						if t.echo {
							t.WriteString(string(c))
						}
						continue
					}

					raux := t.InputField[t.bufferPosition+1:]
					t.InputField = append(t.InputField[:t.bufferPosition], c)
					t.InputField = append(t.InputField, raux...)
					t.bufferPosition++
				} else {

					if t.bufferPosition == len(t.InputField) {
						t.InputField = append(t.InputField, c)
						t.bufferPosition++

						if t.echo {
							t.WriteString(string(c))
						}
						continue
					}

					var (
						end   = make([]rune, len(t.InputField[t.bufferPosition:]))
						start = make([]rune, len(t.InputField[:t.bufferPosition]))
					)
					copy(end, t.InputField[t.bufferPosition:])
					copy(start, t.InputField)
					start = append(start, c)
					start = append(start, end...)
					t.InputField = start

					t.bufferPosition++

					if t.echo {
						// save cursor position
						t.WriteString("\033[s")

						// print input field from buffer position
						t.WriteString(string(c))
						t.WriteString(string(t.InputField[t.bufferPosition:]))

						// restore cursor position
						t.WriteString("\033[u")
					}
				}
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
	t.InputField = []rune{}

	<-t.InputTrigger

	res := t.InputField
	t.InputField = []rune{}
	t.echo = false
	return string(res)
}

func (t *Term) GetPassword() string {
	t.echo = false
	t.captureInput = true
	t.InputField = []rune{}

	<-t.InputTrigger

	res := t.InputField
	t.InputField = []rune{}
	return string(res)
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

		t.WriteString(string(CP437_TO_UTF8[b]))
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

func (t *Term) MoveCursor(row, col int) int {
	t.WriteString(fmt.Sprintf("\033[%d;%dH", row, col))
	return 0
}

// Write string.
func (t *Term) Write(s string) int {
	t.WriteString(s)
	return 0
}

func (t *Term) SetOutputMode(mode string) {
	log.Println("Setting output mode to: " + mode)
	mode = strings.ToUpper(mode)
	switch mode {
	case "UTF8", "UTF-8":
		t.OutputMode = UTF8
	case "CP437":
		t.OutputMode = CP437
	case "CP850":
		t.OutputMode = CP850
	default:
		t.OutputMode = UTF8
		log.Printf("Invalid output mode: %s. Using UTF-8", mode)
	}
}

func (t *Term) SetInputLimit(limit int) {
	t.MaxInputLength = limit
}

func (t *Term) SetOutputDelay(delay int) {
	t.OutputDelay = time.Duration(delay) * time.Millisecond
}

func (t *Term) DrawBox(row, col, width, height int) {
	t.MoveCursor(row, col)
	t.WriteString("┌")
	for i := 0; i < width-2; i++ {
		t.WriteString("─")
	}
	t.WriteString("┐")

	for i := 0; i < height-2; i++ {
		t.MoveCursor(row+i+1, col)
		t.WriteString("│")
		t.MoveCursor(row+i+1, col+width-1)
		t.WriteString("│")
	}

	t.MoveCursor(row+height-1, col)
	t.WriteString("└")
	for i := 0; i < width-2; i++ {
		t.WriteString("─")
	}

	t.WriteString("┘")
}

// GetSize
func (t *Term) GetSize() (int, int) {
	return t.Width, t.Height
}

package term

import (
	"io"
	"strconv"
)

type Term struct {
	W int
	H int
	C io.Writer
}

func (t *Term) Clear() error {
	_, err := io.WriteString(t.C, "\u001b[2J")
	return err
}

func (t *Term) Print(lin, col int, s string) {
	c := t.W + 1 - col
	l := len([]rune(s))
	if c > l {
		c = l
	}
	if c <= 0 {
		return
	}

	w := t.C
	s = string([]rune(s)[:c])

	w.Write([]byte("\u001b["))
	w.Write([]byte(strconv.Itoa(lin)))
	w.Write([]byte{';'})
	w.Write([]byte(strconv.Itoa(col)))
	w.Write([]byte{'f'})
	w.Write([]byte(s))
	return
}

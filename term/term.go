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

package term

import (
	"fmt"
	"io"
)

type Term struct {
	W int
	H int
	C io.Writer
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
	s = string([]rune(s)[:c])
	f := fmt.Sprintf("\u001b[%d;%df", lin, col)
	_, err := io.WriteString(t.C, f+s)
	return err
}

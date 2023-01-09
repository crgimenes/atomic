package main

/*
\x1b[38;2;r;g;bm // fg
\x1b[48;2;r;g;bm // bg
\x1b[0m // reset

chars: "█", "▀", "▄", " "
*/

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strconv"

	goconfig "crg.eti.br/go/config"
)

// RGB color
type RGB struct {
	R, G, B uint32
}

// ImgToANSI holds module parameters
type ImgToANSI struct {
	DefaultColor RGB
}

type config struct {
	FileName string `cfg:"f" cfgRequired:"true"`
	RGB      string `cfg:"rgb"`
}

var (
	reset   = []byte("\033[m")
	resetln = []byte("\033[m\r\n")
)

// New create a new instance of ImgToANSI
func New() *ImgToANSI {
	return &ImgToANSI{}
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		fmt.Println("error closing file", err)
	}
}

// SetRGB update RGB values in current instance of ImgToANSI
func (p *ImgToANSI) SetRGB(rgb string) error {
	x, err := strconv.ParseUint(rgb, 16, 64)
	if err != nil {
		return err
	}

	r := uint8(x >> 16)
	g := uint8(x >> 8)
	b := uint8(x)

	p.DefaultColor.R = uint32(r)
	p.DefaultColor.G = uint32(g)
	p.DefaultColor.B = uint32(b)
	return nil
}

// PrintFile print a png file to the stdout using ANSI codes
func (p *ImgToANSI) PrintFile(fileName string, defaultRGB string) error {
	return p.FprintFile(os.Stdout, fileName, defaultRGB)
}

// FprintFile write a file to the Stdout using ANSI codes
func (p *ImgToANSI) FprintFile(w io.Writer, fileName string, defaultRGB string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer closer(f)

	img, err := png.Decode(f)
	if err != nil {
		return err
	}

	if defaultRGB != "" {
		err = p.SetRGB(defaultRGB)
		if err != nil {
			return err
		}
	}
	return p.Fprint(w, img)
}

// Print prints a image in the stdout using ANSI codes
func (p *ImgToANSI) Print(img image.Image) error {
	return p.Fprint(os.Stdout, img)
}

func (p *ImgToANSI) pxColor(x, y int, img image.Image) (r, g, b uint32) {
	px := img.At(x, y)
	r, g, b, a := px.RGBA()
	if a == 0 {
		r = p.DefaultColor.R
		g = p.DefaultColor.G
		b = p.DefaultColor.B
	}
	return r, g, b
}

/*
VGA 4 bit ANSI color codes
Background and Foreground
Name    			fg  	bg		RGB
Black				30		40		0,0,0
Red					31		41		170,0,0
Green				32		42		0,170,0
Yellow				33		43		170,85,0
Blue				34		44		0,0,170
Magenta				35		45		170,0,170
Cyan				36		46		0,170,170
White				37		47		170,170,170
Foreground Only
Bright Black		90		-		85,85,85
Bright Red			91		-		255,85,85
Bright Green		92		-		85,255,85
Bright Yellow		93		-		255,255,85
Bright Blue			94		-		85,85,255
Bright Magenta		95		-		255,85,255
Bright Cyan			96		-		85,255,255
Bright White		97		-	  	255,255,255
*/

// RGB2VGAFg convert RGB to 4 bit VGA ANSI colorcode (foreground)
// or return false if no match found
func RGB2VGAFg(r, g, b uint32) (bool, int) {
	red := r >> 8
	green := g >> 8
	blue := b >> 8
	switch {
	case red == 0 && green == 0 && blue == 0:
		return true, 30
	case red == 170 && green == 0 && blue == 0:
		return true, 31
	case red == 0 && green == 170 && blue == 0:
		return true, 32
	case red == 170 && green == 85 && blue == 0:
		return true, 33
	case red == 0 && green == 0 && blue == 170:
		return true, 34
	case red == 170 && green == 0 && blue == 170:
		return true, 35
	case red == 0 && green == 170 && blue == 170:
		return true, 36
	case red == 170 && green == 170 && blue == 170:
		return true, 37
	case red == 85 && green == 85 && blue == 85:
		return true, 90
	case red == 255 && green == 85 && blue == 85:
		return true, 91
	case red == 85 && green == 255 && blue == 85:
		return true, 92
	case red == 255 && green == 255 && blue == 85:
		return true, 93
	case red == 85 && green == 85 && blue == 255:
		return true, 94
	case red == 255 && green == 85 && blue == 255:
		return true, 95
	case red == 85 && green == 255 && blue == 255:
		return true, 96
	case red == 255 && green == 255 && blue == 255:
		return true, 97
	default:
		return false, 0
	}
}

// RGB2VGABg convert RGB to 4 bit VGA ANSI colorcode (background)
// or return false if no match found
func RGB2VGABg(r, g, b uint32) (bool, int) {
	red := r >> 8
	green := g >> 8
	blue := b >> 8
	switch {
	case red == 0 && green == 0 && blue == 0:
		return true, 40
	case red == 170 && green == 0 && blue == 0:
		return true, 41
	case red == 0 && green == 170 && blue == 0:
		return true, 42
	case red == 170 && green == 85 && blue == 0:
		return true, 43
	case red == 0 && green == 0 && blue == 170:
		return true, 44
	case red == 170 && green == 0 && blue == 170:
		return true, 45
	case red == 0 && green == 170 && blue == 170:
		return true, 46
	case red == 170 && green == 170 && blue == 170:
		return true, 47
	default:
		return false, 0
	}
}

// Fprint prints write a image to a writer using ANSI codes
func (p *ImgToANSI) Fprint(w io.Writer, img image.Image) error {
	bound := img.Bounds()

	var (
		fgCode string
		bgCode string
		err    error
	)

	for y := bound.Min.Y; y < bound.Max.Y; y += 2 {
		for x := bound.Min.X; x < bound.Max.X; x++ {

			fr, fg, fb := p.pxColor(x, y, img)
			br, bg, bb := p.pxColor(x, y+1, img)

			okfg, fgColor := RGB2VGAFg(fr, fg, fb)
			okbg, bgColor := RGB2VGABg(br, bg, bb)

			if okfg && okbg {
				s := fmt.Sprintf("\033[%d;%dm▀", fgColor, bgColor)
				_, err = w.Write([]byte(s))
				if err != nil {
					return err
				}
				continue
			}

			fgCode = fmt.Sprintf("\033[38;2;%d;%d;%dm",
				uint8(fr), uint8(fg), uint8(fb))
			bgCode = fmt.Sprintf("\033[48;2;%d;%d;%dm",
				uint8(br), uint8(bg), uint8(bb))

			/////////////////////////////////
			_, err = w.Write([]byte(bgCode))
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(fgCode))
			if err != nil {
				return err
			}
			/////////////////////////////////

			_, err = w.Write([]byte("▀"))
			if err != nil {
				return err
			}
		}
		_, err = w.Write(resetln)
		if err != nil {
			return err
		}
		//lastFgCode = ""
		//lastBgCode = ""
	}
	return nil
}

func main() {
	cfg := config{}
	goconfig.PrefixEnv = "PNG2ANSI"
	err := goconfig.Parse(&cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := New()
	err = p.PrintFile(cfg.FileName, cfg.RGB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

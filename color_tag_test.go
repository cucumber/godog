package godog

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cucumber/godog/colors"
)

type csiState int

const (
	outsideCsiCode csiState = iota
	firstCsiCode
	secondCsiCode
)

type tagColorWriter struct {
	w             io.Writer
	state         csiState
	paramStartBuf bytes.Buffer
	paramBuf      bytes.Buffer
	tag           string
}

const (
	firstCsiChar   byte = '\x1b'
	secondeCsiChar byte = '['
	separatorChar  byte = ';'
	sgrCode        byte = 'm'
)

const (
	ansiReset        = "0"
	ansiIntensityOn  = "1"
	ansiIntensityOff = "21"
	ansiUnderlineOn  = "4"
	ansiUnderlineOff = "24"
	ansiBlinkOn      = "5"
	ansiBlinkOff     = "25"

	ansiForegroundBlack   = "30"
	ansiForegroundRed     = "31"
	ansiForegroundGreen   = "32"
	ansiForegroundYellow  = "33"
	ansiForegroundBlue    = "34"
	ansiForegroundMagenta = "35"
	ansiForegroundCyan    = "36"
	ansiForegroundWhite   = "37"
	ansiForegroundDefault = "39"

	ansiBackgroundBlack   = "40"
	ansiBackgroundRed     = "41"
	ansiBackgroundGreen   = "42"
	ansiBackgroundYellow  = "43"
	ansiBackgroundBlue    = "44"
	ansiBackgroundMagenta = "45"
	ansiBackgroundCyan    = "46"
	ansiBackgroundWhite   = "47"
	ansiBackgroundDefault = "49"

	ansiLightForegroundGray    = "90"
	ansiLightForegroundRed     = "91"
	ansiLightForegroundGreen   = "92"
	ansiLightForegroundYellow  = "93"
	ansiLightForegroundBlue    = "94"
	ansiLightForegroundMagenta = "95"
	ansiLightForegroundCyan    = "96"
	ansiLightForegroundWhite   = "97"

	ansiLightBackgroundGray    = "100"
	ansiLightBackgroundRed     = "101"
	ansiLightBackgroundGreen   = "102"
	ansiLightBackgroundYellow  = "103"
	ansiLightBackgroundBlue    = "104"
	ansiLightBackgroundMagenta = "105"
	ansiLightBackgroundCyan    = "106"
	ansiLightBackgroundWhite   = "107"
)

var colorMap = map[string]string{
	ansiForegroundBlack:   "black",
	ansiForegroundRed:     "red",
	ansiForegroundGreen:   "green",
	ansiForegroundYellow:  "yellow",
	ansiForegroundBlue:    "blue",
	ansiForegroundMagenta: "magenta",
	ansiForegroundCyan:    "cyan",
	ansiForegroundWhite:   "white",
	ansiForegroundDefault: "",
}

func (cw *tagColorWriter) flushBuffer() (int, error) {
	return cw.flushTo(cw.w)
}

func (cw *tagColorWriter) resetBuffer() (int, error) {
	return cw.flushTo(nil)
}

func (cw *tagColorWriter) flushTo(w io.Writer) (int, error) {
	var n1, n2 int
	var err error

	startBytes := cw.paramStartBuf.Bytes()
	cw.paramStartBuf.Reset()
	if w != nil {
		n1, err = cw.w.Write(startBytes)
		if err != nil {
			return n1, err
		}
	} else {
		n1 = len(startBytes)
	}
	paramBytes := cw.paramBuf.Bytes()
	cw.paramBuf.Reset()
	if w != nil {
		n2, err = cw.w.Write(paramBytes)
		if err != nil {
			return n1 + n2, err
		}
	} else {
		n2 = len(paramBytes)
	}
	return n1 + n2, nil
}

func isParameterChar(b byte) bool {
	return ('0' <= b && b <= '9') || b == separatorChar
}

func (cw *tagColorWriter) Write(p []byte) (int, error) {
	r, nw, first, last := 0, 0, 0, 0

	var err error
	for i, ch := range p {
		switch cw.state {
		case outsideCsiCode:
			if ch == firstCsiChar {
				cw.paramStartBuf.WriteByte(ch)
				cw.state = firstCsiCode
			}
		case firstCsiCode:
			switch ch {
			case firstCsiChar:
				cw.paramStartBuf.WriteByte(ch)
				break
			case secondeCsiChar:
				cw.paramStartBuf.WriteByte(ch)
				cw.state = secondCsiCode
				last = i - 1
			default:
				cw.resetBuffer()
				cw.state = outsideCsiCode
			}
		case secondCsiCode:
			if isParameterChar(ch) {
				cw.paramBuf.WriteByte(ch)
			} else {
				nw, err = cw.w.Write(p[first:last])
				r += nw
				if err != nil {
					return r, err
				}
				first = i + 1
				if ch == sgrCode {
					cw.changeColor()
				}
				n, _ := cw.resetBuffer()
				// Add one more to the size of the buffer for the last ch
				r += n + 1

				cw.state = outsideCsiCode
			}
		default:
			cw.state = outsideCsiCode
		}
	}

	if cw.state == outsideCsiCode {
		nw, err = cw.w.Write(p[first:])
		r += nw
	}

	return r, err
}

func (cw *tagColorWriter) changeColor() {
	strParam := cw.paramBuf.String()
	if len(strParam) <= 0 {
		strParam = "0"
	}
	csiParam := strings.Split(strParam, string(separatorChar))
	for _, p := range csiParam {
		c, ok := colorMap[p]
		switch {
		case !ok:
			switch p {
			case ansiReset:
				fmt.Fprint(cw.w, "</"+cw.tag+">")
				cw.tag = ""
			case ansiIntensityOn:
				cw.tag = "bold-" + cw.tag
			case ansiIntensityOff:
			case ansiUnderlineOn:
			case ansiUnderlineOff:
			case ansiBlinkOn:
			case ansiBlinkOff:
			default:
				// unknown code
			}
		default:
			cw.tag += c
			fmt.Fprint(cw.w, "<"+cw.tag+">")
		}
	}
}

func TestTagColorWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &tagColorWriter{w: &buf}

	s := fmt.Sprintf("text %s then %s", colors.Red("in red"), colors.Yellow("yel"))
	fmt.Fprint(w, s)

	expected := "text <red>in red</red> then <yellow>yel</yellow>"
	if buf.String() != expected {
		t.Fatalf("expected `%s` but got `%s`", expected, buf.String())
	}
}

func TestTagBoldColorWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &tagColorWriter{w: &buf}

	s := fmt.Sprintf(
		"text %s then %s",
		colors.Bold(colors.Red)("in red"),
		colors.Bold(colors.Yellow)("yel"),
	)
	fmt.Fprint(w, s)

	expected := "text <bold-red>in red</bold-red> then <bold-yellow>yel</bold-yellow>"
	if buf.String() != expected {
		t.Fatalf("expected `%s` but got `%s`", expected, buf.String())
	}
}

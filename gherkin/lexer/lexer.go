package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Lexer struct {
	reader *bufio.Reader
	peek   *Token
	lines  int
}

func New(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(r),
	}
}

func (l *Lexer) Next(skip ...TokenType) (t *Token) {
	if l.peek != nil {
		t = l.peek
		l.peek = nil
	} else {
		t = l.read()
	}

	for _, typ := range skip {
		if t.Type == typ {
			return l.Next(skip...)
		}
	}
	return
}

func (l *Lexer) Peek() *Token {
	if l.peek == nil {
		l.peek = l.read()
	}
	return l.peek
}

func (l *Lexer) read() *Token {
	line, err := l.reader.ReadString(byte('\n'))
	if err != nil && len(line) == 0 {
		return &Token{
			Type: EOF,
			Line: l.lines,
		}
	}
	l.lines++
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	// newline
	if len(line) == 0 {
		return &Token{
			Type: NEW_LINE,
			Line: l.lines - 1,
		}
	}
	// comment
	if m := matchers["comment"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   COMMENT,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[2],
			Text:   line,
		}
	}
	// pystring
	if m := matchers["pystring"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   PYSTRING,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Text:   line,
		}
	}
	// step
	if m := matchers["step"].FindStringSubmatch(line); len(m) > 0 {
		tok := &Token{
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[3],
			Text:   line,
		}
		switch m[2] {
		case "Given":
			tok.Type = GIVEN
		case "When":
			tok.Type = WHEN
		case "Then":
			tok.Type = THEN
		case "And":
			tok.Type = AND
		case "But":
			tok.Type = BUT
		}
		return tok
	}
	// scenario
	if m := matchers["scenario"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   SCENARIO,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[2],
			Text:   line,
		}
	}
	// background
	if m := matchers["background"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   BACKGROUND,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Text:   line,
		}
	}
	// feature
	if m := matchers["feature"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   FEATURE,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[2],
			Text:   line,
		}
	}
	// tags
	if m := matchers["tags"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   TAGS,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[2],
			Text:   line,
		}
	}
	// table row
	if m := matchers["table_row"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   TABLE_ROW,
			Indent: len(m[1]),
			Line:   l.lines - 1,
			Value:  m[2],
			Text:   line,
		}
	}
	// text
	text := strings.TrimLeftFunc(line, unicode.IsSpace)
	return &Token{
		Type:   TEXT,
		Line:   l.lines - 1,
		Value:  text,
		Indent: len(line) - len(text),
		Text:   line,
	}
}

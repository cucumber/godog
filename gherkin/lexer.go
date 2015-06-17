package gherkin

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
)

var matchers = map[string]*regexp.Regexp{
	"feature":          regexp.MustCompile("^(\\s*)Feature:\\s*([^#]*)(#.*)?"),
	"scenario":         regexp.MustCompile("^(\\s*)Scenario:\\s*([^#]*)(#.*)?"),
	"scenario_outline": regexp.MustCompile("^(\\s*)Scenario Outline:\\s*([^#]*)(#.*)?"),
	"examples":         regexp.MustCompile("^(\\s*)Examples:(\\s*#.*)?"),
	"background":       regexp.MustCompile("^(\\s*)Background:(\\s*#.*)?"),
	"step":             regexp.MustCompile("^(\\s*)(Given|When|Then|And|But)\\s+([^#]*)(#.*)?"),
	"comment":          regexp.MustCompile("^(\\s*)#(.+)"),
	"pystring":         regexp.MustCompile("^(\\s*)\\\"\\\"\\\""),
	"tags":             regexp.MustCompile("^(\\s*)@([^#]*)(#.*)?"),
	"table_row":        regexp.MustCompile("^(\\s*)\\|([^#]*)(#.*)?"),
}

type lexer struct {
	reader *bufio.Reader
	lines  int
}

func newLexer(r io.Reader) *lexer {
	return &lexer{
		reader: bufio.NewReader(r),
	}
}

func (l *lexer) read() *Token {
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
			Line: l.lines,
		}
	}
	// comment
	if m := matchers["comment"].FindStringSubmatch(line); len(m) > 0 {
		comment := strings.TrimSpace(m[2])
		return &Token{
			Type:    COMMENT,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   comment,
			Text:    line,
			Comment: comment,
		}
	}
	// pystring
	if m := matchers["pystring"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:   PYSTRING,
			Indent: len(m[1]),
			Line:   l.lines,
			Text:   line,
		}
	}
	// step
	if m := matchers["step"].FindStringSubmatch(line); len(m) > 0 {
		tok := &Token{
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[3]),
			Text:    line,
			Comment: strings.Trim(m[4], " #"),
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
			Type:    SCENARIO,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[2]),
			Text:    line,
			Comment: strings.Trim(m[3], " #"),
		}
	}
	// background
	if m := matchers["background"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    BACKGROUND,
			Indent:  len(m[1]),
			Line:    l.lines,
			Text:    line,
			Comment: strings.Trim(m[2], " #"),
		}
	}
	// feature
	if m := matchers["feature"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    FEATURE,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[2]),
			Text:    line,
			Comment: strings.Trim(m[3], " #"),
		}
	}
	// tags
	if m := matchers["tags"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    TAGS,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[2]),
			Text:    line,
			Comment: strings.Trim(m[3], " #"),
		}
	}
	// table row
	if m := matchers["table_row"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    TABLE_ROW,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[2]),
			Text:    line,
			Comment: strings.Trim(m[3], " #"),
		}
	}
	// scenario outline
	if m := matchers["scenario_outline"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    SCENARIO_OUTLINE,
			Indent:  len(m[1]),
			Line:    l.lines,
			Value:   strings.TrimSpace(m[2]),
			Text:    line,
			Comment: strings.Trim(m[3], " #"),
		}
	}
	// examples
	if m := matchers["examples"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    EXAMPLES,
			Indent:  len(m[1]),
			Line:    l.lines,
			Text:    line,
			Comment: strings.Trim(m[2], " #"),
		}
	}
	// text
	text := strings.TrimLeftFunc(line, unicode.IsSpace)
	return &Token{
		Type:   TEXT,
		Line:   l.lines,
		Value:  text,
		Indent: len(line) - len(text),
		Text:   line,
	}
}

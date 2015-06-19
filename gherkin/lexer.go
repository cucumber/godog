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

// for now only english language is supported
var keywords = map[TokenType]string{
	// special
	ILLEGAL:   "Illegal",
	EOF:       "End of file",
	NEW_LINE:  "New line",
	TAGS:      "Tags",
	COMMENT:   "Comment",
	PYSTRING:  "PyString",
	TABLE_ROW: "Table row",
	TEXT:      "Text",
	// general
	GIVEN:            "Given",
	WHEN:             "When",
	THEN:             "Then",
	AND:              "And",
	BUT:              "But",
	FEATURE:          "Feature",
	BACKGROUND:       "Background",
	SCENARIO:         "Scenario",
	SCENARIO_OUTLINE: "Scenario Outline",
	EXAMPLES:         "Examples",
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
			Type:    EOF,
			Line:    l.lines + 1,
			Keyword: keywords[EOF],
		}
	}
	l.lines++
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	// newline
	if len(line) == 0 {
		return &Token{
			Type:    NEW_LINE,
			Line:    l.lines,
			Keyword: keywords[NEW_LINE],
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
			Keyword: keywords[COMMENT],
		}
	}
	// pystring
	if m := matchers["pystring"].FindStringSubmatch(line); len(m) > 0 {
		return &Token{
			Type:    PYSTRING,
			Indent:  len(m[1]),
			Line:    l.lines,
			Text:    line,
			Keyword: keywords[PYSTRING],
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
		tok.Keyword = keywords[tok.Type]
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
			Keyword: keywords[SCENARIO],
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
			Keyword: keywords[BACKGROUND],
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
			Keyword: keywords[FEATURE],
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
			Keyword: keywords[TAGS],
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
			Keyword: keywords[TABLE_ROW],
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
			Keyword: keywords[SCENARIO_OUTLINE],
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
			Keyword: keywords[EXAMPLES],
		}
	}
	// text
	text := strings.TrimLeftFunc(line, unicode.IsSpace)
	return &Token{
		Type:    TEXT,
		Line:    l.lines,
		Value:   text,
		Indent:  len(line) - len(text),
		Text:    line,
		Keyword: keywords[TEXT],
	}
}

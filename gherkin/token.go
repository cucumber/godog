package gherkin

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	COMMENT
	NEW_LINE
	EOF
	TEXT
	TAGS
	TABLE_ROW
	PYSTRING
	FEATURE
	BACKGROUND
	SCENARIO
	SCENARIO_OUTLINE
	EXAMPLES
	GIVEN
	WHEN
	THEN
	AND
	BUT
)

// String gives a string representation of token type
func (t TokenType) String() string {
	return keywords[t]
}

// Token represents a line in gherkin feature file
type Token struct {
	Type         TokenType // type of token
	Line, Indent int       // line and indentation number
	Value        string    // interpreted value
	Text         string    // same text as read
	Keyword      string    // @TODO: the translated keyword
	Comment      string    // a comment
}

// OfType checks whether token is one of types
func (t *Token) OfType(all ...TokenType) bool {
	for _, typ := range all {
		if typ == t.Type {
			return true
		}
	}
	return false
}

// Length gives a token text length with indentation
// and keyword, but without comment
func (t *Token) Length() int {
	if pos := strings.Index(t.Text, "#"); pos != -1 {
		return len(strings.TrimRightFunc(t.Text[:pos], unicode.IsSpace))
	}
	return len(t.Text)
}

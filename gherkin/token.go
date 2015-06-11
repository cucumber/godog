package gherkin

type TokenType int

const (
	ILLEGAL TokenType = iota

	specials
	COMMENT
	NEW_LINE
	EOF

	elements
	TEXT
	TAGS
	TABLE_ROW
	PYSTRING

	keywords
	FEATURE
	BACKGROUND
	SCENARIO
	SCENARIO_OUTLINE
	EXAMPLES

	steps
	GIVEN
	WHEN
	THEN
	AND
	BUT
)

func (t TokenType) String() string {
	switch t {
	case COMMENT:
		return "comment"
	case NEW_LINE:
		return "new line"
	case EOF:
		return "end of file"
	case TEXT:
		return "text"
	case TAGS:
		return "tags"
	case TABLE_ROW:
		return "table row"
	case PYSTRING:
		return "pystring"
	case FEATURE:
		return "feature"
	case BACKGROUND:
		return "background"
	case SCENARIO:
		return "scenario"
	case SCENARIO_OUTLINE:
		return "scenario outline"
	case EXAMPLES:
		return "examples"
	case GIVEN:
		return "given step"
	case WHEN:
		return "when step"
	case THEN:
		return "then step"
	case AND:
		return "and step"
	case BUT:
		return "but step"
	}
	return "illegal"
}

type Token struct {
	Type         TokenType // type of token
	Line, Indent int       // line and indentation number
	Value        string    // interpreted value
	Text         string    // same text as read
	Comment      string    // a comment
}

func (t *Token) OfType(all ...TokenType) bool {
	for _, typ := range all {
		if typ == t.Type {
			return true
		}
	}
	return false
}

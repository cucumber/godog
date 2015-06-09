package lexer

type Token struct {
	Type         TokenType // type of token
	Line, Indent int       // line and indentation number
	Value        string    // interpreted value
	Text         string    // same text as read
}

func (t *Token) OfType(all ...TokenType) bool {
	for _, typ := range all {
		if typ == t.Type {
			return true
		}
	}
	return false
}

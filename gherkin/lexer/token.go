package lexer

type Token struct {
	Type         TokenType
	Line, Indent int
	Value        string
}

func (t *Token) OfType(all ...TokenType) bool {
	for _, typ := range all {
		if typ == t.Type {
			return true
		}
	}
	return false
}

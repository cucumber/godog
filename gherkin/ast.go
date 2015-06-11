package gherkin

import "github.com/DATA-DOG/godog/gherkin/lexer"

type item struct {
	next, prev *item
	value      *lexer.Token
}

// AST is a linked list to store gherkin Tokens
// used to insert errors and other details into
// the token tree
type AST struct {
	head, tail *item
}

func newAST() *AST {
	return &AST{}
}

func (l *AST) addTail(t *lexer.Token) *item {
	it := &item{next: nil, prev: l.tail, value: t}
	if l.head == nil {
		l.head = it
	} else {
		l.tail.next = it
	}
	l.tail = it
	return l.tail
}

func (l *AST) addBefore(t *lexer.Token, i *item) *item {
	it := &item{next: i, prev: i.prev, value: t}
	i.prev = it
	if it.prev == nil {
		l.head = it
	}
	return it
}

func (l *AST) addAfter(t *lexer.Token, i *item) *item {
	it := &item{next: i.next, prev: i, value: t}
	i.next = it
	if it.next == nil {
		l.tail = it
	}
	return it
}

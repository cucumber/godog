package gherkin

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/l3pp4rd/go-behat/gherkin/lexer"
)

type Tag string

type Tags []Tag

func (t Tags) Has(tag Tag) bool {
	for _, tg := range t {
		if tg == tag {
			return true
		}
	}
	return false
}

type Scenario struct {
	Title   string
	Steps   []*Step
	Tags    Tags
	Comment string
}

type Background struct {
	Steps   []*Step
	Comment string
}

type StepType string

const (
	Given StepType = "Given"
	When  StepType = "When"
	Then  StepType = "Then"
)

type Step struct {
	Text     string
	Comment  string
	Type     StepType
	PyString *PyString
	Table    *Table
}

type Feature struct {
	Path        string
	Tags        Tags
	Description string
	Title       string
	Background  *Background
	Scenarios   []*Scenario
	AST         *AST
	Comment     string
}

type PyString struct {
	Body string
}

type Table struct {
	rows [][]string
}

var allSteps = []lexer.TokenType{
	lexer.GIVEN,
	lexer.WHEN,
	lexer.THEN,
	lexer.AND,
	lexer.BUT,
}

var ErrEmpty = errors.New("the feature file is empty")

type parser struct {
	lx     *lexer.Lexer
	path   string
	ast    *AST
	peeked *lexer.Token
}

func Parse(path string) (*Feature, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return (&parser{
		lx:   lexer.New(file),
		path: path,
		ast:  newAST(),
	}).parseFeature()
}

// reads tokens into AST and skips comments or new lines
func (p *parser) next() *lexer.Token {
	if p.ast.tail != nil && p.ast.tail.value.Type == lexer.EOF {
		return p.ast.tail.value // has reached EOF, do not record it more than once
	}
	tok := p.peek()
	p.ast.addTail(tok)
	p.peeked = nil
	return tok
}

// peaks into next token, skips comments or new lines
func (p *parser) peek() *lexer.Token {
	if p.peeked != nil {
		return p.peeked
	}

	for p.peeked = p.lx.Next(); p.peeked.OfType(lexer.COMMENT, lexer.NEW_LINE); p.peeked = p.lx.Next() {
		p.ast.addTail(p.peeked) // record comments and newlines
	}

	return p.peeked
}

func (p *parser) err(s string, l int) error {
	return fmt.Errorf("%s on %s:%d", s, p.path, l)
}

func (p *parser) parseFeature() (ft *Feature, err error) {

	ft = &Feature{Path: p.path, AST: p.ast}
	switch p.peek().Type {
	case lexer.EOF:
		return ft, ErrEmpty
	case lexer.TAGS:
		ft.Tags = p.parseTags()
	}

	tok := p.next()
	if tok.Type != lexer.FEATURE {
		return ft, p.err("expected a file to begin with a feature definition, but got '"+tok.Type.String()+"' instead", tok.Line)
	}
	ft.Title = tok.Value
	ft.Comment = tok.Comment

	var desc []string
	for ; p.peek().Type == lexer.TEXT; tok = p.next() {
		desc = append(desc, tok.Value)
	}
	ft.Description = strings.Join(desc, "\n")

	for tok = p.peek(); tok.Type != lexer.EOF; tok = p.peek() {
		// there may be a background
		if tok.Type == lexer.BACKGROUND {
			if ft.Background != nil {
				return ft, p.err("there can only be a single background section, but found another", tok.Line)
			}

			ft.Background = &Background{Comment: tok.Comment}
			p.next() // jump to background steps
			if ft.Background.Steps, err = p.parseSteps(); err != nil {
				return ft, err
			}
			tok = p.peek() // peek to scenario or tags
		}

		// there may be tags before scenario
		sc := &Scenario{Tags: ft.Tags}
		if tok.Type == lexer.TAGS {
			for _, t := range p.parseTags() {
				if !sc.Tags.Has(t) {
					sc.Tags = append(sc.Tags, t)
				}
			}
			tok = p.peek()
		}

		// there must be a scenario otherwise
		if tok.Type != lexer.SCENARIO {
			return ft, p.err("expected a scenario, but got '"+tok.Type.String()+"' instead", tok.Line)
		}

		sc.Title = tok.Value
		sc.Comment = tok.Comment
		p.next() // jump to scenario steps
		if sc.Steps, err = p.parseSteps(); err != nil {
			return ft, err
		}
		ft.Scenarios = append(ft.Scenarios, sc)
	}

	return ft, nil
}

func (p *parser) parseSteps() (steps []*Step, err error) {
	for tok := p.peek(); tok.OfType(allSteps...); tok = p.peek() {
		step := &Step{Text: tok.Value, Comment: tok.Comment}
		switch tok.Type {
		case lexer.GIVEN:
			step.Type = Given
		case lexer.WHEN:
			step.Type = When
		case lexer.THEN:
			step.Type = Then
		case lexer.AND, lexer.BUT:
			if len(steps) > 0 {
				step.Type = steps[len(steps)-1].Type
			} else {
				step.Type = Given
			}
		}

		p.next() // have read a peeked step
		if step.Text[len(step.Text)-1] == ':' {
			tok = p.peek()
			switch tok.Type {
			case lexer.PYSTRING:
				if err := p.parsePystring(step); err != nil {
					return steps, err
				}
			case lexer.TABLE_ROW:
				if err := p.parseTable(step); err != nil {
					return steps, err
				}
			default:
				return steps, p.err("pystring or table row was expected, but got: '"+tok.Type.String()+"' instead", tok.Line)
			}
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func (p *parser) parsePystring(s *Step) error {
	var tok *lexer.Token
	started := p.next() // skip the start of pystring
	var lines []string
	for tok = p.next(); !tok.OfType(lexer.EOF, lexer.PYSTRING); tok = p.next() {
		lines = append(lines, tok.Text)
	}
	if tok.Type == lexer.EOF {
		return fmt.Errorf("pystring which was opened on %s:%d was not closed", p.path, started.Line)
	}
	s.PyString = &PyString{Body: strings.Join(lines, "\n")}
	return nil
}

func (p *parser) parseTable(s *Step) error {
	s.Table = &Table{}
	for row := p.peek(); row.Type == lexer.TABLE_ROW; row = p.peek() {
		var cols []string
		for _, r := range strings.Split(strings.Trim(row.Value, "|"), "|") {
			cols = append(cols, strings.TrimFunc(r, unicode.IsSpace))
		}
		// ensure the same colum number for each row
		if len(s.Table.rows) > 0 && len(s.Table.rows[0]) != len(cols) {
			return p.err("table row has not the same number of columns compared to previous row", row.Line)
		}
		s.Table.rows = append(s.Table.rows, cols)
		p.next() // jump over the peeked token
	}
	return nil
}

func (p *parser) parseTags() (tags Tags) {
	for _, tag := range strings.Split(p.next().Value, " ") {
		t := Tag(strings.Trim(tag, "@ "))
		if len(t) > 0 && !tags.Has(t) {
			tags = append(tags, t)
		}
	}
	return
}

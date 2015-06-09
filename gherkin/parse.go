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

type Scenario struct {
	Title string
	Steps []*Step
	Tags  []Tag
}

type Background struct {
	Steps []*Step
}

type StepType string

const (
	Given StepType = "Given"
	When  StepType = "When"
	Then  StepType = "Then"
)

type Step struct {
	Text     string
	Type     StepType
	PyString *PyString
	Table    *Table
}

type Feature struct {
	Tags        []Tag
	Description string
	Title       string
	Background  *Background
	Scenarios   []*Scenario
	AST         *AST
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
	lx   *lexer.Lexer
	path string
	ast  *AST
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
	tok := p.lx.Next()
	p.ast.addTail(tok)
	if tok.OfType(lexer.COMMENT, lexer.NEW_LINE) {
		return p.next()
	}
	return tok
}

// peaks into next token, skips comments or new lines
func (p *parser) peek() *lexer.Token {
	if tok := p.lx.Peek(); !tok.OfType(lexer.COMMENT, lexer.NEW_LINE) {
		return tok
	}
	p.next()
	return p.peek()
}

func (p *parser) err(s string, l int) error {
	return fmt.Errorf("%s on %s:%d", s, p.path, l)
}

func (p *parser) parseFeature() (ft *Feature, err error) {
	var tok *lexer.Token = p.next()
	if tok.Type == lexer.EOF {
		return nil, ErrEmpty
	}

	ft = &Feature{}
	if tok.Type == lexer.TAGS {
		if p.peek().Type != lexer.FEATURE {
			return ft, p.err("tags must be a single line next to a feature definition", tok.Line)
		}
		ft.Tags = p.parseTags(tok.Value)
		tok = p.next()
	}

	if tok.Type != lexer.FEATURE {
		return ft, p.err("expected a file to begin with a feature definition, but got '"+tok.Type.String()+"' instead", tok.Line)
	}

	ft.Title = tok.Value
	var desc []string
	for ; p.peek().Type == lexer.TEXT; tok = p.next() {
		desc = append(desc, tok.Value)
	}
	ft.Description = strings.Join(desc, "\n")

	tok = p.next()
	for tok = p.next(); tok.Type != lexer.EOF; p.next() {
		// there may be a background
		if tok.Type == lexer.BACKGROUND {
			if ft.Background != nil {
				return ft, p.err("there can only be a single background section, but found another", tok.Line)
			}
			ft.Background = &Background{}
			if ft.Background.Steps, err = p.parseSteps(); err != nil {
				return ft, err
			}
			continue
		}
		// there may be tags before scenario
		sc := &Scenario{}
		if tok.Type == lexer.TAGS {
			sc.Tags, tok = p.parseTags(tok.Value), p.next()
		}

		// there must be a scenario otherwise
		if tok.Type != lexer.SCENARIO {
			return ft, p.err("expected a scenario, but got '"+tok.Type.String()+"' instead", tok.Line)
		}

		sc.Title = tok.Value
		if sc.Steps, err = p.parseSteps(); err != nil {
			return ft, err
		}
		ft.Scenarios = append(ft.Scenarios, sc)
	}

	ft.AST = p.ast
	return ft, nil
}

func (p *parser) parseSteps() (steps []*Step, err error) {
	for tok := p.peek(); tok.OfType(allSteps...); tok = p.peek() {
		p.next() // move over the step
		step := &Step{Text: tok.Value}
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

		if step.Text[len(step.Text)-1] == ':' {
			next := p.peek()
			switch next.Type {
			case lexer.PYSTRING:
				if err := p.parsePystring(step); err != nil {
					return steps, err
				}
			case lexer.TABLE_ROW:
				if err := p.parseTable(step); err != nil {
					return steps, err
				}
			default:
				return steps, p.err("pystring or table row was expected, but got: '"+next.Type.String()+"' instead", next.Line)
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

func (p *parser) parseTags(s string) (tags []Tag) {
	for _, tag := range strings.Split(s, " ") {
		t := strings.Trim(tag, "@ ")
		if len(t) > 0 {
			tags = append(tags, Tag(t))
		}
	}
	return
}

package gherkin

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/l3pp4rd/behat/gherkin/lexer"
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
	Text string
	Type StepType
}

type Feature struct {
	Tags        []Tag
	Description string
	Title       string
	Background  *Background
	Scenarios   []*Scenario
}

var steps = []lexer.TokenType{
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
	tok := p.lx.Next()
	p.ast.addTail(tok)
	if tok.OfType(lexer.COMMENT, lexer.NEW_LINE) {
		return p.next()
	}
	return tok
}

// peaks into next token, skips comments or new lines
func (p *parser) peek() *lexer.Token {
	if tok := p.lx.Peek(); tok.OfType(lexer.COMMENT, lexer.NEW_LINE) {
		p.lx.Next()
	}
	return p.peek()
}

func (p *parser) err(s string, l int) error {
	return fmt.Errorf("%s on %s:%d", s, p.path, l)
}

func (p *parser) parseFeature() (*Feature, error) {
	var tok *lexer.Token = p.next()
	if tok.Type == lexer.EOF {
		return nil, ErrEmpty
	}

	ft := &Feature{}
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
			ft.Background = p.parseBackground()
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
		p.parseSteps(sc)
		ft.Scenarios = append(ft.Scenarios, sc)
	}

	return ft, nil
}

func (p *parser) parseBackground() *Background {
	return nil
}

func (p *parser) parseSteps(s *Scenario) error {
	var tok *lexer.Token
	for ; p.peek().OfType(steps...); tok = p.next() {
		step := &Step{Text: tok.Value}
		switch tok.Type {
		case lexer.GIVEN:
			step.Type = Given
		case lexer.WHEN:
			step.Type = When
		case lexer.THEN:
			step.Type = Then
		case lexer.AND:
		case lexer.BUT:
			if len(s.Steps) > 0 {
				step.Type = s.Steps[len(s.Steps)-1].Type
			} else {
				step.Type = Given
			}
		}
		for ; p.peek().OfType(lexer.TEXT); tok = p.next() {
			step.Text += " " + tok.Value
		}
		// now look for pystring or table

		s.Steps = append(s.Steps, step)
		// return fmt.Errorf("A step was expected, but got: '%s' instead on %s:%d", tok.Type, "file", tok.Line)
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

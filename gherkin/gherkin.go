/*
Package gherkin is a gherkin language parser based on https://cucumber.io/docs/reference
specification. It parses a feature file into the it's structural representation. It also
creates an AST tree of gherkin Tokens read from the file.

With gherkin language you can describe your application behavior as features in
human-readable and machine friendly language.

For example, imagine you’re about to create the famous UNIX ls command.
Before you begin, you describe how the feature should work, see the example below..

Example:
	Feature: ls
	  In order to see the directory structure
	  As a UNIX user
	  I need to be able to list the current directory's contents

	  Scenario:
		Given I am in a directory "test"
		And I have a file named "foo"
		And I have a file named "bar"
		When I run "ls"
		Then I should get:
		  """
		  bar
		  foo
		  """

As a developer, your work is done as soon as you’ve made the ls command behave as
described in the Scenario.

To read the feature in the example above..

Example:
	package main

	import (
		"log"
		"os"

		"github.com/DATA-DOG/godog/gherkin"
	)

	func main() {
		feature, err := gherkin.Parse("ls.feature")
		switch {
		case err == gherkin.ErrEmpty:
			log.Println("the feature file is empty and does not describe any feature")
			return
		case err != nil:
			log.Println("the feature file is incorrect or could not be read:", err)
			os.Exit(1)
		}
		log.Println("have parsed a feature:", feature.Title, "with", len(feature.Scenarios), "scenarios")
	}

Now the feature is available in the structure.
*/
package gherkin

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// Tag is gherkin feature or scenario tag.
// it may be used to filter scenarios.
//
// tags may be set for a feature, in that case it will
// be merged with all scenario tags. or specifically
// to a single scenario
type Tag string

// Tags is an array of tags
type Tags []Tag

// Has checks whether the tag list has a tag
func (t Tags) Has(tag Tag) bool {
	for _, tg := range t {
		if tg == tag {
			return true
		}
	}
	return false
}

// Scenario describes the scenario details
//
// if Examples table is not nil, then it
// means that this is an outline scenario
// with a table of examples to be run for
// each and every row
//
// Scenario may have tags which later may
// be used to filter out or run specific
// initialization tasks
type Scenario struct {
	*Token
	Title    string
	Steps    []*Step
	Tags     Tags
	Examples *Table
	Feature  *Feature
}

// Background steps are run before every scenario
type Background struct {
	*Token
	Steps   []*Step
	Feature *Feature
}

// Step describes a Scenario or Background step
type Step struct {
	*Token
	Text       string
	Type       string
	PyString   *PyString
	Table      *Table
	Scenario   *Scenario
	Background *Background
}

// Feature describes the whole feature
type Feature struct {
	*Token
	Path        string
	Tags        Tags
	Description string
	Title       string
	Background  *Background
	Scenarios   []*Scenario
	AST         *AST
}

// PyString is a multiline text object used with step definition
type PyString struct {
	*Token
	Raw   string   // raw multiline string body
	Lines []string // trimmed lines
	Step  *Step
}

// String returns raw multiline string
func (p *PyString) String() string {
	return p.Raw
}

// Table is a row group object used with
// step definition or outline scenario
type Table struct {
	*Token
	OutlineScenario *Scenario
	Step            *Step
	rows            [][]string
}

var allSteps = []TokenType{
	GIVEN,
	WHEN,
	THEN,
	AND,
	BUT,
}

// ErrEmpty is returned in case if feature file
// is completely empty. May be ignored in some use cases
var ErrEmpty = errors.New("the feature file is empty")

type parser struct {
	lx     *lexer
	path   string
	ast    *AST
	peeked *Token
}

// Parse the feature file on the given path into
// the Feature struct
// Returns a Feature struct and error if there is any
func Parse(path string) (*Feature, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return (&parser{
		lx:   newLexer(file),
		path: path,
		ast:  newAST(),
	}).parseFeature()
}

// reads tokens into AST and skips comments or new lines
func (p *parser) next() *Token {
	if p.ast.tail != nil && p.ast.tail.value.Type == EOF {
		return p.ast.tail.value // has reached EOF, do not record it more than once
	}
	tok := p.peek()
	p.ast.addTail(tok)
	p.peeked = nil
	return tok
}

// peaks into next token, skips comments or new lines
func (p *parser) peek() *Token {
	if p.peeked != nil {
		return p.peeked
	}

	for p.peeked = p.lx.read(); p.peeked.OfType(COMMENT, NEW_LINE); p.peeked = p.lx.read() {
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
	case EOF:
		return ft, ErrEmpty
	case TAGS:
		ft.Tags = p.parseTags()
	}

	tok := p.next()
	if tok.Type != FEATURE {
		return ft, p.err("expected a file to begin with a feature definition, but got '"+tok.Type.String()+"' instead", tok.Line)
	}
	ft.Title = tok.Value
	ft.Token = tok

	var desc []string
	for ; p.peek().Type == TEXT; tok = p.next() {
		desc = append(desc, p.peek().Text)
	}
	ft.Description = strings.Join(desc, "\n")

	for tok = p.peek(); tok.Type != EOF; tok = p.peek() {
		// there may be a background
		if tok.Type == BACKGROUND {
			if ft.Background != nil {
				return ft, p.err("there can only be a single background section, but found another", tok.Line)
			}

			ft.Background = &Background{Token: tok, Feature: ft}
			p.next() // jump to background steps
			if ft.Background.Steps, err = p.parseSteps(); err != nil {
				return ft, err
			}
			for _, step := range ft.Background.Steps {
				step.Background = ft.Background
			}
			tok = p.peek() // peek to scenario or tags
		}

		// there may be tags before scenario
		var tags Tags
		tags = append(tags, ft.Tags...)
		if tok.Type == TAGS {
			for _, t := range p.parseTags() {
				if !tags.Has(t) {
					tags = append(tags, t)
				}
			}
			tok = p.peek()
		}

		// there must be a scenario or scenario outline otherwise
		if !tok.OfType(SCENARIO, SCENARIO_OUTLINE) {
			if tok.Type == EOF {
				return ft, nil // there may not be a scenario defined after background
			}
			return ft, p.err("expected a scenario or scenario outline, but got '"+tok.Type.String()+"' instead", tok.Line)
		}

		scenario, err := p.parseScenario()
		if err != nil {
			return ft, err
		}

		scenario.Tags = tags
		scenario.Feature = ft
		ft.Scenarios = append(ft.Scenarios, scenario)
	}

	return ft, nil
}

func (p *parser) parseScenario() (s *Scenario, err error) {
	tok := p.next()
	s = &Scenario{Title: tok.Value, Token: tok}
	if s.Steps, err = p.parseSteps(); err != nil {
		return s, err
	}
	for _, step := range s.Steps {
		step.Scenario = s
	}
	if examples := p.peek(); examples.Type == EXAMPLES {
		p.next() // jump over the peeked token
		peek := p.peek()
		if peek.Type != TABLE_ROW {
			return s, p.err(strings.Join([]string{
				"expected a table row,",
				"but got '" + peek.Type.String() + "' instead, for scenario outline examples",
			}, " "), examples.Line)
		}
		if s.Examples, err = p.parseTable(); err != nil {
			return s, err
		}
		s.Examples.OutlineScenario = s
	}
	return s, nil
}

func (p *parser) parseSteps() (steps []*Step, err error) {
	for tok := p.peek(); tok.OfType(allSteps...); tok = p.peek() {
		step := &Step{Text: tok.Value, Token: tok}

		p.next() // have read a peeked step
		if step.Text[len(step.Text)-1] == ':' {
			tok = p.peek()
			switch tok.Type {
			case PYSTRING:
				if step.PyString, err = p.parsePystring(); err != nil {
					return steps, err
				}
				step.PyString.Step = step
			case TABLE_ROW:
				if step.Table, err = p.parseTable(); err != nil {
					return steps, err
				}
				step.Table.Step = step
			default:
				return steps, p.err("pystring or table row was expected, but got: '"+tok.Type.String()+"' instead", tok.Line)
			}
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func (p *parser) parsePystring() (*PyString, error) {
	var tok *Token
	started := p.next() // skip the start of pystring
	var lines, trimmed []string
	for tok = p.next(); !tok.OfType(EOF, PYSTRING); tok = p.next() {
		lines = append(lines, tok.Text)
		trimmed = append(trimmed, strings.TrimSpace(tok.Text))
	}
	if tok.Type == EOF {
		return nil, fmt.Errorf("pystring which was opened on %s:%d was not closed", p.path, started.Line)
	}
	return &PyString{
		Raw:   strings.Join(lines, "\n"),
		Lines: trimmed,
	}, nil
}

func (p *parser) parseTable() (*Table, error) {
	tbl := &Table{}
	for row := p.peek(); row.Type == TABLE_ROW; row = p.peek() {
		var cols []string
		for _, r := range strings.Split(strings.Trim(row.Value, "|"), "|") {
			cols = append(cols, strings.TrimFunc(r, unicode.IsSpace))
		}
		// ensure the same colum number for each row
		if len(tbl.rows) > 0 && len(tbl.rows[0]) != len(cols) {
			return tbl, p.err("table row has not the same number of columns compared to previous row", row.Line)
		}
		tbl.rows = append(tbl.rows, cols)
		p.next() // jump over the peeked token
	}
	return tbl, nil
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

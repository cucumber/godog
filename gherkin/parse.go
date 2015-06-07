package gherkin

import (
	"errors"
	"io"
	"strings"

	"github.com/l3pp4rd/behat/gherkin/lexer"
)

type Tag string

type Scenario struct {
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

var ErrNotFeature = errors.New("expected a file to begin with a feature definition")
var ErrEmpty = errors.New("the feature file is empty")
var ErrTagsNextToFeature = errors.New("tags must be a single line next to a feature definition")
var ErrSingleBackground = errors.New("there can only be a single background section")

type parser struct {
	lx *lexer.Lexer
}

func Parse(r io.Reader) (*Feature, error) {
	return (parser{lx: lexer.New(r)}).parseFeature()
}

func (p parser) parseFeature() (*Feature, error) {
	var tok *lexer.Token = p.lx.Next(lexer.COMMENT, lexer.NEW_LINE)
	if tok.Type == lexer.EOF {
		return nil, ErrEmpty
	}

	ft := &Feature{}
	if tok.Type == lexer.TAGS {
		if p.lx.Peek().Type != lexer.FEATURE {
			return ft, ErrTagsNextToFeature
		}
		ft.Tags = p.parseTags(tok.Value)
		tok = p.lx.Next()
	}

	if tok.Type != lexer.FEATURE {
		return ft, ErrNotFeature
	}

	ft.Title = tok.Value
	var desc []string
	for ; p.lx.Peek().Type == lexer.TEXT; tok = p.lx.Next() {
		desc = append(desc, tok.Value)
	}
	ft.Description = strings.Join(desc, "\n")

	tok = p.lx.Next(lexer.COMMENT, lexer.NEW_LINE)
	for ; tok.Type != lexer.EOF; p.lx.Next(lexer.COMMENT, lexer.NEW_LINE) {
		if tok.Type == lexer.BACKGROUND {
			if ft.Background != nil {
				return ft, ErrSingleBackground
			}
			ft.Background = p.parseBackground()
			continue
		}
	}

	return ft, nil
}

func (p parser) parseBackground() *Background {
	return nil
}

func (p parser) parseTags(s string) (tags []Tag) {
	for _, tag := range strings.Split(s, " ") {
		t := strings.Trim(tag, "@ ")
		if len(t) > 0 {
			tags = append(tags, Tag(t))
		}
	}
	return
}

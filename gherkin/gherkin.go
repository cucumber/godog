package gherkin

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Parser interface {
	StopAtFirstError(b bool)
	Parse(s Scanner, m Matcher) (err error)
}

/*
The scanner reads a gherkin doc (typically read from a .feature file) and creates a token for
each line. The tokens are passed to the parser, which outputs an AST (Abstract Syntax Tree).

If the scanner sees a # language header, it will reconfigure itself dynamically to look for
Gherkin keywords for the associated language. The keywords are defined in gherkin-languages.json.
*/
type Scanner interface {
	Scan() (line *Line, atEof bool, err error)
}

type Builder interface {
	Build(*Token) (bool, error)
	StartRule(RuleType) (bool, error)
	EndRule(RuleType) (bool, error)
	Reset()
}

type Token struct {
	Type           TokenType
	Keyword        string
	Text           string
	Items          []*LineSpan
	GherkinDialect string
	Indent         string
	Location       *Location
}

func (t *Token) IsEOF() bool {
	return t.Type == TokenType_EOF
}
func (t *Token) String() string {
	return fmt.Sprintf("%s: %s/%s", t.Type.Name(), t.Keyword, t.Text)
}

type LineSpan struct {
	Column int
	Text   string
}

func (l *LineSpan) String() string {
	return fmt.Sprintf("%d:%s", l.Column, l.Text)
}

type parser struct {
	builder          Builder
	stopAtFirstError bool
}

func NewParser(b Builder) Parser {
	return &parser{
		builder: b,
	}
}

func (p *parser) StopAtFirstError(b bool) {
	p.stopAtFirstError = b
}

func NewScanner(r io.Reader) Scanner {
	return &scanner{
		s:    bufio.NewScanner(r),
		line: 0,
	}
}

type scanner struct {
	s    *bufio.Scanner
	line int
}

func (t *scanner) Scan() (line *Line, atEof bool, err error) {
	scanning := t.s.Scan()
	if !scanning {
		err = t.s.Err()
		if err == nil {
			atEof = true
		}
	}
	if err == nil {
		t.line += 1
		str := t.s.Text()
		line = &Line{str, t.line, strings.TrimLeft(str, " \t"), atEof}
	}
	return
}

type Line struct {
	LineText        string
	LineNumber      int
	TrimmedLineText string
	AtEof           bool
}

func (g *Line) Indent() int {
	return len(g.LineText) - len(g.TrimmedLineText)
}

func (g *Line) IsEmpty() bool {
	return len(g.TrimmedLineText) == 0
}

func (g *Line) IsEof() bool {
	return g.AtEof
}

func (g *Line) StartsWith(prefix string) bool {
	return strings.HasPrefix(g.TrimmedLineText, prefix)
}

func ParseFeature(in io.Reader) (feature *Feature, err error) {

	builder := NewAstBuilder()
	parser := NewParser(builder)
	parser.StopAtFirstError(false)
	matcher := NewMatcher(GherkinDialectsBuildin())

	scanner := NewScanner(in)

	err = parser.Parse(scanner, matcher)

	return builder.GetFeature(), err
}

//ParseStruct convert a feature data table to struct using tags
// "godog" configured in struct
func ParseStruct(dataTable *DataTable, i interface{}) error {
	var d decode
	if err := d.init(dataTable, i); err != nil {
		return err
	}
	return d.unmarshal(i)
}

type decode struct {
	Value     reflect.Value
	Type      reflect.Type
	DataTable map[string]string
	NumField  int
	Tag       string
}

func (d *decode) insertValue(i interface{}, index int, value interface{}) error {
	t := d.Type.Field(index).Type.String()
	switch t {
	case "string":
		d.Value.Elem().Field(index).SetString(fmt.Sprintf("%s", value))
		return nil
	case "int":
		v, err := strconv.ParseInt(fmt.Sprintf("%s", value), 10, 64)
		if err != nil {
			return err
		}
		d.Value.Elem().Field(index).SetInt(v)
		return nil
	case "float64":
		v, err := strconv.ParseFloat(fmt.Sprintf("%s", value), 64)
		if err != nil {
			return err
		}
		d.Value.Elem().Field(index).SetFloat(v)
		return nil
	case "float32":
		v, err := strconv.ParseFloat(fmt.Sprintf("%s", value), 32)
		if err != nil {
			return err
		}
		d.Value.Elem().Field(index).SetFloat(v)
		return nil
	case "bool":
		v, err := strconv.ParseBool(fmt.Sprintf("%s", value))
		if err != nil {
			return err
		}
		d.Value.Elem().Field(index).SetBool(v)
		return nil
	default:
		return fmt.Errorf("invalid type %v of field in struct", t)
	}
}

func (d *decode) init(dataTable *DataTable, i interface{}) error {
	d.Value = reflect.ValueOf(i)
	d.Type = reflect.Indirect(d.Value).Type()
	m, err := ParseMap(dataTable)
	if err != nil {
		return err
	}
	d.DataTable = m
	d.NumField = d.Type.NumField()
	d.Tag = "godog"
	return nil
}

func (d *decode) unmarshal(v interface{}) error {
	for i := 0; i < d.NumField; i++ {
		key := d.Type.Field(i).Tag.Get(d.Tag)
		if key == "" {
			if value, ok := d.DataTable[d.Type.Field(i).Name]; ok {
				if err := d.insertValue(v, i, value); err != nil {
					return err
				}
			}
		} else {
			if value, ok := d.DataTable[key]; ok {
				if err := d.insertValue(v, i, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//ParseMap convert data table of feature to map
//data table need to have exactly two rows
func ParseMap(dataTable *DataTable) (map[string]string, error) {
	rows := dataTable.Rows
	if len(rows) != 2 {
		return nil, fmt.Errorf("data table need to have two rows")
	}
	m := map[string]string{}
	for _, row := range rows[1:] {
		cells := row.Cells
		for i, cell := range cells {
			m[rows[0].Cells[i].Value] = cell.Value
		}
	}
	return m, nil
}

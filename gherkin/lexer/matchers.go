package lexer

import "regexp"

var matchers = map[string]*regexp.Regexp{
	"feature":    regexp.MustCompile("^(\\s*)Feature:\\s*(.*)"),
	"scenario":   regexp.MustCompile("^(\\s*)Scenario:\\s*(.*)"),
	"background": regexp.MustCompile("^(\\s*)Background:"),
	"step":       regexp.MustCompile("^(\\s*)(Given|When|Then|And|But)\\s+(.+)"),
	"comment":    regexp.MustCompile("^(\\s*)#(.+)"),
	"pystring":   regexp.MustCompile("^(\\s*)\\\"\\\"\\\""),
	"tags":       regexp.MustCompile("^(\\s*)(@.+)"),
	"table_row":  regexp.MustCompile("^(\\s*)(\\|.+)"),
}

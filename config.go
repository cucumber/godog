package godog

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

type registeredFormatter struct {
	name string
	fmt  Formatter
}

var formatters []*registeredFormatter

// RegisterFormatter registers a feature suite output
// Formatter for its given name
func RegisterFormatter(name string, f Formatter) {
	formatters = append(formatters, &registeredFormatter{
		name: name,
		fmt:  f,
	})
}

var cfg config

func init() {
	// @TODO: colorize flag help output
	flag.StringVar(&cfg.featuresPath, "features", "features", "Path to feature files")
	flag.StringVar(&cfg.formatterName, "formatter", "pretty", "Formatter name")
}

type config struct {
	featuresPath  string
	formatterName string
}

func (c config) validate() error {
	// feature path
	inf, err := os.Stat(c.featuresPath)
	if err != nil {
		return err
	}
	if !inf.IsDir() {
		return fmt.Errorf("feature path \"%s\" is not a directory.", c.featuresPath)
	}

	// formatter
	var found bool
	var names []string
	for _, f := range formatters {
		if f.name == c.formatterName {
			found = true
			break
		}
		names = append(names, f.name)
	}

	if !found {
		return fmt.Errorf(`unregistered formatter name: "%s", use one of: %s`, c.formatterName, strings.Join(names, ", "))
	}
	return nil
}

func (c config) features() (lst []*gherkin.Feature, err error) {
	return lst, filepath.Walk(cfg.featuresPath, func(p string, f os.FileInfo, err error) error {
		if err == nil && !f.IsDir() && strings.HasSuffix(p, ".feature") {
			ft, err := gherkin.Parse(p)
			if err != nil {
				return err
			}
			lst = append(lst, ft)
		}
		return err
	})
}

func (c config) formatter() Formatter {
	return &pretty{}
}

func fatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

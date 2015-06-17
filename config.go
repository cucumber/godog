package godog

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

type registeredFormatter struct {
	name        string
	fmt         Formatter
	description string
}

var formatters []*registeredFormatter

// RegisterFormatter registers a feature suite output
// Formatter as the name and descriptiongiven.
// Formatter is used to represent suite output
func RegisterFormatter(name, description string, f Formatter) {
	formatters = append(formatters, &registeredFormatter{
		name:        name,
		fmt:         f,
		description: description,
	})
}

var cfg *config

func s(n int) string {
	return strings.Repeat(" ", n)
}

func init() {
	cfg = &config{}

	flag.StringVar(&cfg.format, "format", "pretty", "")
	flag.StringVar(&cfg.format, "f", "pretty", "")
	flag.Usage = func() {
		// prints an option or argument with a description, or only description
		opt := func(name, desc string) string {
			if len(name) > 0 {
				name += ":"
			}
			return s(2) + cl(name, green) + s(30-len(name)) + desc
		}

		// --- GENERAL ---
		fmt.Println(cl("Usage:", yellow))
		fmt.Println(s(2) + "godog [options] [<paths>]\n")

		// --- ARGUMENTS ---
		fmt.Println(cl("Arguments:", yellow))
		// --> paths
		fmt.Println(opt("paths", "Optional path(s) to execute. Can be:"))
		fmt.Println(opt("", s(4)+"- dir "+cl("(features/)", yellow)))
		fmt.Println(opt("", s(4)+"- feature "+cl("(*.feature)", yellow)))
		fmt.Println(opt("", s(4)+"- scenario at specific line "+cl("(*.feature:10)", yellow)))
		fmt.Println(opt("", "If no paths are listed, suite tries "+cl("features", yellow)+" path by default."))
		fmt.Println("")

		// --- OPTIONS ---
		fmt.Println(cl("Options:", yellow))
		// --> format
		fmt.Println(opt("-f, --format=pretty", "How to format tests output. Available formats:"))
		for _, f := range formatters {
			fmt.Println(opt("", s(4)+"- "+cl(f.name, yellow)+": "+f.description))
		}
		fmt.Println("")
	}
}

type config struct {
	paths  []string
	format string
}

func (c *config) validate() error {
	c.paths = flag.Args()
	// check the default path
	if len(c.paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			c.paths = []string{"features"}
		}
	}
	// formatter
	var found bool
	var names []string
	for _, f := range formatters {
		if f.name == c.format {
			found = true
			break
		}
		names = append(names, f.name)
	}

	if !found {
		return fmt.Errorf(`unregistered formatter name: "%s", use one of: %s`, c.format, strings.Join(names, ", "))
	}
	return nil
}

func (c *config) features() (lst []*gherkin.Feature, err error) {
	for _, pat := range c.paths {
		// check if line number is specified
		parts := strings.Split(pat, ":")
		path := parts[0]
		line := -1
		if len(parts) > 1 {
			line, err = strconv.Atoi(parts[1])
			if err != nil {
				return lst, fmt.Errorf("line number should follow after colon path delimiter")
			}
		}
		// parse features
		err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
			if err == nil && !f.IsDir() && strings.HasSuffix(p, ".feature") {
				ft, err := gherkin.Parse(p)
				switch {
				case err == gherkin.ErrEmpty:
					// its ok, just skip it
				case err != nil:
					return err
				default:
					lst = append(lst, ft)
				}
				// filter scenario by line number
				if line != -1 {
					var scenarios []*gherkin.Scenario
					for _, s := range ft.Scenarios {
						if s.Token.Line == line {
							scenarios = append(scenarios, s)
							break
						}
					}
					ft.Scenarios = scenarios
				}
			}
			return err
		})
		// check error
		switch {
		case os.IsNotExist(err):
			return lst, fmt.Errorf(`feature path "%s" is not available`, path)
		case os.IsPermission(err):
			return lst, fmt.Errorf(`feature path "%s" is not accessible`, path)
		case err != nil:
			return lst, err
		}
	}
	return
}

func (c *config) formatter() (f Formatter) {
	for _, fmt := range formatters {
		if fmt.name == cfg.format {
			return fmt.fmt
		}
	}
	panic("formatter name had to be validated")
}

func fatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

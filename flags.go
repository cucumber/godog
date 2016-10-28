package godog

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/DATA-DOG/godog/colors"
)

var descFeaturesArgument = "Optional feature(s) to run. Can be:\n" +
	s(4) + "- dir " + colors.Yellow("(features/)") + "\n" +
	s(4) + "- feature " + colors.Yellow("(*.feature)") + "\n" +
	s(4) + "- scenario at specific line " + colors.Yellow("(*.feature:10)") + "\n" +
	"If no feature paths are listed, suite tries " + colors.Yellow("features") + " path by default.\n"

var descConcurrencyOption = "Run the test suite with concurrency level:\n" +
	s(4) + "- " + colors.Yellow(`= 1`) + ": supports all types of formats.\n" +
	s(4) + "- " + colors.Yellow(`>= 2`) + ": only supports " + colors.Yellow("progress") + ". Note, that\n" +
	s(4) + "your context needs to support parallel execution."

var descTagsOption = "Filter scenarios by tags. Expression can be:\n" +
	s(4) + "- " + colors.Yellow(`"@wip"`) + ": run all scenarios with wip tag\n" +
	s(4) + "- " + colors.Yellow(`"~@wip"`) + ": exclude all scenarios with wip tag\n" +
	s(4) + "- " + colors.Yellow(`"@wip && ~@new"`) + ": run wip scenarios, but exclude new\n" +
	s(4) + "- " + colors.Yellow(`"@wip,@undone"`) + ": run wip or undone scenarios"

// FlagSet allows to manage flags by external suite runner
func FlagSet(w io.Writer, format, tags *string, defs, sof, noclr *bool, cr *int) *flag.FlagSet {
	descFormatOption := "How to format tests output. Available formats:\n"
	// @TODO: sort by name
	for name, desc := range AvailableFormatters() {
		descFormatOption += s(4) + "- " + colors.Yellow(name) + ": " + desc + "\n"
	}
	descFormatOption = strings.TrimSpace(descFormatOption)

	set := flag.NewFlagSet("godog", flag.ExitOnError)
	set.StringVar(format, "format", "pretty", descFormatOption)
	set.StringVar(format, "f", "pretty", descFormatOption)
	set.StringVar(tags, "tags", "", descTagsOption)
	set.StringVar(tags, "t", "", descTagsOption)
	set.IntVar(cr, "concurrency", 1, descConcurrencyOption)
	set.IntVar(cr, "c", 1, descConcurrencyOption)
	set.BoolVar(defs, "definitions", false, "Print all available step definitions.")
	set.BoolVar(defs, "d", false, "Print all available step definitions.")
	set.BoolVar(sof, "stop-on-failure", false, "Stop processing on first failed scenario.")
	set.BoolVar(noclr, "no-colors", false, "Disable ansi colors.")
	set.Usage = usage(set, w)
	return set
}

type flagged struct {
	short, long, descr, dflt string
}

func (f *flagged) name() string {
	var name string
	switch {
	case len(f.short) > 0 && len(f.long) > 0:
		name = fmt.Sprintf("-%s, --%s", f.short, f.long)
	case len(f.long) > 0:
		name = fmt.Sprintf("--%s", f.long)
	case len(f.short) > 0:
		name = fmt.Sprintf("-%s", f.short)
	}
	if f.dflt != "true" && f.dflt != "false" {
		name += "=" + f.dflt
	}
	return name
}

func usage(set *flag.FlagSet, w io.Writer) func() {
	return func() {
		var list []*flagged
		var longest int
		set.VisitAll(func(f *flag.Flag) {
			var fl *flagged
			for _, flg := range list {
				if flg.descr == f.Usage {
					fl = flg
					break
				}
			}
			if nil == fl {
				fl = &flagged{
					dflt:  f.DefValue,
					descr: f.Usage,
				}
				list = append(list, fl)
			}
			if len(f.Name) > 2 {
				fl.long = f.Name
			} else {
				fl.short = f.Name
			}
		})

		for _, f := range list {
			if len(f.name()) > longest {
				longest = len(f.name())
			}
		}

		// prints an option or argument with a description, or only description
		opt := func(name, desc string) string {
			var ret []string
			lines := strings.Split(desc, "\n")
			ret = append(ret, s(2)+colors.Green(name)+s(longest+2-len(name))+lines[0])
			if len(lines) > 1 {
				for _, ln := range lines[1:] {
					ret = append(ret, s(2)+s(longest+2)+ln)
				}
			}
			return strings.Join(ret, "\n")
		}

		// --- GENERAL ---
		fmt.Fprintln(w, colors.Yellow("Usage:"))
		fmt.Printf(s(2) + "godog [options] [<features>]\n\n")
		// description
		fmt.Fprintln(w, "Builds a test package and runs given feature files.")
		fmt.Fprintf(w, "Command should be run from the directory of tested package and contain buildable go source.\n\n")

		// --- ARGUMENTS ---
		fmt.Fprintln(w, colors.Yellow("Arguments:"))
		// --> features
		fmt.Fprintln(w, opt("features", descFeaturesArgument))

		// --- OPTIONS ---
		fmt.Fprintln(w, colors.Yellow("Options:"))
		for _, f := range list {
			fmt.Fprintln(w, opt(f.name(), f.descr))
		}
		fmt.Fprintln(w, "")
	}
}

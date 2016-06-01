package godog

import (
	"flag"
	"fmt"
)

// FlagSet allows to manage flags by external suite runner
func FlagSet(format, tags *string, defs, sof, vers *bool, cl *int) *flag.FlagSet {
	set := flag.NewFlagSet("godog", flag.ExitOnError)
	set.StringVar(format, "format", "pretty", "")
	set.StringVar(format, "f", "pretty", "")
	set.StringVar(tags, "tags", "", "")
	set.StringVar(tags, "t", "", "")
	set.IntVar(cl, "concurrency", 1, "")
	set.IntVar(cl, "c", 1, "")
	set.BoolVar(defs, "definitions", false, "")
	set.BoolVar(defs, "d", false, "")
	set.BoolVar(sof, "stop-on-failure", false, "")
	set.BoolVar(vers, "version", false, "")
	set.Usage = usage
	return set
}

func usage() {
	// prints an option or argument with a description, or only description
	opt := func(name, desc string) string {
		if len(name) > 0 {
			name += ":"
		}
		return s(2) + cl(name, green) + s(22-len(name)) + desc
	}

	// --- GENERAL ---
	fmt.Println(cl("Usage:", yellow))
	fmt.Printf(s(2) + "godog [options] [<features>]\n\n")
	// description
	fmt.Println("Builds a test package and runs given feature files.")
	fmt.Printf("Command should be run from the directory of tested package and contain buildable go source.\n\n")

	// --- ARGUMENTS ---
	fmt.Println(cl("Arguments:", yellow))
	// --> paths
	fmt.Println(opt("features", "Optional feature(s) to run. Can be:"))
	fmt.Println(opt("", s(4)+"- dir "+cl("(features/)", yellow)))
	fmt.Println(opt("", s(4)+"- feature "+cl("(*.feature)", yellow)))
	fmt.Println(opt("", s(4)+"- scenario at specific line "+cl("(*.feature:10)", yellow)))
	fmt.Println(opt("", "If no feature paths are listed, suite tries "+cl("features", yellow)+" path by default."))
	fmt.Println("")

	// --- OPTIONS ---
	fmt.Println(cl("Options:", yellow))
	// --> step definitions
	fmt.Println(opt("-d, --definitions", "Print all available step definitions."))
	// --> concurrency
	fmt.Println(opt("-c, --concurrency=1", "Run the test suite with concurrency level:"))
	fmt.Println(opt("", s(4)+"- "+cl(`= 1`, yellow)+": supports all types of formats."))
	fmt.Println(opt("", s(4)+"- "+cl(`>= 2`, yellow)+": only supports "+cl("progress", yellow)+". Note, that"))
	fmt.Println(opt("", s(4)+"your context needs to support parallel execution."))
	// --> format
	fmt.Println(opt("-f, --format=pretty", "How to format tests output. Available formats:"))
	for _, f := range formatters {
		fmt.Println(opt("", s(4)+"- "+cl(f.name, yellow)+": "+f.description))
	}
	// --> tags
	fmt.Println(opt("-t, --tags", "Filter scenarios by tags. Expression can be:"))
	fmt.Println(opt("", s(4)+"- "+cl(`"@wip"`, yellow)+": run all scenarios with wip tag"))
	fmt.Println(opt("", s(4)+"- "+cl(`"~@wip"`, yellow)+": exclude all scenarios with wip tag"))
	fmt.Println(opt("", s(4)+"- "+cl(`"@wip && ~@new"`, yellow)+": run wip scenarios, but exclude new"))
	fmt.Println(opt("", s(4)+"- "+cl(`"@wip,@undone"`, yellow)+": run wip or undone scenarios"))
	// --> stop on failure
	fmt.Println(opt("--stop-on-failure", "Stop processing on first failed scenario."))
	// --> version
	fmt.Println(opt("--version", "Show current "+cl("godog", yellow)+" version."))
	fmt.Println("")
}

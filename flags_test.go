package godog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/godog/colors"
)

func TestFlagsShouldRandomizeAndGenerateSeed(t *testing.T) {
	var opt Options
	flags := FlagSet(&opt)
	if err := flags.Parse([]string{"--random"}); err != nil {
		t.Fatalf("unable to parse flags: %v", err)
	}

	if opt.Randomize <= 0 {
		t.Fatal("should have generated random seed")
	}
}

func TestFlagsShouldRandomizeByGivenSeed(t *testing.T) {
	var opt Options
	flags := FlagSet(&opt)
	if err := flags.Parse([]string{"--random=123"}); err != nil {
		t.Fatalf("unable to parse flags: %v", err)
	}

	if opt.Randomize != 123 {
		t.Fatalf("expected random seed to be: 123, but it was: %d", opt.Randomize)
	}
}

func TestFlagsShouldParseFormat(t *testing.T) {
	cases := map[string][]string{
		"pretty":   {},
		"progress": {"-f", "progress"},
		"junit":    {"-f=junit"},
		"custom":   {"--format", "custom"},
		"cust":     {"--format=cust"},
	}

	for format, args := range cases {
		var opt Options
		flags := FlagSet(&opt)
		if err := flags.Parse(args); err != nil {
			t.Fatalf("unable to parse flags: %v", err)
		}

		if opt.Format != format {
			t.Fatalf("expected format: %s, but it was: %s", format, opt.Format)
		}
	}
}

func TestFlagsUsageShouldIncludeFormatDescriptons(t *testing.T) {
	var buf bytes.Buffer
	output := colors.Uncolored(&buf)

	// register some custom formatter
	Format("custom", "custom format description", junitFunc)

	var opt Options
	flags := FlagSet(&opt)
	usage(flags, output)()

	out := buf.String()

	for name, desc := range AvailableFormatters() {
		match := fmt.Sprintf("%s: %s\n", name, desc)
		if idx := strings.Index(out, match); idx == -1 {
			t.Fatalf("could not locate format: %s description in flag usage", name)
		}
	}
}

package godog

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/formatters"
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
	Format("custom", "custom format description", formatters.JUnitFormatterFunc)

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

func TestBindFlagsShouldRespectFlagDefaults(t *testing.T) {
	opts := Options{}

	BindFlags("flagDefaults.", flag.CommandLine, &opts)

	if opts.Format != "pretty" {
		t.Fatalf("expected Format: pretty, but it was: %s", opts.Format)
	}
	if opts.Tags != "" {
		t.Fatalf("expected Tags: '', but it was: %s", opts.Tags)
	}
	if opts.Concurrency != 1 {
		t.Fatalf("expected Concurrency: 1, but it was: %d", opts.Concurrency)
	}
	if opts.ShowStepDefinitions {
		t.Fatalf("expected ShowStepDefinitions: false, but it was: %t", opts.ShowStepDefinitions)
	}
	if opts.StopOnFailure {
		t.Fatalf("expected StopOnFailure: false, but it was: %t", opts.StopOnFailure)
	}
	if opts.Strict {
		t.Fatalf("expected Strict: false, but it was: %t", opts.Strict)
	}
	if opts.NoColors {
		t.Fatalf("expected NoColors: false, but it was: %t", opts.NoColors)
	}
	if opts.Randomize != 0 {
		t.Fatalf("expected Randomize: 0, but it was: %d", opts.Randomize)
	}
}

func TestBindFlagsShouldRespectOptDefaults(t *testing.T) {
	opts := Options{
		Format:              "progress",
		Tags:                "test",
		Concurrency:         2,
		ShowStepDefinitions: true,
		StopOnFailure:       true,
		Strict:              true,
		NoColors:            true,
		Randomize:           int64(7),
	}

	flagSet := flag.FlagSet{}

	BindFlags("optDefaults.", &flagSet, &opts)

	if opts.Format != "progress" {
		t.Fatalf("expected Format: progress, but it was: %s", opts.Format)
	}
	if opts.Tags != "test" {
		t.Fatalf("expected Tags: 'test', but it was: %s", opts.Tags)
	}
	if opts.Concurrency != 2 {
		t.Fatalf("expected Concurrency: 2, but it was: %d", opts.Concurrency)
	}
	if !opts.ShowStepDefinitions {
		t.Fatalf("expected ShowStepDefinitions: true, but it was: %t", opts.ShowStepDefinitions)
	}
	if !opts.StopOnFailure {
		t.Fatalf("expected StopOnFailure: true, but it was: %t", opts.StopOnFailure)
	}
	if !opts.Strict {
		t.Fatalf("expected Strict: true, but it was: %t", opts.Strict)
	}
	if !opts.NoColors {
		t.Fatalf("expected NoColors: true, but it was: %t", opts.NoColors)
	}
	if opts.Randomize != 7 {
		t.Fatalf("expected Randomize: 7, but it was: %d", opts.Randomize)
	}
}

func TestBindFlagsShouldRespectFlagOverrides(t *testing.T) {
	opts := Options{
		Format:              "progress",
		Tags:                "test",
		Concurrency:         2,
		ShowStepDefinitions: true,
		StopOnFailure:       true,
		Strict:              true,
		NoColors:            true,
		Randomize:           11,
	}
	flagSet := flag.FlagSet{}

	BindFlags("optOverrides.", &flagSet, &opts)

	flagSet.Parse([]string{
		"--optOverrides.format=junit",
		"--optOverrides.tags=test2",
		"--optOverrides.concurrency=3",
		"--optOverrides.definitions=false",
		"--optOverrides.stop-on-failure=false",
		"--optOverrides.strict=false",
		"--optOverrides.no-colors=false",
		"--optOverrides.random=2",
	})

	if opts.Format != "junit" {
		t.Fatalf("expected Format: junit, but it was: %s", opts.Format)
	}
	if opts.Tags != "test2" {
		t.Fatalf("expected Tags: 'test2', but it was: %s", opts.Tags)
	}
	if opts.Concurrency != 3 {
		t.Fatalf("expected Concurrency: 3, but it was: %d", opts.Concurrency)
	}
	if opts.ShowStepDefinitions {
		t.Fatalf("expected ShowStepDefinitions: true, but it was: %t", opts.ShowStepDefinitions)
	}
	if opts.StopOnFailure {
		t.Fatalf("expected StopOnFailure: true, but it was: %t", opts.StopOnFailure)
	}
	if opts.Strict {
		t.Fatalf("expected Strict: true, but it was: %t", opts.Strict)
	}
	if opts.NoColors {
		t.Fatalf("expected NoColors: true, but it was: %t", opts.NoColors)
	}
	if opts.Randomize != 2 {
		t.Fatalf("expected Randomize: 2, but it was: %d", opts.Randomize)
	}
}

package godog_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog"
)

func Test_FindFmt(t *testing.T) {
	cases := map[string]bool{
		"cucumber": true,
		"custom":   true, // is available for test purposes only
		"events":   true,
		"junit":    true,
		"pretty":   true,
		"progress": true,
		"unknown":  false,
		"undef":    false,
	}

	for name, expected := range cases {
		t.Run(
			name,
			func(t *testing.T) {
				actual := godog.FindFmt(name)

				if expected {
					assert.NotNilf(t, actual, "expected %s formatter should be available", name)
				} else {
					assert.Nilf(t, actual, "expected %s formatter should be available", name)
				}
			},
		)
	}
}

func Test_AvailableFormatters(t *testing.T) {
	expected := map[string]string{
		"cucumber": "Produces cucumber JSON format output.",
		"custom":   "custom format description", // is available for test purposes only
		"events":   "Produces JSON event stream, based on spec: 0.1.0.",
		"junit":    "Prints junit compatible xml to stdout",
		"pretty":   "Prints every feature with runtime statuses.",
		"progress": "Prints a character per step.",
	}

	actual := godog.AvailableFormatters()
	assert.Equal(t, expected, actual)
}

func Test_Format(t *testing.T) {
	actual := godog.FindFmt("Test_Format")
	require.Nil(t, actual)

	godog.Format("Test_Format", "...", testFormatterFunc)
	actual = godog.FindFmt("Test_Format")

	assert.NotNil(t, actual)
}

func testFormatterFunc(suiteName string, out io.Writer, snippetFunc string) godog.Formatter {
	return nil
}

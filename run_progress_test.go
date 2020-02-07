package godog

import (
	"fmt"
	"testing"
)

func TestSucceedRun(t *testing.T) {
	const format = "progress"

	// Will test concurrency setting 0, 1, 2 and 3.
	for concurrency := range make([]int, 4) {
		t.Run(
			fmt.Sprintf("%s_concurrency_%d", format, concurrency),
			func(t *testing.T) {
				testSucceedRun(t, format, concurrency, expectedOutputProgress)
			},
		)
	}
}

const expectedOutputProgress = `...................................................................... 70
...................................................................... 140
...................................................................... 210
.......................................                                249


60 scenarios (60 passed)
249 steps (249 passed)
0s`

package godog

import "testing"

func TestShouldFindFormatter(t *testing.T) {
	cases := map[string]bool{
		"progress": true, // true means should be available
		"unknown":  false,
		"junit":    true,
		"cucumber": true,
		// "pretty":   true,
		"custom": true, // is available for test purposes only
		"undef":  false,
	}

	for name, shouldFind := range cases {
		actual := FindFmt(name)
		if actual == nil && shouldFind {
			t.Fatalf("expected %s formatter should be available", name)
		}
		if actual != nil && !shouldFind {
			t.Fatalf("expected %s formatter should not be available", name)
		}
	}
}

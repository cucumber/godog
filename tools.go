//go:build tools

// Based on @rsc's recommendation for best practices for tool dependencies:
// https://github.com/golang/go/issues/25922#issuecomment-413898264

// To update the version, run: go get -d honnef.co/go/tools/cmd/staticcheck@latest

package tools

import (
	_ "honnef.co/go/tools/cmd/staticcheck"
)

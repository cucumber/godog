package godog

import (
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/parser"
)

// ParseFeatures allows users to parse their feature files to in-memory objects
func ParseFeatures(filter string, paths []string) ([]*Feature, error) {
	return parser.ParseFeatures(filter, paths)
}

// Feature is an exported version of models.Feature
type Feature = models.Feature

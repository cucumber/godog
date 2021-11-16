package models

import (
	"github.com/cucumber/messages-go/v16"
)

// Feature is an internal object to group together
// the parsed gherkin document, the pickles and the
// raw content.
type Feature struct {
	*messages.GherkinDocument
	Pickles []*messages.Pickle
	Content []byte
}

// FindRule returns the rule containing astScenarioID.
func (f Feature) FindRule(astScenarioID string) *messages.Rule {
	for _, child := range f.GherkinDocument.Feature.Children {
		if ru := child.Rule; ru != nil {
			for _, ruc := range ru.Children {
				if sc := ruc.Scenario; sc != nil && sc.Id == astScenarioID {
					return ru
				}
			}
		}
	}

	return nil
}

// FindScenario returns the scenario matching astScenarioID. The scenario
// might be a direct child of Feature, or a child of a Rule within a Feature.
func (f Feature) FindScenario(astScenarioID string) *messages.Scenario {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.Scenario; sc != nil && sc.Id == astScenarioID {
			return sc
		}
		if rc := child.Rule; rc != nil {
			for _, rcc := range rc.Children {
				if sc := rcc.Scenario; sc != nil && sc.Id == astScenarioID {
					return sc
				}
			}
		}
	}

	return nil
}

// FindBackground returns the background belonging to the given astScenarioID.
// It returns the closest background in case there are multiple, e.g. if there
// is a background for both feature and rule, then the background of the rule
// is returned.
func (f Feature) FindBackground(astScenarioID string) *messages.Background {
	var bg *messages.Background

	for _, child := range f.GherkinDocument.Feature.Children {
		if rc := child.Rule; rc != nil {
			for _, rcc := range rc.Children {
				if tmp := rcc.Background; tmp != nil {
					bg = tmp
				}

				if sc := rcc.Scenario; sc != nil && sc.Id == astScenarioID {
					return bg
				}
			}
		}

		if tmp := child.Background; tmp != nil {
			bg = tmp
		}

		if sc := child.Scenario; sc != nil && sc.Id == astScenarioID {
			return bg
		}
	}

	return nil
}

// FindExample ...
func (f Feature) FindExample(exampleAstID string) (*messages.Examples, *messages.TableRow) {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.Scenario; sc != nil {
			for _, example := range sc.Examples {
				for _, row := range example.TableBody {
					if row.Id == exampleAstID {
						return example, row
					}
				}
			}
		}
	}

	return nil, nil
}

// FindStep returns the step matching astStepID. The step
// might be a child of a Scenario or a Background (which might be contained
// inside a Rule).
func (f Feature) FindStep(astStepID string) *messages.Step {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.Scenario; sc != nil {
			for _, step := range sc.Steps {
				if step.Id == astStepID {
					return step
				}
			}
		}

		if bg := child.Background; bg != nil {
			for _, step := range bg.Steps {
				if step.Id == astStepID {
					return step
				}
			}
		}

		if ru := child.Rule; ru != nil {
			for _, ruc := range ru.Children {
				if sc := ruc.Scenario; sc != nil {
					for _, step := range sc.Steps {
						if step.Id == astStepID {
							return step
						}
					}
				}

				if bg := ruc.Background; bg != nil {
					for _, step := range bg.Steps {
						if step.Id == astStepID {
							return step
						}
					}
				}
			}
		}
	}

	return nil
}

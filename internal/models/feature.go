package models

import (
	"github.com/cucumber/messages-go/v10"
)

// Feature is an internal object to group together
// the parsed gherkin document, the pickles and the
// raw content.
type Feature struct {
	*messages.GherkinDocument
	Pickles []*messages.Pickle
	Content []byte
}

// FindScenario ...
func (f Feature) FindScenario(astScenarioID string) *messages.GherkinDocument_Feature_Scenario {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.GetScenario(); sc != nil && sc.Id == astScenarioID {
			return sc
		}
	}

	return nil
}

// FindBackground ...
func (f Feature) FindBackground(astScenarioID string) *messages.GherkinDocument_Feature_Background {
	var bg *messages.GherkinDocument_Feature_Background

	for _, child := range f.GherkinDocument.Feature.Children {
		if tmp := child.GetBackground(); tmp != nil {
			bg = tmp
		}

		if sc := child.GetScenario(); sc != nil && sc.Id == astScenarioID {
			return bg
		}
	}

	return nil
}

// FindExample ...
func (f Feature) FindExample(exampleAstID string) (*messages.GherkinDocument_Feature_Scenario_Examples, *messages.GherkinDocument_Feature_TableRow) {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.GetScenario(); sc != nil {
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

// FindStep ...
func (f Feature) FindStep(astStepID string) *messages.GherkinDocument_Feature_Step {
	for _, child := range f.GherkinDocument.Feature.Children {
		if sc := child.GetScenario(); sc != nil {
			for _, step := range sc.GetSteps() {
				if step.Id == astStepID {
					return step
				}
			}
		}

		if bg := child.GetBackground(); bg != nil {
			for _, step := range bg.GetSteps() {
				if step.Id == astStepID {
					return step
				}
			}
		}
	}

	return nil
}

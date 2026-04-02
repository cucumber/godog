package formatters_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cucumber/godog"
)

func TestCucumberFormatterStopOnFirstFailure(t *testing.T) {
	featureFile := "formatter-tests/features/stop_on_first_failure.feature"

	var buf bytes.Buffer
	opts := godog.Options{
		Format:        "cucumber",
		Paths:         []string{featureFile},
		Output:        &buf,
		Strict:        true,
		StopOnFailure: true,
	}

	suite := godog.TestSuite{
		Name: "Cucumber - Ensure JSON report is produced when StopOnFailure is set",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			setupStopOnFailureSteps(sc)
		},
		Options: &opts,
	}

	if status := suite.Run(); status != 1 {
		t.Fatalf("expected status 1, but got %d", status)
	}

	var features []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &features); err != nil {
		t.Fatalf("failed to parse cucumber: %v\noutput:\n%s", err, buf.String())
	}

	if len(features) != 1 {
		t.Fatalf("expected exactly one feature, got %d", len(features))
	}

	elements := features[0]["elements"].([]interface{})
	if len(elements) != 2 {
		t.Fatalf("expected two scenarios, got %d", len(elements))
	}

	second := elements[1].(map[string]interface{})
	if _, ok := second["steps"]; ok { // stopOnFailure prevents steps from being added
		t.Fatalf("expected second scenario to have no steps, but got %v", second["steps"])
	}
}

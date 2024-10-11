package solution

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog/internal/formatters"
)

// Cucumber combiner - "knows" how to combine multiple "cucumber" reports into one

func CombineCukeOutputs(outputs [][]byte) ([]byte, error) {
	var result []formatters.CukeFeatureJSON

	for _, output := range outputs {
		var cukeFeatureJSONS []formatters.CukeFeatureJSON

		err := json.Unmarshal(output, &cukeFeatureJSONS)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal cuke feature JSON: %w", err)
		}

		result = append(result, cukeFeatureJSONS...)
	}

	aggregatedResults, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("can't marshal combined cuke feature JSON: %w", err)
	}

	return aggregatedResults, nil
}

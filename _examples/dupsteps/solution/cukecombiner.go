package solution

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog/internal/formatters"
)

// CombineCukeReports "knows" how to combine multiple "cucumber" reports into one
func CombineCukeReports(cukeReportOutputs [][]byte) ([]byte, error) {
	var allCukeFeatureJSONs []formatters.CukeFeatureJSON

	for _, output := range cukeReportOutputs {
		var cukeFeatureJSONS []formatters.CukeFeatureJSON

		err := json.Unmarshal(output, &cukeFeatureJSONS)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal cuke feature JSON: %w", err)
		}

		allCukeFeatureJSONs = append(allCukeFeatureJSONs, cukeFeatureJSONS...)
	}

	combinedCukeReport, err := json.MarshalIndent(allCukeFeatureJSONs, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("can't marshal combined cuke feature JSON: %w", err)
	}

	return combinedCukeReport, nil
}

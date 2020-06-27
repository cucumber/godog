package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/models"
)

type stepResultStatusTestCase struct {
	st  models.StepResultStatus
	str string
	clr colors.ColorFunc
}

var stepResultStatusTestCases = []stepResultStatusTestCase{
	{st: models.Passed, str: "passed", clr: colors.Green},
	{st: models.Failed, str: "failed", clr: colors.Red},
	{st: models.Skipped, str: "skipped", clr: colors.Cyan},
	{st: models.Undefined, str: "undefined", clr: colors.Yellow},
	{st: models.Pending, str: "pending", clr: colors.Yellow},
	{st: -1, str: "unknown", clr: colors.Yellow},
}

func Test_StepResultStatus(t *testing.T) {
	for _, tc := range stepResultStatusTestCases {
		t.Run(tc.str, func(t *testing.T) {
			assert.Equal(t, tc.str, tc.st.String())
			assert.Equal(t, tc.clr(tc.str), tc.st.Color()(tc.str))
		})
	}
}

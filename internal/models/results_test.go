package models_test

import (
	"fmt"
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
	{st: models.Ambiguous, str: "ambiguous", clr: colors.Yellow},
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

func Test_NewStepResuklt(t *testing.T) {
	status := models.StepResultStatus(123)
	pickleID := "pickleId"
	pickleStepID := "pickleStepID"
	match := &models.StepDefinition{}
	attachments := make([]models.PickleAttachment, 0)
	err := fmt.Errorf("intentional")

	results := models.NewStepResult(status, pickleID, pickleStepID, match, attachments, err)

	assert.Equal(t, status, results.Status)
	assert.Equal(t, pickleID, results.PickleID)
	assert.Equal(t, pickleStepID, results.PickleStepID)
	assert.Equal(t, match, results.Def)
	assert.Equal(t, attachments, results.Attachments)
	assert.Equal(t, err, results.Err)
}

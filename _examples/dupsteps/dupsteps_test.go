package dupsteps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog"
)

func TestDuplicateSteps(t *testing.T) {

	featureContents := `
Feature:	check this out

	Scenario: Flat Tire
		Given I ran over a nail and got a flat tire
		Then I fixed it
		Then I can continue on my way

	Scenario: Clogged Drain
		Given I accidentally poured concrete down my drain and clogged the sewer line
		Then I fixed it
		Then I can once again use my sink
`

	suite := godog.TestSuite{
		Name: t.Name(),
		ScenarioInitializer: func(context *godog.ScenarioContext) {
			// NOTE: loading implementations of steps for different scenarios here
			(&cloggedDrain{}).addCloggedDrainSteps(context)
			(&flatTire{}).addFlatTireSteps(context)
		},
		Options: &godog.Options{
			Format: "pretty",
			Strict: true,
			FeatureContents: []godog.Feature{
				{
					Name:     fmt.Sprintf("%s contents", t.Name()),
					Contents: []byte(featureContents),
				},
			},
		},
	}

	rc := suite.Run()
	assert.Zero(t, rc)
}

// Implementation of the steps for the "Clogged Drain" scenario

type cloggedDrain struct {
	drainIsClogged bool
}

func (cd *cloggedDrain) addCloggedDrainSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I accidentally poured concrete down my drain and clogged the sewer line$`, cd.clogSewerLine)
	ctx.Step(`^I fixed it$`, cd.iFixedIt)
	ctx.Step(`^I can once again use my sink$`, cd.useTheSink)
}

func (cd *cloggedDrain) clogSewerLine() error {
	cd.drainIsClogged = true

	return nil
}

func (cd *cloggedDrain) iFixedIt() error {
	cd.drainIsClogged = false

	return nil
}

func (cd *cloggedDrain) useTheSink() error {
	if cd.drainIsClogged {
		return fmt.Errorf("drain is clogged")
	}

	return nil
}

// Implementation of the steps for the "Flat Tire" scenario

type flatTire struct {
	tireIsFlat bool
}

func (ft *flatTire) addFlatTireSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I ran over a nail and got a flat tire$`, ft.gotFlatTire)
	ctx.Step(`^I fixed it$`, ft.iFixedIt)
	ctx.Step(`^I can continue on my way$`, ft.continueOnMyWay)
}

func (ft *flatTire) gotFlatTire() error {
	ft.tireIsFlat = true

	return nil
}

func (ft *flatTire) iFixedIt() error {
	ft.tireIsFlat = false

	return nil
}

func (ft *flatTire) continueOnMyWay() error {
	if ft.tireIsFlat {
		return fmt.Errorf("tire was not fixed")
	}

	return nil
}

package attachments_test

// This example shows how to set up test suite runner with Go subtests and godog command line parameters.
// Sample commands:
// * run all scenarios from default directory (features): go test -test.run "^TestFeatures/"
// * run all scenarios and list subtest names: go test -test.v -test.run "^TestFeatures/"
// * run all scenarios from one feature file: go test -test.v -godog.paths features/nodogs.feature -test.run "^TestFeatures/"
// * run all scenarios from multiple feature files: go test -test.v -godog.paths features/nodogs.feature,features/godogs.feature -test.run "^TestFeatures/"
// * run single scenario as a subtest: go test -test.v -test.run "^TestFeatures/Eat_5_out_of_12$"
// * show usage help: go test -godog.help
// * show usage help if there were other test files in directory: go test -godog.help godogs_test.go
// * run scenarios with multiple formatters: go test -test.v -godog.format cucumber:cuc.json,pretty -test.run "^TestFeatures/"

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "cucumber", // cucumber json format
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                "attachments",
		Options:             &o,
		ScenarioInitializer: InitializeScenario,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {

	ctx.Step(`^I have attached two documents in sequence$`, func(ctx context.Context) (context.Context, error) {
		// the attached bytes will be base64 encoded by the framework and placed in the embeddings section of the cuke report
		ctx = godog.Attach(ctx,
			godog.Attachment{Body: []byte("TheData1"), FileName: "Data Attachment", MediaType: "text/plain"},
		)
		ctx = godog.Attach(ctx,
			godog.Attachment{Body: []byte("{ \"a\" : 1 }"), FileName: "Json Attachment", MediaType: "application/json"},
		)

		return ctx, nil
	})
	ctx.Step(`^I have attached two documents at once$`, func(ctx context.Context) (context.Context, error) {
		ctx = godog.Attach(ctx,
			godog.Attachment{Body: []byte("TheData1"), FileName: "Data Attachment 1", MediaType: "text/plain"},
			godog.Attachment{Body: []byte("TheData2"), FileName: "Data Attachment 2", MediaType: "text/plain"},
		)

		return ctx, nil
	})
}

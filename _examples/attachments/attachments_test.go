package attachments_test

// This example shows how to attach data to the cucumber reports
// Run the sample with : go test -v attachments_test.go
// Then review the "embeddings" within the JSON emitted on the console.

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

	ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		ctx = godog.Attach(ctx,
			godog.Attachment{Body: []byte("BeforeStepAttachment"), FileName: "Data Attachment", MediaType: "text/plain"},
		)
		return ctx, nil
	})
	ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		ctx = godog.Attach(ctx,
			godog.Attachment{Body: []byte("AfterStepAttachment"), FileName: "Data Attachment", MediaType: "text/plain"},
		)
		return ctx, nil
	})

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

- run.go
  - TestSuite and it's Run() method are defined
  - creates a runner instance and calls runWithOptions, which is where:
    - the formatters are set up
    - parser.ParseFeatures is called
      - Defined
    - runner.concurrent is called
      - this is where TestRunStarted happens as execution starts


OK, so questions overall about what's called where?


https://github.com/cucumber/cucumber-jvm/blob/bb8f6bb1e007870b79d8d43d828be6490b6d3189/core/src/main/java/io/cucumber/core/runner/TestCase.java
    void run(EventBus bus) {
        ExecutionMode nextExecutionMode = this.executionMode;
        emitTestCaseMessage(bus);

        Instant start = bus.getInstant();
        UUID executionId = bus.generateId();
        emitTestCaseStarted(bus, start, executionId);


Meta:
{
	"meta": {
		"ci": {
			"buildNumber": "154666429",
			"git": {
				"remote": "https://github.com/cucumber-ltd/shouty.rb.git",
				"revision": "99684bcacf01d95875834d87903dcb072306c9ad"
			},
			"name": "GitHub Actions",
			"url": "https://github.com/cucumber-ltd/shouty.rb/actions/runs/154666429"
		},
		"cpu": {
			"name": "x64"
		},
		"implementation": {
			"name": "fake-cucumber",
			"version": "16.0.0"
		},
		"os": {
			"name": "linux",
			"version": "5.10.102.1-microsoft-standard-WSL2"
		},
		"protocolVersion": "19.1.2",
		"runtime": {
			"name": "node.js",
			"version": "16.4.0"
		}
	}
}

Source:

{
	"source": {
		"data": "Feature: minimal\n  \n  Cucumber doesn't execute this markdown, but @cucumber/react renders it\n  \n  * This is\n  * a bullet\n  * list\n  \n  Scenario: cukes\n    Given I have 42 cukes in my belly\n",
		"mediaType": "text/x.cucumber.gherkin+plain",
		"uri": "samples/minimal/minimal.feature"
	}
}

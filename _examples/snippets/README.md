# Snippets

Snippets are generated when undefined steps are created. 

Currently, we support the following snippet functions:

| name       | description                                                       |
|------------|-------------------------------------------------------------------|
| step_func  | Generates steps with the Step keyword and function bodies         |
| gwt_func   | Generates steps with Given/When/Then keywords and function bodies |

Examples show the difference between each snippet generator.

## Examples

The first example uses the *step_func* snippet function to generate the snippet. This works by not providing a 
snippet func or explicitly providing *step_func*. This example does not provide the snippet function explicitly. 

Run the following command to view the output of the step_func example.

```shell
go test -test.v ./step_func
```

The output should be:

```
=== RUN   TestFeatures
Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs
=== RUN   TestFeatures/Eat_12_out_of_12
=== RUN   TestFeatures/Eat_5_out_of_12

  Scenario: Eat 12 out of 12            # features/godogs.feature:11
    Given there are 12 godogs

  Scenario: Eat 5 out of 12          # features/godogs.feature:6
    Given there are 12 godogs
    When I eat 12
    Then there should be none remaining
    When I eat 5
    Then there should be 7 remaining

2 scenarios (2 undefined)
6 steps (6 undefined)
271.125µs

You can implement step definitions for undefined steps with these snippets:

func iEat(arg1 int) error {
        return godog.ErrPending
}

func thereAreGodogs(arg1 int) error {
        return godog.ErrPending
}

func thereShouldBeNoneRemaining() error {
        return godog.ErrPending
}

func thereShouldBeRemaining(arg1 int) error {
        return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
        ctx.When(`^I eat (\d+)$`, iEat)
        ctx.Given(`^there are (\d+) godogs$`, thereAreGodogs)
        ctx.Then(`^there should be none remaining$`, thereShouldBeNoneRemaining)
        ctx.Then(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
}

--- PASS: TestFeatures (0.00s)
    --- PASS: TestFeatures/Eat_12_out_of_12 (0.00s)
    --- PASS: TestFeatures/Eat_5_out_of_12 (0.00s)
PASS
ok      github.com/cucumber/godog/_examples/snippets/gwt_func   (cached)
```

The second example uses the *gwt_func*  snippet function to generate the snippet. This works by providing a
snippet func or explicitly.

```go
var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	SnippetFunc: "gwt_func",
	Concurrency: 4,
}
```

Run the following command to view the output of the gwt_func example.

```shell
go test -test.v ./gwt_func
```

The output should be:

```
=== RUN   TestFeatures
Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs
=== RUN   TestFeatures/Eat_12_out_of_12
=== RUN   TestFeatures/Eat_5_out_of_12

  Scenario: Eat 12 out of 12            # features/godogs.feature:11
    Given there are 12 godogs

  Scenario: Eat 5 out of 12          # features/godogs.feature:6
    Given there are 12 godogs
    When I eat 12
    Then there should be none remaining
    When I eat 5
    Then there should be 7 remaining

2 scenarios (2 undefined)
6 steps (6 undefined)
271.125µs

You can implement step definitions for undefined steps with these snippets:

func iEat(arg1 int) error {
        return godog.ErrPending
}

func thereAreGodogs(arg1 int) error {
        return godog.ErrPending
}

func thereShouldBeNoneRemaining() error {
        return godog.ErrPending
}

func thereShouldBeRemaining(arg1 int) error {
        return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
        ctx.When(`^I eat (\d+)$`, iEat)
        ctx.Given(`^there are (\d+) godogs$`, thereAreGodogs)
        ctx.Then(`^there should be none remaining$`, thereShouldBeNoneRemaining)
        ctx.Then(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
}

--- PASS: TestFeatures (0.00s)
```
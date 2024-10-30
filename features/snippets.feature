@john3
Feature: undefined step snippets
  In order to implement step definitions faster
  As a test suite user
  I need to be able to get undefined step snippets

  Scenario: should generate snippets
    Given a feature "undefined.feature" file:
      """
      Feature: undefined steps
        Scenario: has undefined
          When some "undefined" step
          And another undefined step
          And a table:
            | col1 | val1 |
          And a docstring:
            \"\"\"
            Hello World
            \"\"\"
      """
    When I run feature suite
    Then the following steps should be undefined:
      """
      a docstring:
      a table:
      another undefined step
      some "undefined" step
      """
    And the undefined step snippets should be:
      """
        func aDocstring(arg1 *godog.DocString) error {
        	return godog.ErrPending
        }

        func aTable(arg1 *godog.Table) error {
        	return godog.ErrPending
        }

        func anotherUndefinedStep() error {
        	return godog.ErrPending
        }

        func someStep(arg1 string) error {
        	return godog.ErrPending
        }

        func InitializeScenario(ctx *godog.ScenarioContext) {
        	ctx.Step(`^a docstring:$`, aDocstring)
        	ctx.Step(`^a table:$`, aTable)
        	ctx.Step(`^another undefined step$`, anotherUndefinedStep)
        	ctx.Step(`^some "([^"]*)" step$`, someStep)
        }
      """

  Scenario: should generate snippets with more arguments
    Given a feature "undefined.feature" file:
      """
      Feature: undefined steps

        Scenario: get version number from api
          When I send "GET" request to "/version" with:
            | col1 | val1 |
            | col2 | val2 |
          Then the response code should be 200 and header "X-Powered-By" should be "godog"
          And the response body should be:
          \"\"\"
          Hello World
          \"\"\"
      """
    When I run feature suite
    Then the undefined step snippets should be:
      """
      func iSendRequestToWith(arg1, arg2 string, arg3 *godog.Table) error {
              return godog.ErrPending
      }

      func theResponseBodyShouldBe(arg1 *godog.DocString) error {
              return godog.ErrPending
      }

      func theResponseCodeShouldBeAndHeaderShouldBe(arg1 int, arg2, arg3 string) error {
              return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
              ctx.Step(`^I send "([^"]*)" request to "([^"]*)" with:$`, iSendRequestToWith)
              ctx.Step(`^the response body should be:$`, theResponseBodyShouldBe)
              ctx.Step(`^the response code should be (\d+) and header "([^"]*)" should be "([^"]*)"$`, theResponseCodeShouldBeAndHeaderShouldBe)
      }
      """

  Scenario: should handle escaped symbols
    Given a feature "undefined.feature" file:
      """
      Feature: undefined steps

        Scenario: get version number from api
          When I pull from github.com
          Then the project should be there
      """
    When I run feature suite
    Then the following steps should be undefined:
      """
      I pull from github.com
      the project should be there
      """
    And the undefined step snippets should be:
      """
      func iPullFromGithubcom() error {
              return godog.ErrPending
      }

      func theProjectShouldBeThere() error {
              return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
              ctx.Step(`^I pull from github\.com$`, iPullFromGithubcom)
              ctx.Step(`^the project should be there$`, theProjectShouldBeThere)
      }
      """

  Scenario: should handle string argument followed by comma
    Given a feature "undefined.feature" file:
      """
      Feature: undefined

        Scenario: add item to basket
          Given there is a "Sith Lord Lightsaber", which costs £5
          When I add the "Sith Lord Lightsaber" to the basket
      """
    When I run feature suite
    And the undefined step snippets should be:
      """
      func iAddTheToTheBasket(arg1 string) error {
              return godog.ErrPending
      }

      func thereIsAWhichCosts(arg1 string, arg2 int) error {
              return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
              ctx.Step(`^I add the "([^"]*)" to the basket$`, iAddTheToTheBasket)
              ctx.Step(`^there is a "([^"]*)", which costs £(\d+)$`, thereIsAWhichCosts)
      }
      """

  Scenario: should handle arguments in the beggining or end of the step
    Given a feature "undefined.feature" file:
      """
      Feature: undefined

        Scenario: add item to basket
          Given "Sith Lord Lightsaber", which costs £5
          And 12 godogs
      """
    When I run feature suite
    And the undefined step snippets should be:
      """
      func godogs(arg1 int) error {
              return godog.ErrPending
      }

      func whichCosts(arg1 string, arg2 int) error {
              return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
              ctx.Step(`^(\d+) godogs$`, godogs)
              ctx.Step(`^"([^"]*)", which costs £(\d+)$`, whichCosts)
      }
      """

Feature: undefined step snippets
  In order to implement step definitions faster
  As a test suite user
  I need to be able to get undefined step snippets

  Scenario: should generate snippets
    Given a feature "undefined.feature" file:
      """
      Feature: undefined steps

        Scenario: get version number from api
          When I send "GET" request to "/version"
          Then the response code should be 200
      """
    When I run feature suite
    Then the following steps should be undefined:
      """
      I send "GET" request to "/version"
      the response code should be 200
      """
    And the undefined step snippets should be:
      """
      func iSendRequestTo(arg1, arg2 string) error {
              return godog.ErrPending
      }

      func theResponseCodeShouldBe(arg1 int) error {
              return godog.ErrPending
      }

      func InitializeScenario(ctx *godog.ScenarioContext) {
              ctx.Step(`^I send "([^"]*)" request to "([^"]*)"$`, iSendRequestTo)
              ctx.Step(`^the response code should be (\d+)$`, theResponseCodeShouldBe)
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

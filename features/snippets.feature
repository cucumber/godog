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
      func iSendrequestTo(arg1, arg2 string) error {
              return godog.ErrPending
      }

      func theResponseCodeShouldBe(arg1 int) error {
              return godog.ErrPending
      }

      func featureContext(s *godog.Suite) {
              s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, iSendrequestTo)
              s.Step(`^the response code should be (\d+)$`, theResponseCodeShouldBe)
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
      """
    When I run feature suite
    Then the undefined step snippets should be:
      """
      func iSendrequestTowith(arg1, arg2 string, arg3 *gherkin.DataTable) error {
              return godog.ErrPending
      }

      func theResponseCodeShouldBeAndHeadershouldBe(arg1 int, arg2, arg3 string) error {
              return godog.ErrPending
      }

      func featureContext(s *godog.Suite) {
              s.Step(`^I send "([^"]*)" request to "([^"]*)" with:$`, iSendrequestTowith)
              s.Step(`^the response code should be (\d+) and header "([^"]*)" should be "([^"]*)"$`, theResponseCodeShouldBeAndHeadershouldBe)
      }
      """


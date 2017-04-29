Feature: load features
  In order to run features
  As a test suite
  I need to be able to load features

  Scenario: load features within path
    Given a feature path "features"
    When I parse features
    Then I should have 10 feature files:
      """
      features/background.feature
      features/events.feature
      features/formatter/cucumber.feature
      features/formatter/events.feature
      features/lang.feature
      features/load.feature
      features/multistep.feature
      features/outline.feature
      features/run.feature
      features/snippets.feature
      """

  Scenario: load a specific feature file
    Given a feature path "features/load.feature"
    When I parse features
    Then I should have 1 feature file:
      """
      features/load.feature
      """

  Scenario Outline: loaded feature should have a number of scenarios
    Given a feature path "<feature>"
    When I parse features
    Then I should have <number> scenario registered

    Examples:
      | feature                 | number |
      | features/load.feature:3 | 0      |
      | features/load.feature:6 | 1      |
      | features/load.feature   | 4      |

  Scenario: load a number of feature files
    Given a feature path "features/load.feature"
    And a feature path "features/events.feature"
    When I parse features
    Then I should have 2 feature files:
      """
      features/events.feature
      features/load.feature
      """

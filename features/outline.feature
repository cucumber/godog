Feature: run outline
  In order to test application behavior
  As a test suite
  I need to be able to run outline scenarios

  Scenario: should run a normal outline
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given passing step

        Scenario Outline: parse a scenario
          Given a feature path "<path>"
          When I parse features
          Then I should have <num> scenario registered

          Examples:
            | path                    | num |
            | features/load.feature:6 | 1   |
            | features/load.feature:3 | 0   |
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      I parse features
      I parse features
      passing step
      passing step
      I should have 1 scenario registered
      I should have 0 scenario registered
      a feature path "features/load.feature:6"
      a feature path "features/load.feature:3"
      """

  Scenario: should continue through examples on failure
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given passing step

        Scenario Outline: parse a scenario
          Given a feature path "<path>"
          When I parse features
          Then I should have <num> scenario registered

          Examples:
            | path                    | num |
            | features/load.feature:6 | 5   |
            | features/load.feature:3 | 0   |
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      I parse features
      I parse features
      I should have 0 scenario registered
      a feature path "features/load.feature:3"
      a feature path "features/load.feature:6"
      passing step
      passing step
      """
    And the following steps should be failed:
      """
      I should have 5 scenario registered
      """

  Scenario: should skip examples on background failure
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given a failing step

        Scenario Outline: parse a scenario
          Given a feature path "<path>"
          When I parse features
          Then I should have <num> scenario registered

          Examples:
            | path                    | num |
            | features/load.feature:6 | 1   |
            | features/load.feature:3 | 0   |
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be skipped:
      """
      I parse features
      I parse features
      a feature path "features/load.feature:6"
      a feature path "features/load.feature:3"
      I should have 0 scenario registered
      I should have 1 scenario registered
      """
    And the following steps should be failed:
      """
      a failing step
      a failing step
      """

  Scenario: should translate step table body
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given I'm listening to suite events

        Scenario Outline: run with events
          Given a feature path "<path>"
          When I run feature suite
          Then these events had to be fired for a number of times:
            | BeforeScenario | <scen> |
            | BeforeStep     | <step> |

          Examples:
            | path                    | scen | step |
            | features/load.feature:6 | 1    | 3    |
            | features/load.feature   | 6    | 19   |
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      I run feature suite
      I run feature suite
      I'm listening to suite events
      I'm listening to suite events
      a feature path "features/load.feature"
      a feature path "features/load.feature:6"
      these events had to be fired for a number of times:
      these events had to be fired for a number of times:
      """

  Scenario Outline: should translate step doc string argument
    Given a feature "normal.feature" file:
      """
      Feature: scenario events

        Background:
          Given I'm listening to suite events

        Scenario: run with events
          Given a feature path "<path>"
          When I run feature suite
          Then these events had to be fired for a number of times:
            | BeforeScenario | <scen> |
      """
    When I run feature suite
    Then the suite should have passed

    Examples:
      | path                    | scen |
      | features/load.feature:6 | 1    |
      | features/load.feature   | 6    |


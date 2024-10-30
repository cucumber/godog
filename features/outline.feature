Feature: run outline
  In order to test application behavior
  As a test suite
  I need to be able to run outline scenarios

  Scenario: should continue through other examples even if some examples fail
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given second passing step

        Scenario Outline: continue execution despite failing examples
          Then <status> step

          Examples:
            | status  |
            | passing |
            | failing |
            | other passing |
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
     second passing step
     second passing step
     second passing step
     passing step
     other passing step
      """
    And the following steps should be failed:
      """
      failing step
      """

  Scenario: should skip scenario examples if background fails
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given a failing step

        Scenario Outline: parse a scenario
          Given <status> step

          Examples:
            | status  |
            | passing |
            | other passing |
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be skipped:
      """
      passing step
      other passing step
      """
    And the following steps should be failed:
      """
      a failing step
      a failing step
      """

  Scenario: table should be injected with example values
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Scenario Outline: run with events
          Given <status> step
          Then value2 is twice value1:
            | Value1 | <value1> |
            | Value2 | <value2> |

          Examples:
            | status   | value1 | value2 |
            | passing  | 2      | 4    |
            | passing  | 11     | 22   |
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      value2 is twice value1:
      passing step
      value2 is twice value1:
      """

  @john
  Scenario Outline: docstring should be injected with example values
    Given a feature "normal.feature" file:
      """
      Feature: scenario events

        Scenario: run <status> params
          Given <status> step
     """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      <status> step
      """

    Examples:
      | status        |
      | passing       |
      | other passing |


    @john
  Scenario: scenario title may be injected with example values
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario Outline: scenario with <param> in title
        When <param> step

        Examples:
          | param     |
          | passing   |
          | failing   |
      """
    When I run feature suite

    Then the suite should have failed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [scenario with passing in title]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    AfterScenario [scenario with passing in title]
    BeforeScenario [scenario with failing in title]
    BeforeStep [failing step]
    AfterStep [failing step] [failed] [intentional failure]
    AfterScenario [scenario with failing in title] [intentional failure]
    AfterSuite
    """

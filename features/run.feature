
Feature: sequencing of steps and hooks

  Scenario: passing scenario
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And passing step that fires an event
      """
    When I run feature suite

    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [passing step that fires an event]
    Step [passing step that fires an event]
    AfterStep [passing step that fires an event] [passed]
    AfterScenario [passing scenario]
    AfterSuite
    """

  Scenario: should skip steps after undefined
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And an undefined step
        And another undefined step
        And second passing step
      """
    When I run feature suite

    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [an undefined step]
    AfterStep [an undefined step] [undefined] [step is undefined]
    BeforeStep [another undefined step]
    AfterStep [another undefined step] [undefined] [step is undefined]
    BeforeStep [second passing step]
    AfterStep [second passing step] [skipped]
    AfterScenario [passing scenario]
    AfterSuite
    """

    # FIXME JOHN STEP ORDERING ISSUE
  Scenario: should fail if undefined steps in Strict mode
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And an undefined step
        And another undefined step
        And second passing step
      """
    When I run feature suite in Strict mode

    Then the suite should have failed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [an undefined step]
    AfterStep [an undefined step] [undefined] [step is undefined]
    AfterScenario [passing scenario] [step is undefined]
    BeforeStep [another undefined step]
    AfterStep [another undefined step] [undefined] [step is undefined]
    BeforeStep [second passing step]
    AfterStep [second passing step] [skipped]
    AfterSuite
    """

  Scenario: should skip steps after ambiguous
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And an ambiguous step
        And another ambiguous step
        And second passing step
      """
    When I run feature suite

    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [an ambiguous step]
    AfterStep [an ambiguous step] [passed]
    BeforeStep [another ambiguous step]
    AfterStep [another ambiguous step] [passed]
    BeforeStep [second passing step]
    AfterStep [second passing step] [passed]
    AfterScenario [passing scenario]
    AfterSuite
    """

  Scenario: should fail steps after ambiguous steps in Strict mode
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And an ambiguous step
        And another ambiguous step
        And second passing step
      """
    When I run feature suite in Strict mode

    Then the suite should have failed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [an ambiguous step]
    AfterStep [an ambiguous step] [ambiguous] [ambiguous step definition, step text: an ambiguous step
        matches:
            ^.*ambiguous step$
            ^..*ambiguous step$]
    AfterScenario [passing scenario] [ambiguous step definition, step text: an ambiguous step
        matches:
            ^.*ambiguous step$
            ^..*ambiguous step$]
    BeforeStep [another ambiguous step]
    AfterStep [another ambiguous step] [ambiguous] [ambiguous step definition, step text: another ambiguous step
        matches:
            ^.*ambiguous step$
            ^..*ambiguous step$]
    BeforeStep [second passing step]
    AfterStep [second passing step] [skipped]
    AfterSuite
    """

  Scenario: should skip existing steps and detect undefined steps after pending
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And a pending step
        And another undefined step
        And second passing step
      """
    When I run feature suite

    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [a pending step]
    AfterStep [a pending step] [pending] [step implementation is pending]
    BeforeStep [another undefined step]
    AfterStep [another undefined step] [undefined] [step is undefined]
    BeforeStep [second passing step]
    AfterStep [second passing step] [skipped]
    AfterScenario [passing scenario]
    AfterSuite
    """


    # FIXME JOHN THIS IS THE BROKEN ORDERING
  Scenario: scenario hook runs after all passing and failing tests
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And failing step
        And failing step
        And other passing step
        And an undefined step
        And a pending step
      """
    When I run feature suite

    Then the suite should have failed
    And the following events should be fired:
      """
      BeforeSuite
      BeforeScenario [passing scenario]
      BeforeStep [passing step]
      AfterStep [passing step] [passed]
      BeforeStep [failing step]
      AfterStep [failing step] [failed] [intentional failure]
      AfterScenario [passing scenario] [intentional failure]
      BeforeStep [failing step]
      AfterStep [failing step] [skipped]
      BeforeStep [other passing step]
      AfterStep [other passing step] [skipped]
      BeforeStep [an undefined step]
      AfterStep [an undefined step] [undefined] [step is undefined]
      BeforeStep [a pending step]
      AfterStep [a pending step] [skipped]
      AfterSuite
      """

  Scenario: no errors event check
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
        And passing step that fires an event
      """
    Given a feature "other.feature" file:
      """
    Feature: the other feature
      Scenario: other passing scenario
        When other passing step
        And other passing step that fires an event
      """
    When I run feature suite

    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    BeforeScenario [passing scenario]
    BeforeStep [passing step]
    AfterStep [passing step] [passed]
    BeforeStep [passing step that fires an event]
    Step [passing step that fires an event]
    AfterStep [passing step that fires an event] [passed]
    AfterScenario [passing scenario]
    BeforeScenario [other passing scenario]
    BeforeStep [other passing step]
    AfterStep [other passing step] [passed]
    BeforeStep [other passing step that fires an event]
    Step [other passing step that fires an event]
    AfterStep [other passing step that fires an event] [passed]
    AfterScenario [other passing scenario]
    AfterSuite
    """

  Scenario: should not trigger events on empty feature
    Given a feature "normal.feature" file:
      """
      Feature: empty

        Scenario: one

        Scenario: two
      """
    When I run feature suite
    Then the suite should have passed
    And the following events should be fired:
    """
    BeforeSuite
    AfterSuite
    """

  Scenario: should not trigger events on empty scenarios
    Given a feature "normal.feature" file:
  """
      Feature: half empty

        Scenario: one

        Scenario: two
          And passing step that fires an event
          And another passing step that fires an event
          And failing step

        Scenario Outline: three
          Then passing step

          Examples:
            | a |
            | 1 |
      """
    When I run feature suite
    Then the suite should have failed
    And the following events should be fired:
      """
      BeforeSuite
      BeforeScenario [two]
      BeforeStep [passing step that fires an event]
      Step [passing step that fires an event]
      AfterStep [passing step that fires an event] [passed]
      BeforeStep [another passing step that fires an event]
      Step [another passing step that fires an event]
      AfterStep [another passing step that fires an event] [passed]
      BeforeStep [failing step]
      AfterStep [failing step] [failed] [intentional failure]
      AfterScenario [two] [intentional failure]
      BeforeScenario [three]
      BeforeStep [passing step]
      AfterStep [passing step] [passed]
      AfterScenario [three]
      AfterSuite
      """

    And the suite should have failed


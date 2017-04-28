Feature: event stream formatter
  In order to have universal cucumber formatter
  As a test suite
  I need to be able to support event stream formatter

  Scenario: should fire only suite events without any scenario
    Given a feature path "features/load.feature:4"
    When I run feature suite with formatter "events"
    Then the following events should be fired:
      """
        TestRunStarted
        TestSource
        TestRunFinished
      """

  Scenario: should process simple scenario
    Given a feature path "features/load.feature:23"
    When I run feature suite with formatter "events"
    Then the following events should be fired:
      """
        TestRunStarted
        TestSource
        TestCaseStarted
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        TestCaseFinished
        TestRunFinished
      """

  Scenario: should process outline scenario
    Given a feature path "features/load.feature:31"
    When I run feature suite with formatter "events"
    Then the following events should be fired:
      """
        TestRunStarted
        TestSource
        TestCaseStarted
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        TestCaseFinished
        TestCaseStarted
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        TestCaseFinished
        TestCaseStarted
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        StepDefinitionFound
        TestStepStarted
        TestStepFinished
        TestCaseFinished
        TestRunFinished
      """

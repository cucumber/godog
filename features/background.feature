Feature: run background
  In order to test application behavior
  As a test suite
  I need to be able to run background correctly

  Scenario: should run background steps
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given a background step is defined

        Scenario: parse a scenario
          Then step 'a background step is defined' should have been executed
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      a background step is defined
      step 'a background step is defined' should have been executed
      """

  Scenario: should skip all subsequent steps on failure
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given a failing step
          Then this step should not be called

        Scenario: parse a scenario
          And this other step should not be called
          And this last step should not be called
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      this step should not be called
      this other step should not be called
      this last step should not be called
      """

  Scenario: should continue undefined steps
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given an undefined step

        Scenario: parse a scenario
          When some other undefined step
          Then this step should not be called
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      an undefined step
      some other undefined step
      """
    And the following steps should be skipped:
      """
      this step should not be called
      """

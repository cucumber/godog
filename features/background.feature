Feature: run background
  In order to test application behavior
  As a test suite
  I need to be able to run background correctly

  Scenario: should run background steps
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given a feature path "features/load.feature:6"

        Scenario: parse a scenario
          When I parse features
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      a feature path "features/load.feature:6"
      I parse features
      I should have 1 scenario registered
      """

  Scenario: should skip all consequent steps on failure
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given a failing step
          And a feature path "features/load.feature:6"

        Scenario: parse a scenario
          When I parse features
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      a feature path "features/load.feature:6"
      I parse features
      I should have 1 scenario registered
      """

  Scenario: should continue undefined steps
    Given a feature "normal.feature" file:
      """
      Feature: with background

        Background:
          Given an undefined step

        Scenario: parse a scenario
          When I do undefined action
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      an undefined step
      I do undefined action
      """
    And the following steps should be skipped:
      """
      I should have 1 scenario registered
      """

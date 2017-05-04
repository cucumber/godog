Feature: run features
  In order to test application behavior
  As a test suite
  I need to be able to run features

  Scenario: should run a normal feature
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

        Scenario: parse a scenario
          Given a feature path "features/load.feature:6"
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

  Scenario: should skip steps after failure
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: parse a scenario
          Given a failing step
          When I parse features
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have failed
    And the following step should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      I parse features
      I should have 1 scenario registered
      """

  Scenario: should skip all scenarios if background fails
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Background:
          Given a failing step

        Scenario: parse a scenario
          Given a feature path "features/load.feature:6"
          When I parse features
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have failed
    And the following step should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      a feature path "features/load.feature:6"
      I parse features
      I should have 1 scenario registered
      """

  Scenario: should skip steps after undefined
    Given a feature "undefined.feature" file:
      """
      Feature: undefined feature

        Scenario: parse a scenario
          Given a feature path "features/load.feature:6"
          When undefined action
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following step should be passed:
      """
      a feature path "features/load.feature:6"
      """
    And the following step should be undefined:
      """
      undefined action
      """
    And the following step should be skipped:
      """
      I should have 1 scenario registered
      """

  Scenario: should match undefined steps in a row
    Given a feature "undefined.feature" file:
      """
      Feature: undefined feature

        Scenario: parse a scenario
          Given undefined step
          When undefined action
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      undefined step
      undefined action
      """
    And the following step should be skipped:
      """
      I should have 1 scenario registered
      """

  Scenario: should skip steps on pending
    Given a feature "pending.feature" file:
      """
      Feature: pending feature

        Scenario: parse a scenario
          Given undefined step
          When pending step
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following step should be undefined:
      """
      undefined step
      """
    And the following step should be skipped:
      """
      pending step
      I should have 1 scenario registered
      """

  Scenario: should handle pending step
    Given a feature "pending.feature" file:
      """
      Feature: pending feature

        Scenario: parse a scenario
          Given a feature path "features/load.feature:6"
          When pending step
          Then I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following step should be passed:
      """
      a feature path "features/load.feature:6"
      """
    And the following step should be pending:
      """
      pending step
      """
    And the following step should be skipped:
      """
      I should have 1 scenario registered
      """

  Scenario: should mark undefined steps after pending
    Given a feature "pending.feature" file:
      """
      Feature: pending feature

        Scenario: parse a scenario
          Given pending step
          When undefined
          Then undefined 2
          And I should have 1 scenario registered
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      undefined
      undefined 2
      """
    And the following step should be pending:
      """
      pending step
      """
    And the following step should be skipped:
      """
      I should have 1 scenario registered
      """

  Scenario: should fail suite if undefined steps follow after the failure
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: parse a scenario
          Given a failing step
          When an undefined step
          Then another undefined step
      """
    When I run feature suite
    Then the following step should be failed:
      """
      a failing step
      """
    And the following steps should be undefined:
      """
      an undefined step
      another undefined step
      """
    And the suite should have failed

  Scenario: should fail suite and skip pending step after failed step
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: parse a scenario
          Given a failing step
          When pending step
          Then another undefined step
      """
    When I run feature suite
    Then the following step should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      pending step
      """
    And the following steps should be undefined:
      """
      another undefined step
      """
    And the suite should have failed

  Scenario: should fail suite and skip next step after failed step
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: parse a scenario
          Given a failing step
          When a failing step
          Then another undefined step
      """
    When I run feature suite
    Then the following step should be failed:
      """
      a failing step
      """
    And the following steps should be skipped:
      """
      a failing step
      """
    And the following steps should be undefined:
      """
      another undefined step
      """
    And the suite should have failed

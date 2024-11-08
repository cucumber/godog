Feature: run features with nested steps
  In order to test multisteps
  As a test suite
  I need to be able to execute multisteps

  Scenario: should run passing multistep successfully
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

        Scenario: run passing multistep
          Given passing step
          Then passing multistep
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      passing multistep
      """

  Scenario: should fail multistep
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: run failing multistep
          Given passing step
          When failing multistep
          Then other passing step
      """
    When I run feature suite
    Then the suite should have failed
    And the following step should be failed:
      """
      failing multistep
      """
    And the following steps should be skipped:
      """
      other passing step
      """
    And the following steps should be passed:
      """
      passing step
      """

  Scenario: should fail nested multistep
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: run failing nested multistep
          Given failing nested multistep
          When passing step
      """
    When I run feature suite
    Then the suite should have failed
    And the following step should be failed:
      """
      failing nested multistep
      """
    And the following steps should be skipped:
      """
      passing step
      """

  Scenario: should skip steps after undefined multistep
    Given a feature "undefined.feature" file:
      """
      Feature: run undefined multistep

        Scenario: run undefined multistep
          Given passing step
          When undefined multistep
          Then passing multistep
      """
    When I run feature suite
    Then the suite should have passed
    And the following step should be passed:
      """
      passing step
      """
    And the following step should be undefined:
      """
      undefined multistep
      """
    And the following step should be skipped:
      """
      passing multistep
      """

  Scenario: should match undefined steps in a row
    Given a feature "undefined.feature" file:
      """
      Feature: undefined feature

        Scenario: parse a scenario
          Given undefined step
          When undefined multistep
          Then passing step
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      undefined step
      undefined multistep
      """
    And the following step should be skipped:
      """
      passing step
      """

  Scenario: should mark undefined steps after pending
    Given a feature "pending.feature" file:
      """
      Feature: pending feature

        Scenario: parse a scenario
          Given pending step
          When undefined step
          Then undefined multistep
          And other passing step
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be undefined:
      """
      undefined step
      undefined multistep
      """
    And the following step should be pending:
      """
      pending step
      """
    And the following step should be skipped:
      """
      other passing step
      """


  Scenario: context passed between steps
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

        Scenario: run passing multistep
          Given I return a context from a step
          Then I should see the context in the next step
      """
    When I run feature suite
    Then the suite should have passed

  Scenario: context passed between steps
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

      Scenario: run passing multistep
        Given I can see contexts passed in multisteps
      """
    When I run feature suite
    Then the suite should have passed

  Scenario: should run passing multistep using keyword function successfully
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

        Scenario: run passing multistep
          Given passing step
          Then passing multistep using 'then' function
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      passing multistep using 'then' function
      """

  Scenario: should identify undefined multistep using keyword function
    Given a feature "normal.feature" file:
      """
      Feature: normal feature

        Scenario: run passing multistep
          Given passing step
          Then undefined multistep using 'then' function
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      """
    And the following step should be undefined:
      """
      undefined multistep using 'then' function
      """

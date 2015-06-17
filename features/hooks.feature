Feature: suite hooks
  In order to run tasks before and after important events
  As a test suite
  I need to provide a way to hook into these events

  Background:
    Given I have a before scenario hook
    And a feature path "features/load_features.feature:6"
    And I parse features

  Scenario: hi there
    When I run features
    Then I should have a scenario "" recorded in the hook

  Scenario: and there
    When I run features
    Then I should have a scenario "" recorded in the hook

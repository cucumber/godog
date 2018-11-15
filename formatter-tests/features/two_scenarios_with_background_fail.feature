Feature: two scenarios with background fail

  Background:
    Given passing step
    And failing step

  Scenario: one
    When passing step
    Then passing step

  Scenario: two
    Then passing step

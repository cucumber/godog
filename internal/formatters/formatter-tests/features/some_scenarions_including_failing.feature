Feature: some scenarios

  Scenario: failing
    Given passing step
    When failing step
    Then passing step

  Scenario: pending
    When pending step
    Then passing step

  Scenario: undefined
    When undefined
    Then passing step

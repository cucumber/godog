Feature: Stop on first failure

  Scenario: First scenario - should run and fail
    Given a passing step
    When a failing step
    Then a passing step

  Scenario: Second scenario - should be skipped
    Given a passing step
    Then a passing step 
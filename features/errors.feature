Feature: scenario hook errors
   This feature checks the handling of errors in scenario hooks and steps

  Scenario: no errors
    Given a feature "normal.feature" file:
      """
    Feature: the feature
      Scenario: passing scenario
        When passing step
      """
    When I run feature suite

    Then the suite should have passed
    And the trace should be:
    """
    Feature: the feature
      Scenario: passing scenario
        Step: passing step : passed
    """

  Scenario: hook failures
    Given a feature "normal.feature" file:
      """
    Feature: failures
      @fail_before_scenario
      Scenario: fail before scenario
        When passing step

      @fail_after_scenario
      Scenario: failing after scenario
        And passing step

      @fail_before_scenario
      @fail_after_scenario
      Scenario: failing before and after scenario
        When passing step

      @fail_before_scenario
      Scenario: failing before scenario with failing step
        When failing step

      @fail_after_scenario
      Scenario: failing after scenario with failing step
        And failing step

      @fail_before_scenario
      @fail_after_scenario
      Scenario: failing before and after scenario with failing step
        When failing step
       """
    When I run feature suite

    Then the suite should have failed
    And the trace should be:
    """
    Feature: failures
      Scenario: fail before scenario
        Step: passing step : failed
        Error: before scenario hook failed: failed in before scenario hook
      Scenario: failing after scenario
        Step: passing step : failed
        Error: after scenario hook failed: failed in after scenario hook
      Scenario: failing before and after scenario
        Step: passing step : failed
        Error: after scenario hook failed: failed in after scenario hook, step error: before scenario hook failed: failed in before scenario hook
      Scenario: failing before scenario with failing step
        Step: failing step : failed
        Error: before scenario hook failed: failed in before scenario hook
      Scenario: failing after scenario with failing step
        Step: failing step : failed
        Error: after scenario hook failed: failed in after scenario hook, step error: intentional failure
      Scenario: failing before and after scenario with failing step
        Step: failing step : failed
        Error: after scenario hook failed: failed in after scenario hook, step error: before scenario hook failed: failed in before scenario hook
"""



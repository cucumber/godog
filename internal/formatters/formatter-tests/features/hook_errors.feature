Feature: scenario hook errors

  Scenario: ok scenario
    When passing step

  @fail_before_scenario
  Scenario: failing before scenario
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

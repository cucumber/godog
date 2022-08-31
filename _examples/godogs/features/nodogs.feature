Feature: do not eat godogs
  In order to be fit
  As a well-fed gopher
  I need to be able to avoid godogs

  Scenario: Eat 0 out of 12
    Given there are 12 godogs
    When I eat 0
    Then there should be 12 remaining

  Scenario: Eat 0 out of 0
    Given there are 0 godogs
    When I eat 0
    Then there should be none remaining

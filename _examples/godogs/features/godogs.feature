Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining

  Scenario: Eat 12 out of 12
    Given there are 12 godogs
    When I eat 12
    Then there should be none remaining

# file: $GOPATH/godogs/features/godogs.feature
Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 4
    Then there should be 7 remaining

  Scenario: Eat 12 out of 12
    Given there are 12 godogs
    When I eat 11
    Then there should be none remaining

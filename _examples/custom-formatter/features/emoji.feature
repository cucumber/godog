# file: $GOPATH/godogs/features/godogs.feature
Feature: Custom emoji formatter examples
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Passing test
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining

  Scenario: Failing and Skipped test
    Given there are 12 godogs
    When I eat 5
    Then there should be 6 remaining
     And there should be 4 remaining

  Scenario: Undefined steps
    Given there are 12 godogs
    When I eat 5
    Then this step is not defined

  Scenario: Pending step
    Given there are 12 godogs
    When I eat 5
    Then this step is pending

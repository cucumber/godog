Feature: providing testingT compatibility
  In order to test application behavior using standard go assertion techniques
  As a test suite
  I need to be able to provide a testing.T compatible interface

  Scenario Outline: should fail test with no message if <op> called on testing T
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When my step fails the test by calling <op> on testing T
      """
    When I run feature suite
    Then the suite should have failed
      #YODO WRITE ME ...
    Then testing T should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following step should be failed:
      """
      my step fails the test by calling <op> on testing T
      """
    Examples:
      | op      |
      | Fail    |
      | FailNow |

  Scenario Outline: should fail test with message if <op> called on T
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When my step fails the test by calling <op> on testing T with message "an unformatted message"
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following step should be failed:
      """
      my step fails the test by calling <op> on testing T with message "an unformatted message"
      """
    Examples:
      | op    |
      | Error |
      | Fatal |


  Scenario Outline: should fail test with formatted message if <op> called on T
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When my step fails the test by calling <op> on testing T with message "a formatted message %s" and argument "arg1"
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following step should be failed:
      """
      my step fails the test by calling <op> on testing T with message "a formatted message %s" and argument "arg1"
      """
    Examples:
      | op     |
      | Errorf |
      | Fatalf |


  Scenario: should pass test when testify assertions pass
    Given a feature "testify.feature" file:
      """
      Feature: passed feature

        Scenario: pass a scenario
          Given passing step
          When my step calls testify's assert.Equal with expected "exp" and actual "exp"
          When my step calls testify's require.Equal with expected "exp" and actual "exp"
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      my step calls testify's assert.Equal with expected "exp" and actual "exp"
      my step calls testify's require.Equal with expected "exp" and actual "exp"
      """

  Scenario: should fail test when testify assertions do not pass
    Given a feature "testify.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When my step calls testify's assert.Equal with expected "exp" and actual "not"
          And my step calls testify's assert.Equal with expected "exp2" and actual "not"
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be failed:
      """
      my step calls testify's assert.Equal with expected "exp" and actual "not"
      """
    And the following steps should be skipped:
      """
      my step calls testify's assert.Equal with expected "exp2" and actual "not"
      """


  Scenario: should fail test when multiple testify assertions are used in a step
    Given a feature "testify.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When my step calls testify's assert.Equal 3 times
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be failed:
      """
      my step calls testify's assert.Equal 3 times
      """

  Scenario: should pass test when multiple testify assertions are used successfully in a step
    Given a feature "testify.feature" file:
      """
      Feature: passed feature

        Scenario: pass a scenario
          Given passing step
          When my step calls testify's assert.Equal 3 times with match
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      my step calls testify's assert.Equal 3 times with match
      """

  Scenario Outline: should skip test when <op> is called on the testing.T
    Given a feature "testify.feature" file:
      """
      Feature: skipped feature

        Scenario: skip a scenario
          Given passing step
          When my step skips the test by calling <op> on testing T
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be skipped:
      """
      my step skips the test by calling <op> on testing T
      """
    Examples:
      | op      |
      | Skip    |
      | SkipNow |

  Scenario: should log when Logf/Log called on testing.T
    When my step calls Logf on testing T with message "format this %s" and argument "formatparam1"
    And my step calls Log on testing T with message "log this message"
    Then the logged messages should include "format this formatparam1"
    And the logged messages should include "log this message"

  Scenario: should log when godog.Logf/Log called
    When my step calls godog.Logf with message "format this %s" and argument "formatparam1"
    And my step calls godog.Log with message "log this message"
    Then the logged messages should include "format this formatparam1"
    And the logged messages should include "log this message"

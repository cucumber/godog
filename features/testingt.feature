Feature: providing testingT compatibility
  In order to test application behavior using standard go assertion techniques
  As a test suite
  I need to be able to provide a testing.T compatible interface

  Scenario: should fail test if FailNow called on testing T
    Given a feature "failed.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When I fail the test by calling FailNow on testing T
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following step should be failed:
      """
      I fail the test by calling FailNow on testing T
      """

  Scenario: should pass test when testify assertions pass
    Given a feature "testify.feature" file:
      """
      Feature: passed feature

        Scenario: pass a scenario
          Given passing step
          When I call testify's assert.Equal with expected "exp" and actual "exp"
          When I call testify's require.Equal with expected "exp" and actual "exp"
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      I call testify's assert.Equal with expected "exp" and actual "exp"
      I call testify's require.Equal with expected "exp" and actual "exp"
      """

  Scenario: should fail test when testify assertions do not pass
    Given a feature "testify.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When I call testify's assert.Equal with expected "exp" and actual "not"
          And I call testify's assert.Equal with expected "exp2" and actual "not"
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be failed:
      """
      I call testify's assert.Equal with expected "exp" and actual "not"
      """
    And the following steps should be skipped:
      """
      I call testify's assert.Equal with expected "exp2" and actual "not"
      """

  Scenario: should fail test when multiple testify assertions are used in a step
    Given a feature "testify.feature" file:
      """
      Feature: failed feature

        Scenario: fail a scenario
          Given passing step
          When I call testify's assert.Equal 3 times
      """
    When I run feature suite
    Then the suite should have failed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be failed:
      """
      I call testify's assert.Equal 3 times
      """

  Scenario: should pass test when multiple testify assertions are used successfully in a step
    Given a feature "testify.feature" file:
      """
      Feature: passed feature

        Scenario: pass a scenario
          Given passing step
          When I call testify's assert.Equal 3 times with match
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      I call testify's assert.Equal 3 times with match
      """

  Scenario: should skip test when skip is called on the testing.T
    Given a feature "testify.feature" file:
      """
      Feature: skipped feature

        Scenario: skip a scenario
          Given passing step
          When I skip the test by calling Skip on testing T
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      """
    And the following steps should be skipped:
      """
      I skip the test by calling Skip on testing T
      """

  Scenario: should log to testing.T
    Given a feature "logging.feature" file:
      """
      Feature: logged feature

        Scenario: logged scenario
          Given passing step
          When I call Logf on testing T with message "format this %s" and argument "formatparam1"
          And I call Log on testing T with message "log this message"
      """
    When I run feature suite
    Then the suite should have passed
    And the following steps should be passed:
      """
      passing step
      I call Logf on testing T with message "format this %s" and argument "formatparam1"
      I call Log on testing T with message "log this message"
      """

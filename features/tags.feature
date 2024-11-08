Feature: tag filters
  In order to test application behavior
  As a test suite
  I need to be able to filter features and scenarios by tags

  Scenario: should filter outline examples by tags
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Scenario Outline: only tagged examples should run
          Given <status> step

          Examples:
            | status |
            | passing |

          @run_these_examples_only
          Examples:
            | status |
            | other passing |
      """
    When I run feature suite with tags "@run_these_examples_only"
    Then the suite should have passed
    And only the following steps should have run and should be passed:
      """
      other passing step
      """

  Scenario: should filter scenarios by single tag
    Given a feature "normal.feature" file:
      """
      Feature: outline

        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given passing step

        @some_other_tag
        Scenario Outline: only tagged examples should run
          Given second passing step

        @some_other_tag
        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given third passing step

      """
    When I run feature suite with tags "@run_these_examples_only"
    Then the suite should have passed
    And only the following steps should have run and should be passed:
      """
      passing step
      third passing step
      """

  Scenario: should filter scenarios by And-Not tag expression
    Given a feature "normal.feature" file:
      """
      Feature: outline

        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given passing step

        @some_other_tag
        Scenario Outline: only tagged examples should run
          Given second passing step

        @some_other_tag
        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given third passing step

      """
    When I run feature suite with tags "@run_these_examples_only && ~@some_other_tag"
    Then the suite should have passed
    And only the following steps should have run and should be passed:
      """
      passing step
      """

  Scenario: should filter scenarios by And tag expression
    Given a feature "normal.feature" file:
      """
      Feature: outline

        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given passing step

        @some_other_tag
        Scenario Outline: only tagged examples should run
          Given second passing step

        @some_other_tag
        @run_these_examples_only
        Scenario Outline: only tagged examples should run
          Given third passing step

      """
    When I run feature suite with tags "@run_these_examples_only && @some_other_tag"
    Then the suite should have passed
    And only the following steps should have run and should be passed:
      """
      third passing step
      """

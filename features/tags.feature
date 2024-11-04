Feature: tag filters
  In order to test application behavior
  As a test suite
  I need to be able to filter features and scenarios by tags

  Scenario: should filter outline examples by tags
    Given a feature "normal.feature" file:
      """
      Feature: outline

        Background:
          Given passing step
          And passing step without return

        Scenario Outline: parse a scenario
          Given a feature path "<path>"
          When I parse features
          Then I should have <num> scenario registered

          Examples:
            | path                    | num |
            | features/load.feature:3 | 0   |

          @used
          Examples:
            | path                    | num |
            | features/load.feature:6 | 1   |
      """
    When I run feature suite with tags "@used"
    Then the suite should have passed
    And the following steps should be passed:
      """
      I parse features
      I should have 1 scenario registered
      a feature path "features/load.feature:6"
      passing step
      passing step without return
      """
    And I should have 1 scenario registered

  Scenario: should filter scenarios by X tag
    Given a feature "normal.feature" file:
      """
      Feature: tagged

        @x
        Scenario: one
          Given a feature path "one"

        @x
        Scenario: two
          Given a feature path "two"

        @x @y
        Scenario: three
          Given a feature path "three"

        @y
        Scenario: four
          Given a feature path "four"
      """
    When I run feature suite with tags "@x"
    Then the suite should have passed
    And I should have 3 scenario registered
    And the following steps should be passed:
      """
      a feature path "one"
      a feature path "two"
      a feature path "three"
      """

  Scenario: should filter scenarios by X tag not having Y
    Given a feature "normal.feature" file:
      """
      Feature: tagged

        @x
        Scenario: one
          Given a feature path "one"

        @x
        Scenario: two
          Given a feature path "two"

        @x @y
        Scenario: three
          Given a feature path "three"

        @y @z
        Scenario: four
          Given a feature path "four"
      """
    When I run feature suite with tags "@x && ~@y"
    Then the suite should have passed
    And I should have 2 scenario registered
    And the following steps should be passed:
      """
      a feature path "one"
      a feature path "two"
      """

  Scenario: should filter scenarios having Y and Z tags
    Given a feature "normal.feature" file:
      """
      Feature: tagged

        @x
        Scenario: one
          Given a feature path "one"

        @x
        Scenario: two
          Given a feature path "two"

        @x @y
        Scenario: three
          Given a feature path "three"

        @y @z
        Scenario: four
          Given a feature path "four"
      """
    When I run feature suite with tags "@y && @z"
    Then the suite should have passed
    And I should have 1 scenario registered
    And the following steps should be passed:
      """
      a feature path "four"
      """

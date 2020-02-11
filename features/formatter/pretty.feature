Feature: pretty formatter
  In order to support tools that import pretty output
  I need to be able to support pretty formatted output

  Scenario: Support of Feature Plus Scenario Node
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:3

      1 scenarios (1 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Scenario Node With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description
        @TAG2 @TAG3
        Scenario: simple scenario
            simple scenario description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:5

      1 scenarios (1 undefined)
      No steps
      0s
    """
  Scenario: Support of Feature Plus Scenario Outline
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario Outline: simple scenario
            simple scenario description

        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:4

          Examples: simple examples
            | status |
            | pass   |
            | fail   |

      2 scenarios (2 undefined)
      No steps
      0s
    """

  Scenario: Support of Feature Plus Scenario Outline With Tags
    Given a feature "features/simple.feature" file:
    """
        @TAG1
        Feature: simple feature
            simple feature description

        @TAG2
        Scenario Outline: simple scenario
            simple scenario description

        @TAG3
        Examples: simple examples
        | status |
        | pass   |
        | fail   |
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:6

          Examples: simple examples
            | status |
            | pass   |
            | fail   |

      2 scenarios (2 undefined)
      No steps
      0s
    """
  Scenario: Support of Feature Plus Scenario With Steps
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
            simple feature description

        Scenario: simple scenario
            simple scenario description

        Given passing step
        Then a failing step

    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:4
          Given passing step      # suite_context.go:0 -> SuiteContext.func2
          Then a failing step     # suite_context.go:0 -> *suiteContext
          intentional failure

      --- Failed steps:

        Scenario: simple scenario # features/simple.feature:4
          Then a failing step # features/simple.feature:8
            Error: intentional failure


      1 scenarios (1 failed)
      2 steps (1 passed, 1 failed)
      0s
    """
  Scenario: Support of Feature Plus Scenario Outline With Steps
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario
        simple scenario description

          Given <status> step

        Examples: simple examples
        | status |
        | passing |
        | failing |

    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario Outline: simple scenario # features/simple.feature:4
          Given <status> step             # suite_context.go:0 -> SuiteContext.func2

          Examples: simple examples
            | status  |
            | passing |
            | failing |
              intentional failure

      --- Failed steps:

        Scenario Outline: simple scenario # features/simple.feature:4
          Given failing step # features/simple.feature:7
            Error: intentional failure


      2 scenarios (1 passed, 1 failed)
      2 steps (1 passed, 1 failed)
      0s
    """

  # Currently godog only supports comments on Feature and not
  # scenario and steps.
  Scenario: Support of Comments
    Given a feature "features/simple.feature" file:
    """
        #Feature comment
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple description

        Scenario: simple scenario # features/simple.feature:5

      1 scenarios (1 undefined)
      No steps
      0s
    """
  Scenario: Support of Docstrings
    Given a feature "features/simple.feature" file:
    """
        Feature: simple feature
          simple description

          Scenario: simple scenario
          simple feature description

          Given passing step
          \"\"\" content type
          step doc string
          \"\"\"
    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple description

        Scenario: simple scenario # features/simple.feature:4
          Given passing step      # suite_context.go:0 -> SuiteContext.func2
            \"\"\"  content type
            step doc string
            \"\"\"

      1 scenarios (1 passed)
      1 steps (1 passed)
      0s
    """
  Scenario: Support of Undefined, Pending and Skipped status
    Given a feature "features/simple.feature" file:
    """
      Feature: simple feature
      simple feature description

      Scenario: simple scenario
      simple scenario description

        Given passing step
        And pending step
        And undefined
        And passing step

    """
    When I run feature suite with formatter "pretty"
    Then the rendered output will be as follows:
    """
      Feature: simple feature
        simple feature description

        Scenario: simple scenario # features/simple.feature:4
          Given passing step      # suite_context.go:0 -> SuiteContext.func2
          And pending step        # suite_context.go:0 -> SuiteContext.func1
            TODO: write pending definition
          And undefined
          And passing step        # suite_context.go:0 -> SuiteContext.func2

      1 scenarios (1 pending, 1 undefined)
      4 steps (1 passed, 1 pending, 1 undefined, 1 skipped)
      0s

      You can implement step definitions for undefined steps with these snippets:

      func undefined() error {
        return godog.ErrPending
      }

      func FeatureContext(s *godog.Suite) {
        s.Step(`^undefined$`, undefined)
      }
    """
